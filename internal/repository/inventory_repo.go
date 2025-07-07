package repository

import (
	"context"
	"demo01/internal/model"

	"gorm.io/gorm"
)

// 库存相关操作
type InventoryRepo struct {
	db *gorm.DB
}

// 工厂模式创建一个仓库实例
func NewInventoryRepo(db *gorm.DB) *InventoryRepo {
	return &InventoryRepo{db: db}
}

// 核心操作 执行商品扣减
// DecreaseStock 扣减库存（乐观锁）
// 1.上下文  2.商品id 3.库存
func (r *InventoryRepo) DecreaseStock(ctx context.Context, productID int, quantity int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var inventory model.Inventory
		if err := tx.Where("product_id = ?", productID).First(&inventory).Error; err != nil {
			return err
		}

		if inventory.Stock < quantity {
			return gorm.ErrRecordNotFound // 库存不足
		}

		// 乐观锁更新
		result := tx.Model(&model.Inventory{}).
			Where("product_id = ? AND version = ?", productID, inventory.Version).
			Updates(map[string]interface{}{
				"stock":   inventory.Stock - quantity,
				"version": inventory.Version + 1,
			})

		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound // 版本冲突
		}

		return result.Error
	})
}
