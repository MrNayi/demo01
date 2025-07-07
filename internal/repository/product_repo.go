package repository

import (
	"context"
	"demo01/internal/model"
	"errors"

	"gorm.io/gorm"
)

var (
	ErrInsufficientStock = errors.New("库存不足")
)

type ProductRepo struct {
	db *gorm.DB
}

func NewProductRepo(db *gorm.DB) *ProductRepo {
	return &ProductRepo{db: db}
}

// Create 创建商品
func (r *ProductRepo) Create(ctx context.Context, product *model.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

// GetByID 根据ID获取商品
func (r *ProductRepo) GetByID(ctx context.Context, id int) (*model.Product, error) {
	var product model.Product
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

// GetAll 获取所有商品（分页）
func (r *ProductRepo) GetAll(ctx context.Context, page, pageSize int) ([]model.Product, error) {
	var products []model.Product
	offset := (page - 1) * pageSize
	err := r.db.WithContext(ctx).Offset(offset).Limit(pageSize).Find(&products).Error
	return products, err
}

// UpdateStock 更新库存（乐观锁）
func (r *ProductRepo) UpdateStock(ctx context.Context, productID, quantity, version int) error {
	result := r.db.WithContext(ctx).Model(&model.Product{}).
		Where("id = ? AND version = ?", productID, version).
		Updates(map[string]interface{}{
			"stock":   gorm.Expr("stock - ?", quantity),
			"version": gorm.Expr("version + 1"),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// GetStock 获取库存
func (r *ProductRepo) GetStock(ctx context.Context, productID int) (int, error) {
	var product model.Product
	err := r.db.WithContext(ctx).Select("stock").Where("id = ?", productID).First(&product).Error
	if err != nil {
		return 0, err
	}
	return product.Stock, nil
}
