package service

import (
	"context"
	"demo01/internal/model"
	"demo01/internal/repository"
	"encoding/json"
	"sync"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type OrderService struct {
	orderRepo     *repository.OrderRepo
	inventoryRepo *repository.InventoryRepo
	kafkaProducer *kafka.Producer
	rwLock        sync.RWMutex            // 读写锁（读多写少场景）
	localCache    map[string]*model.Order // 本地缓存（体现map底层哈希表）
}

func NewOrderService(orderRepo *repository.OrderRepo, inventoryRepo *repository.InventoryRepo, producer *kafka.Producer) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		inventoryRepo: inventoryRepo,
		kafkaProducer: producer,
		localCache:    make(map[string]*model.Order), // 初始化map
	}
}

// CreateOrder 创建订单
func (s *OrderService) CreateOrder(ctx context.Context, userID string, items []model.OrderItem) (*model.Order, error) {
	// 1. 生成订单ID（string拼接）
	orderID := "order_" + time.Now().Format("20060102150405") + "_" + userID

	// 2. 并发扣减库存（Goroutine+WaitGroup）
	var wg sync.WaitGroup
	errChan := make(chan error, len(items)) // channel用于传递错误

	for _, item := range items {
		wg.Add(1)
		go func(productID int, quantity int) { // Goroutine并发处理
			defer wg.Done()
			// 扣减库存时加写锁
			s.rwLock.Lock()
			defer s.rwLock.Unlock()

			// 乐观锁扣减库存
			if err := s.inventoryRepo.DecreaseStock(ctx, productID, quantity); err != nil {
				errChan <- err
			}
		}(item.ProductID, item.Quantity)
	}

	// 等待所有扣减完成（带超时的Context）
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	go func() {
		wg.Wait()
		close(errChan)
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err() // Context超时控制
	case err := <-errChan:
		if err != nil {
			return nil, err
		}
	}

	// 3. 创建订单
	itemsJSON, _ := json.Marshal(items) // slice转JSON存储
	order := &model.Order{
		ID:          orderID,
		UserID:      userID,
		Items:       string(itemsJSON),
		TotalAmount: calculateTotal(items),
		CreatedAt:   time.Now(),
	}
	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	// 4. 异步发送Kafka消息（体现GMP调度中G的异步执行）
	if s.kafkaProducer != nil {
		go func() {
			topic := "order_created"
			msg := &kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
				Value:          []byte(orderID),
			}
			err := s.kafkaProducer.Produce(msg, nil)
			if err != nil {
				// 日志记录
			}
		}()
	}

	// 5. 写入本地缓存（map操作）
	s.rwLock.Lock()
	s.localCache[orderID] = order
	s.rwLock.Unlock()

	return order, nil
}

// 查询所有订单
// 这里我觉得应该用分页
func (s *OrderService) GetAllOrders(ctx context.Context, page, pageSize int) ([]model.Order, error) {
	orders, err := s.orderRepo.GetAll(ctx, page, pageSize)
	if err != nil {
		return nil, err
	}
	return orders, nil
}

// GetOrder 根据id查询订单
func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*model.Order, error) {
	// 1. 先从本地缓存查询
	s.rwLock.RLock()
	if order, exists := s.localCache[orderID]; exists {
		s.rwLock.RUnlock()
		return order, nil
	}
	s.rwLock.RUnlock()

	// 2. 从数据库查询
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// 3. 写入本地缓存
	s.rwLock.Lock()
	s.localCache[orderID] = order
	s.rwLock.Unlock()

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
