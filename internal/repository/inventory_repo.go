package repository

import (
	"context"
	"demo01/internal/model"
	"demo01/internal/util"
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

// DecreaseStockWithDistributedLock 使用分布式锁扣减库存（优化版）
func (r *InventoryRepo) DecreaseStockWithDistributedLock(ctx context.Context, productID int, quantity int) error {
	// 1. 先快速检查库存是否充足（不加锁）
	var inventory model.Inventory
	if err := r.db.WithContext(ctx).Where("product_id = ?", productID).First(&inventory).Error; err != nil {
		return err
	}

	// 如果库存明显不足，直接返回错误，避免不必要的锁竞争
	if inventory.Stock < quantity {
		return util.NewBusinessError("INSUFFICIENT_STOCK", "库存不足", gorm.ErrRecordNotFound)
	}

	// 2. 使用分布式锁保护扣减操作
	keyGenerator := util.NewLockKeyGenerator()
	lockKey := keyGenerator.GenerateInventoryLockKey(productID)
	lock := util.NewDistributedLock(util.RedisClient, lockKey, 10*time.Second)

	// 使用WithLock方法，自动处理锁的获取和释放
	return lock.WithLock(ctx, func() error {
		// 在锁保护下再次检查库存并扣减
		return r.DecreaseStockWithTx(r.db, productID, quantity)
	})
}

// 核心操作 执行商品扣减
// DecreaseStock 扣减库存（乐观锁 + 自旋重试）
// 1.上下文  2.商品id 3.库存 -- 参数
func (r *InventoryRepo) DecreaseStock(ctx context.Context, productID int, quantity int) error {
	// TODO 这里可以考虑实现RWLock
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

// DecreaseStockWithTx 在外部事务中扣减库存
func (r *InventoryRepo) DecreaseStockWithTx(tx *gorm.DB, productID int, quantity int) error {
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
}
