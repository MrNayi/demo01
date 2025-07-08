package repository

import (
	"context"
	"demo01/internal/model"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// 订单相关操作

type OrderRepo struct {
	db          *gorm.DB
	redisClient *redis.Client
}

func NewOrderRepo(db *gorm.DB, redisClient *redis.Client) *OrderRepo {
	return &OrderRepo{
		db:          db,
		redisClient: redisClient,
	}
}

// Create 新增订单（GORM用法）
func (r *OrderRepo) Create(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

// CreateWithTx 在事务中创建订单
func (r *OrderRepo) CreateWithTx(tx *gorm.DB, order *model.Order) error {
	return tx.Create(order).Error
}

// GetAll 分页查询所有订单
func (r *OrderRepo) GetAll(ctx context.Context, page, pageSize int) ([]model.Order, error) {
	// 参数验证
	if page <= 0 || pageSize <= 0 {
		return nil, gorm.ErrInvalidData
	}

	var orders []model.Order
	err := r.db.WithContext(ctx).Offset((page - 1) * pageSize).Limit(pageSize).Find(&orders).Error
	return orders, err
}

// GetByID 根据ID查询订单
func (r *OrderRepo) GetByID(ctx context.Context, id string) (*model.Order, error) {
	// 参数验证
	if id == "" {
		return nil, gorm.ErrInvalidData
	}

	var order model.Order
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// GetFromCache 从Redis缓存获取订单
func (r *OrderRepo) GetFromCache(ctx context.Context, orderID string) (*model.Order, error) {
	if r.redisClient == nil {
		return nil, redis.Nil
	}

	orderJSON, err := r.redisClient.Get(ctx, "order:"+orderID).Result()
	if err != nil {
		return nil, err
	}

	var order model.Order
	if err := json.Unmarshal([]byte(orderJSON), &order); err != nil {
		return nil, err
	}

	return &order, nil
}

// SetToCache 将订单写入Redis缓存
func (r *OrderRepo) SetToCache(ctx context.Context, orderID string, order *model.Order) error {
	if r.redisClient == nil {
		return nil
	}

	orderJSON, err := json.Marshal(order)
	if err != nil {
		return err
	}

	return r.redisClient.Set(ctx, "order:"+orderID, orderJSON, time.Hour).Err()
}

// GetDB 获取数据库连接（用于事务）
func (r *OrderRepo) GetDB() *gorm.DB {
	return r.db
}
