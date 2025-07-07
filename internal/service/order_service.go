package service

import (
	"context"
	"demo01/internal/model"
	"demo01/internal/repository"
	"demo01/internal/util"
	"encoding/json"
	"sync"
	"time"
)

type OrderService struct {
	orderRepo     *repository.OrderRepo
	inventoryRepo *repository.InventoryRepo
	rwLock        sync.RWMutex // 读写锁（读多写少场景）
	localCache    sync.Map     // 本地缓存 这里Sync.Map解决了并发访问缓存的问题 并且适合读多写少的场景
}

func NewOrderService(orderRepo *repository.OrderRepo, inventoryRepo *repository.InventoryRepo) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		inventoryRepo: inventoryRepo,
		// sync.Map 不需要初始化，零值即可使用
	}
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(ctx context.Context, userID string, items []model.OrderItem) (*model.Order, error) {
	// 记录开始时间，用于性能监控
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

	// 2. 生成订单ID（string拼接）
	orderID := "order_" + time.Now().Format("20060102150405") + "_" + userID

	util.GlobalLogger.Debug(ctx, "生成订单ID",
		util.Field{Key: "order_id", Value: orderID},
	)

	// 3. 扣减库存（使用事务保证一致性）
	err := s.decreaseInventoryWithTransaction(ctx, items)
	if err != nil {
		util.GlobalLogger.Error(ctx, "库存扣减失败", err,
			util.Field{Key: "order_id", Value: orderID},
		)
		return nil, util.NewBusinessError("INSUFFICIENT_STOCK", "库存扣减失败", err)
	}

	// 4. 创建订单
	itemsJSON, _ := json.Marshal(items) // slice转JSON存储
	order := &model.Order{
		ID:          orderID,
		UserID:      userID,
		Items:       string(itemsJSON),
		TotalAmount: calculateTotal(items),
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		util.GlobalLogger.Error(ctx, "订单创建失败", err,
			util.Field{Key: "order_id", Value: orderID},
		)
		return nil, util.NewBusinessError("ORDER_CREATE_FAILED", "订单创建失败", err)
	}

	// 5. 写入本地缓存
	s.localCache.Store(orderID, order)

	// 记录完成时间
	duration := time.Since(startTime)
	util.GlobalLogger.Info(ctx, "订单创建成功",
		util.Field{Key: "order_id", Value: orderID},
		util.Field{Key: "duration_ms", Value: duration.Milliseconds()},
		util.Field{Key: "total_amount", Value: order.TotalAmount},
	)

	return order, nil
}

// decreaseInventoryWithTransaction 事务性扣减库存
func (s *OrderService) decreaseInventoryWithTransaction(ctx context.Context, items []model.OrderItem) error {
	// 使用读写锁保护库存操作
	s.rwLock.Lock()
	defer s.rwLock.Unlock()

	for _, item := range items {
		util.GlobalLogger.Debug(ctx, "扣减库存",
			util.Field{Key: "product_id", Value: item.ProductID},
			util.Field{Key: "quantity", Value: item.Quantity},
		)

		// 乐观锁扣减库存
		if err := s.inventoryRepo.DecreaseStock(ctx, item.ProductID, item.Quantity); err != nil {
			util.GlobalLogger.Error(ctx, "库存扣减失败", err,
				util.Field{Key: "product_id", Value: item.ProductID},
				util.Field{Key: "quantity", Value: item.Quantity},
			)
			return err
		}
	}

	return nil
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

// GetOrder 根据id查询订单
func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*model.Order, error) {
	util.GlobalLogger.Debug(ctx, "查询订单详情",
		util.Field{Key: "order_id", Value: orderID},
	)

	// 1. 先从本地缓存查询
	if orderValue, exists := s.localCache.Load(orderID); exists {
		if order, ok := orderValue.(*model.Order); ok {
			util.GlobalLogger.Debug(ctx, "从缓存获取订单",
				util.Field{Key: "order_id", Value: orderID},
			)
			return order, nil
		}
	}

	// 2. 从数据库查询
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		util.GlobalLogger.Error(ctx, "查询订单失败", err,
			util.Field{Key: "order_id", Value: orderID},
		)
		return nil, util.NewBusinessError("ORDER_NOT_FOUND", "订单不存在", err)
	}

	// 3. 写入本地缓存
	s.localCache.Store(orderID, order)

	util.GlobalLogger.Debug(ctx, "从数据库获取订单并缓存",
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
