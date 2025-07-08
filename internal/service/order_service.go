package service

import (
	"context"
	"demo01/internal/model"
	"demo01/internal/repository"
	"demo01/internal/util"
	"encoding/json"
	"sync"
	"time"

	"gorm.io/gorm"
)

// service层更加关注业务逻辑执行 调用repo完成和数据库的交互
// 在这里最好不要去直接调用gorm操作数据库

type OrderService struct {
	// 订单服务 需要用到订单repo和库存的repo 去进行数据库的交互
	orderRepo     *repository.OrderRepo
	inventoryRepo *repository.InventoryRepo
	localCache    sync.Map // 本地缓存 使用Sync.Map本地缓存 加速订单查询
}

// NewOrderService 创建订单服务实例
func NewOrderService(orderRepo *repository.OrderRepo, inventoryRepo *repository.InventoryRepo) *OrderService {
	return &OrderService{
		// 需要创建订单和扣减库存
		orderRepo:     orderRepo,
		inventoryRepo: inventoryRepo,
	}
}

// CreateOrder 创建订单（带补偿机制）
func (s *OrderService) CreateOrder(ctx context.Context, userID string, items []model.OrderItem) (*model.Order, error) {
	// 性能监控
	startTime := time.Now()

	util.GlobalLogger.Info(ctx, "开始创建订单",
		util.Field{Key: "user_id", Value: userID},
		util.Field{Key: "item_count", Value: len(items)},
	)

	// 1. 参数验证
	if userID == "" || len(items) == 0 {
		util.GlobalLogger.Error(ctx, "订单参数无效", util.ErrInvalidInput)
		return nil, util.NewBusinessError("INVALID_PARAMS", "订单参数无效", util.ErrInvalidInput)
	}

	// 2. 生成订单ID（缩短格式，避免数据库字段长度限制）
	orderID := "o" + time.Now().Format("0102150405") + userID

	util.GlobalLogger.Debug(ctx, "生成订单ID",
		util.Field{Key: "order_id", Value: orderID},
	)

	// 3. 使用事务保证原子性：库存扣减 + 订单创建
	// 这里将他们放在一起是因为 一开始库存扣减使用事务后 再进行订单创建
	// 但是我认为这样如果订单创建失败且不具备补偿机制的同时 会发生一些"死库存" 就是库存扣减了 但是订单没有创建成功
	// 后续还可能发生未支付等情况 我认为可以使用消息队列的延迟对列 这样扣减库存就不需要和订单绑定 通过mq发送消息
	// 假设订单创建失败还可以进行重试 即使被取消支付或者订单创建失败 也可以通过补偿机制 将库存回补！
	// 但是 事务也只能够保证库存扣减和订单创建的原子性 但是无法保证超卖问题 因为在判断库存扣减的过程中可能会发生库存判断失误
	var order *model.Order
	err := s.orderRepo.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 通过 Repository 层的事务方法扣减库存（使用分布式锁防止超卖）
		for _, item := range items {
			util.GlobalLogger.Debug(ctx, "扣减库存",
				util.Field{Key: "product_id", Value: item.ProductID},
				util.Field{Key: "quantity", Value: item.Quantity},
			)

			// 使用分布式锁保护库存扣减，防止超卖
			if err := s.inventoryRepo.DecreaseStockWithDistributedLock(ctx, item.ProductID, item.Quantity); err != nil {
				util.GlobalLogger.Error(ctx, "库存扣减失败", err,
					util.Field{Key: "product_id", Value: item.ProductID},
					util.Field{Key: "quantity", Value: item.Quantity},
				)
				// 库存扣减失败的原因可能不止库存不足 还可能存在其他问题 比如库存查询失败 库存更新失败 等
				// 所以这里需要返回一个错误 让调用者进行处理 但是这里先返回库存不足
				// TODO error需要进一步精确
				return util.NewBusinessError("INSUFFICIENT_STOCK", "库存不足", err)
			}
		}

		// 分割items 因为这里传递过来的是一个商品列表 需要将商品列表转换为json字符串
		itemsJSON, err := json.Marshal(items)
		if err != nil {
			util.GlobalLogger.Error(ctx, "商品信息序列化失败", err)
			return util.NewBusinessError("JSON_MARSHAL_FAILED", "商品信息序列化失败", err)
		}
		// 组装订单model
		order = &model.Order{
			// 一一对应 创建一个订单的对象 准备通过repo写入数据库
			ID:          orderID,
			UserID:      userID,
			Items:       string(itemsJSON),     // 这里是一个商品列表 但是我在想 一个订单下单多个商品，查看单个商品订单详情怎么处理呢
			TotalAmount: calculateTotal(items), // 计算商品总价值 即订单的总金额
			Status:      "pending",             // 订单状态 待支付 已支付 已取消 已完成 这里可以定义枚举类型 如ORDER_STATUS_PENDING
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// 通过 Repository 层的事务方法创建订单
		// service只关注业务逻辑处理 不应该直接去通过tx操作数据库
		if err := s.orderRepo.CreateWithTx(tx, order); err != nil {
			// error处理
			util.GlobalLogger.Error(ctx, "订单创建失败", err,
				util.Field{Key: "order_id", Value: orderID},
			)
			return util.NewBusinessError("ORDER_CREATE_FAILED", "订单创建失败", err)
		}

		return nil
	})

	// 4. 事务失败处理
	if err != nil {
		util.GlobalLogger.Error(ctx, "订单创建事务失败", err,
			util.Field{Key: "order_id", Value: orderID},
		)
		return nil, err
	}

	// 5. 写入多级缓存
	s.localCache.Store(orderID, order)
	// 通过 Repository 层写入 Redis 缓存
	if err := s.orderRepo.SetToCache(ctx, orderID, order); err != nil {
		util.GlobalLogger.Warn(ctx, "Redis缓存写入失败",
			util.Field{Key: "order_id", Value: orderID},
			util.Field{Key: "error", Value: err.Error()},
		)
	}

	// 6. 记录完成时间
	duration := time.Since(startTime)
	util.GlobalLogger.Info(ctx, "订单创建成功",
		util.Field{Key: "order_id", Value: orderID},
		util.Field{Key: "duration_ms", Value: duration.Milliseconds()},
		util.Field{Key: "total_amount", Value: order.TotalAmount},
	)

	return order, nil
}

// GetAllOrders 查询所有订单（分页）
func (s *OrderService) GetAllOrders(ctx context.Context, page, pageSize int) ([]model.Order, error) {
	util.GlobalLogger.Debug(ctx, "查询订单列表",
		util.Field{Key: "page", Value: page},
		util.Field{Key: "page_size", Value: pageSize},
	)

	orders, err := s.orderRepo.GetAll(ctx, page, pageSize)
	if err != nil {
		util.GlobalLogger.Error(ctx, "查询订单列表失败", err)
		return nil, util.NewBusinessError("QUERY_FAILED", "查询订单列表失败", err)
	}

	util.GlobalLogger.Debug(ctx, "查询订单列表成功",
		util.Field{Key: "count", Value: len(orders)},
	)

	return orders, nil
}

// GetOrder 根据id查询订单（多级缓存）
func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*model.Order, error) {
	util.GlobalLogger.Debug(ctx, "查询订单详情",
		util.Field{Key: "order_id", Value: orderID},
	)

	// 1. 一级缓存：本地缓存查询（最快）
	// 这里有点为了用Sync.Map的意思 但是我认为不能使用普通的map 用户可能会多端同时操作订单 可能频繁刷新修改订单信息 可能使用脚本或者工具...
	// 使用map还是会有订单数据不一致的问题
	// 当然sync.map会增加内存开销和读取耗时 但我觉得是可以接受的 首先他是本地缓存 本地缓存的数据量并没有那么大而且会有过期机制
	// 其次 一般很少出现同时对订单不断刷新和修改的情况 所以并发程度不是很高 读取性能上不会那么耗时
	// 如果追求用户体验的话 多级缓存也是合理的 响应非常快 只是不能跨会话 但是这一点redis能够做到补偿 并且能够反写回本地缓存
	// 即使用户短时间多次刷新页面 也会有一个非常极速的响应
	if orderValue, exists := s.localCache.Load(orderID); exists {
		if order, ok := orderValue.(*model.Order); ok {
			util.GlobalLogger.Debug(ctx, "从本地缓存获取订单",
				util.Field{Key: "order_id", Value: orderID},
			)
			return order, nil
		}
	}

	// 2. 二级缓存：Redis缓存查询（较快）
	if order, err := s.orderRepo.GetFromCache(ctx, orderID); err == nil {
		// 反序列化成功，存入本地缓存
		s.localCache.Store(orderID, order)
		util.GlobalLogger.Debug(ctx, "从Redis缓存获取订单",
			util.Field{Key: "order_id", Value: orderID},
		)
		return order, nil
	}

	// 3. 三级存储：数据库查询（最慢）
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		util.GlobalLogger.Error(ctx, "查询订单失败", err,
			util.Field{Key: "order_id", Value: orderID},
		)
		return nil, util.NewBusinessError("ORDER_NOT_FOUND", "订单不存在", err)
	}

	// 4. 回填多级缓存
	s.localCache.Store(orderID, order)
	// 通过 Repository 层写入 Redis 缓存
	if err := s.orderRepo.SetToCache(ctx, orderID, order); err != nil {
		util.GlobalLogger.Warn(ctx, "Redis缓存写入失败",
			util.Field{Key: "order_id", Value: orderID},
			util.Field{Key: "error", Value: err.Error()},
		)
	}

	util.GlobalLogger.Debug(ctx, "从数据库获取订单并回填多级缓存",
		util.Field{Key: "order_id", Value: orderID},
	)

	return order, nil
}

// calculateTotal 计算订单总金额
func calculateTotal(items []model.OrderItem) float64 {
	total := 0.0
	for _, item := range items {
		total += item.Price * float64(item.Quantity)
	}
	return total
}
