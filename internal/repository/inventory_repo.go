package repository

import (
	"context"
	"demo01/internal/model"
	"time"

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
// DecreaseStock 扣减库存（乐观锁 + 自旋重试）
// 1.上下文  2.商品id 3.库存 -- 参数
func (r *InventoryRepo) DecreaseStock(ctx context.Context, productID int, quantity int) error {
	// 乐观锁参数
	const (
		maxRetries = 8                     // 最大重试次数（增加重试次数）
		baseDelay  = 5 * time.Millisecond  // 基础重试间隔
		maxDelay   = 50 * time.Millisecond // 最大重试间隔
	)

	for attempt := 0; attempt < maxRetries; attempt++ {
		err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
				return gorm.ErrRecordNotFound // 版本错误 说明被修改
			}

			return result.Error
		})

		// 如果成功或非版本冲突错误，直接返回
		if err == nil || err != gorm.ErrRecordNotFound {
			return err
		}

		// 版本冲突 需要重试
		if attempt < maxRetries-1 {
			// 指数退避：延迟时间随重试次数增加
			delay := baseDelay * time.Duration(1<<attempt)
			if delay > maxDelay {
				delay = maxDelay
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
				// 继续重试
			}
		}
	}

	return gorm.ErrRecordNotFound // 达到最大重试次数
}
