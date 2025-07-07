package repository

import (
	"context"
	"demo01/internal/model"

	"gorm.io/gorm"
)

// 订单相关操作

type OrderRepo struct {
	db *gorm.DB
}

func NewOrderRepo(db *gorm.DB) *OrderRepo {
	return &OrderRepo{db: db}
}

// Create 新增订单（GORM用法）
func (r *OrderRepo) Create(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

// 分页查询所有订单
func (r *OrderRepo) GetAll(ctx context.Context, page, pageSize int) ([]model.Order, error) {
	var orders []model.Order
	err := r.db.WithContext(ctx).Offset((page - 1) * pageSize).Limit(pageSize).Find(&orders).Error
	return orders, err
}

// GetByID 查询订单
func (r *OrderRepo) GetByID(ctx context.Context, id string) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&order).Error
	return &order, err
}
