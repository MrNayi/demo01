package util

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// DistributedLock 分布式锁结构
type DistributedLock struct {
	client     *redis.Client
	key        string
	value      string
	expiration time.Duration
}

// LockKeyGenerator 锁key生成器
type LockKeyGenerator struct{}

// NewLockKeyGenerator 创建锁key生成器
func NewLockKeyGenerator() *LockKeyGenerator {
	return &LockKeyGenerator{}
}

// GenerateInventoryLockKey 生成库存锁的key
// 格式: lock:inventory:{product_id}
// 示例: lock:inventory:123    123号商品的库存锁
func (g *LockKeyGenerator) GenerateInventoryLockKey(productID int) string {
	return fmt.Sprintf("lock:inventory:%d", productID)
}

// GenerateOrderLockKey 生成订单锁的key
// 格式: lock:order:{order_id}
// 示例: lock:order:order_20250708135036_test_user
func (g *LockKeyGenerator) GenerateOrderLockKey(orderID string) string {
	return fmt.Sprintf("lock:order:%s", orderID)
}

// GenerateUserLockKey 生成用户锁的key
// 格式: lock:user:{user_id}
// 示例: lock:user:test_user
func (g *LockKeyGenerator) GenerateUserLockKey(userID string) string {
	return fmt.Sprintf("lock:user:%s", userID)
}

// GenerateProductLockKey 生成商品锁的key
// 格式: lock:product:{product_id}
// 示例: lock:product:123
func (g *LockKeyGenerator) GenerateProductLockKey(productID int) string {
	return fmt.Sprintf("lock:product:%d", productID)
}

// NewDistributedLock 创建分布式锁实例
func NewDistributedLock(client *redis.Client, key string, expiration time.Duration) *DistributedLock {
	return &DistributedLock{
		client:     client,
		key:        key,
		value:      generateLockValue(), // 生成唯一标识
		expiration: expiration,
	}
}

// generateLockValue 生成锁的唯一标识
func generateLockValue() string {
	return fmt.Sprintf("%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

// TryLock 尝试获取锁
// 使用SET NX EX命令实现原子性加锁
func (dl *DistributedLock) TryLock(ctx context.Context) (bool, error) {
	// SET key value NX EX seconds
	// NX: 只有当key不存在时才设置
	// EX: 设置过期时间（秒）
	result, err := dl.client.SetNX(ctx, dl.key, dl.value, dl.expiration).Result()
	if err != nil {
		return false, fmt.Errorf("获取分布式锁失败: %w", err)
	}
	return result, nil
}

// TryLockWithRetry 带重试的锁获取（优化版，使用指数退避）
func (dl *DistributedLock) TryLockWithRetry(ctx context.Context, maxRetries int, baseDelay time.Duration) (bool, error) {
	maxDelay := 500 * time.Millisecond // 最大重试间隔

	for i := 0; i < maxRetries; i++ {
		locked, err := dl.TryLock(ctx)
		if err != nil {
			return false, err
		}
		if locked {
			return true, nil
		}

		// 检查上下文是否已取消
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
			// 继续重试
		}

		// 如果不是最后一次重试，则等待后继续
		if i < maxRetries-1 {
			// 指数退避：延迟时间随重试次数增加
			delay := baseDelay * time.Duration(1<<i)
			if delay > maxDelay {
				delay = maxDelay
			}

			select {
			case <-ctx.Done():
				return false, ctx.Err()
			case <-time.After(delay):
				// 继续重试
			}
		}
	}
	return false, fmt.Errorf("获取锁失败，已达到最大重试次数: %d", maxRetries)
}

// Unlock 释放锁
// 使用Lua脚本保证原子性解锁
func (dl *DistributedLock) Unlock(ctx context.Context) error {
	// Lua脚本：只有锁的value匹配时才删除
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := dl.client.Eval(ctx, script, []string{dl.key}, []interface{}{dl.value}).Result()
	if err != nil {
		return fmt.Errorf("释放分布式锁失败: %w", err)
	}

	// 检查删除结果
	if result.(int64) == 0 {
		return fmt.Errorf("锁不存在或已被其他进程持有")
	}

	return nil
}

// ExtendLock 延长锁的过期时间
func (dl *DistributedLock) ExtendLock(ctx context.Context, newExpiration time.Duration) error {
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("expire", KEYS[1], ARGV[2])
		else
			return 0
		end
	`

	result, err := dl.client.Eval(ctx, script, []string{dl.key}, []interface{}{dl.value, newExpiration.Seconds()}).Result()
	if err != nil {
		return fmt.Errorf("延长锁过期时间失败: %w", err)
	}

	if result.(int64) == 0 {
		return fmt.Errorf("锁不存在或已被其他进程持有")
	}

	return nil
}

// IsLocked 检查锁是否被持有
func (dl *DistributedLock) IsLocked(ctx context.Context) (bool, error) {
	value, err := dl.client.Get(ctx, dl.key).Result()
	if err == redis.Nil {
		return false, nil // 锁不存在
	}
	if err != nil {
		return false, fmt.Errorf("检查锁状态失败: %w", err)
	}
	return value == dl.value, nil
}

// GetLockTTL 获取锁的剩余过期时间
func (dl *DistributedLock) GetLockTTL(ctx context.Context) (time.Duration, error) {
	ttl, err := dl.client.TTL(ctx, dl.key).Result()
	if err != nil {
		return 0, fmt.Errorf("获取锁TTL失败: %w", err)
	}
	return ttl, nil
}

// AutoExtendLock 自动续期锁（用于长时间任务）
func (dl *DistributedLock) AutoExtendLock(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := dl.ExtendLock(ctx, dl.expiration); err != nil {
				// 续期失败，可能是锁已被释放
				return
			}
		}
	}
}

// WithLock 使用锁执行函数（推荐用法）
func (dl *DistributedLock) WithLock(ctx context.Context, fn func() error) error {
	// 尝试获取锁
	locked, err := dl.TryLockWithRetry(ctx, 3, 50*time.Millisecond)
	if err != nil {
		return fmt.Errorf("获取锁失败: %w", err)
	}
	if !locked {
		return fmt.Errorf("无法获取锁")
	}

	// 确保释放锁
	defer func() {
		if unlockErr := dl.Unlock(ctx); unlockErr != nil {
			// 记录解锁错误，但不影响主流程
			fmt.Printf("解锁失败: %v\n", unlockErr)
		}
	}()

	// 执行业务逻辑
	return fn()
}
