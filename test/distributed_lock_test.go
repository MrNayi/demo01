package test

import (
	"context"
	"demo01/config"
	"demo01/internal/util"
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestDistributedLockBasic 测试分布式锁基本功能
func TestDistributedLockBasic(t *testing.T) {
	// 加载全局配置并初始化Redis连接
	cfg := config.Load()
	util.InitRedis(cfg.RedisAddr, cfg.RedisPwd)

	ctx := context.Background()
	keyGenerator := util.NewLockKeyGenerator()
	lockKey := keyGenerator.GenerateInventoryLockKey(123)

	// 创建分布式锁
	lock := util.NewDistributedLock(util.RedisClient, lockKey, 10*time.Second)

	// 测试获取锁
	locked, err := lock.TryLock(ctx)
	if err != nil {
		t.Fatalf("获取锁失败: %v", err)
	}
	if !locked {
		t.Fatal("应该成功获取锁")
	}

	// 测试重复获取锁（应该失败）
	locked2, err := lock.TryLock(ctx)
	if err != nil {
		t.Fatalf("重复获取锁时发生错误: %v", err)
	}
	if locked2 {
		t.Fatal("重复获取锁应该失败")
	}

	// 测试释放锁
	err = lock.Unlock(ctx)
	if err != nil {
		t.Fatalf("释放锁失败: %v", err)
	}

	// 测试释放后重新获取锁
	locked3, err := lock.TryLock(ctx)
	if err != nil {
		t.Fatalf("释放后重新获取锁失败: %v", err)
	}
	if !locked3 {
		t.Fatal("释放后应该能重新获取锁")
	}

	// 清理
	lock.Unlock(ctx)
}

// TestDistributedLockConcurrency 测试分布式锁并发安全性
func TestDistributedLockConcurrency(t *testing.T) {
	// 加载全局配置并初始化Redis连接
	cfg := config.Load()
	util.InitRedis(cfg.RedisAddr, cfg.RedisPwd)

	ctx := context.Background()
	keyGenerator := util.NewLockKeyGenerator()
	lockKey := keyGenerator.GenerateInventoryLockKey(456)

	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	// 并发测试：10个goroutine同时尝试获取同一个锁
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			lock := util.NewDistributedLock(util.RedisClient, lockKey, 5*time.Second)

			// 尝试获取锁
			locked, err := lock.TryLock(ctx)
			if err != nil {
				t.Logf("Goroutine %d 获取锁时发生错误: %v", id, err)
				return
			}

			if locked {
				mu.Lock()
				successCount++
				mu.Unlock()

				t.Logf("Goroutine %d 成功获取锁", id)

				// 模拟业务处理时间
				time.Sleep(100 * time.Millisecond)

				// 释放锁
				if err := lock.Unlock(ctx); err != nil {
					t.Logf("Goroutine %d 释放锁失败: %v", id, err)
				}
			} else {
				t.Logf("Goroutine %d 获取锁失败", id)
			}
		}(i)
	}

	wg.Wait()

	// 验证只有一个goroutine成功获取锁
	if successCount != 1 {
		t.Fatalf("期望只有一个goroutine成功获取锁，实际有 %d 个", successCount)
	}

	t.Logf("并发测试通过，成功获取锁的次数: %d", successCount)
}

// TestDistributedLockRetry 测试分布式锁重试机制
func TestDistributedLockRetry(t *testing.T) {
	// 加载全局配置并初始化Redis连接
	cfg := config.Load()
	util.InitRedis(cfg.RedisAddr, cfg.RedisPwd)

	ctx := context.Background()
	keyGenerator := util.NewLockKeyGenerator()
	lockKey := keyGenerator.GenerateInventoryLockKey(789)

	// 先创建一个锁并持有
	lock1 := util.NewDistributedLock(util.RedisClient, lockKey, 2*time.Second)
	locked, err := lock1.TryLock(ctx)
	if err != nil || !locked {
		t.Fatalf("无法获取第一个锁")
	}

	// 创建第二个锁并尝试获取（应该失败，然后重试）
	lock2 := util.NewDistributedLock(util.RedisClient, lockKey, 5*time.Second)

	startTime := time.Now()
	locked2, err := lock2.TryLockWithRetry(ctx, 3, 200*time.Millisecond)
	duration := time.Since(startTime)

	// 第一个锁应该在2秒后自动过期
	if locked2 {
		t.Logf("第二个锁在 %v 后成功获取", duration)
	} else {
		t.Logf("第二个锁获取失败，耗时 %v", duration)
	}

	// 清理
	lock1.Unlock(ctx)
	if locked2 {
		lock2.Unlock(ctx)
	}
}

// TestLockKeyGenerator 测试锁key生成器
func TestLockKeyGenerator(t *testing.T) {
	keyGenerator := util.NewLockKeyGenerator()

	// 测试库存锁key
	inventoryKey := keyGenerator.GenerateInventoryLockKey(123)
	expectedInventoryKey := "lock:inventory:123"
	if inventoryKey != expectedInventoryKey {
		t.Fatalf("库存锁key生成错误，期望: %s, 实际: %s", expectedInventoryKey, inventoryKey)
	}

	// 测试订单锁key
	orderKey := keyGenerator.GenerateOrderLockKey("order_20250708135036_test_user")
	expectedOrderKey := "lock:order:order_20250708135036_test_user"
	if orderKey != expectedOrderKey {
		t.Fatalf("订单锁key生成错误，期望: %s, 实际: %s", expectedOrderKey, orderKey)
	}

	// 测试用户锁key
	userKey := keyGenerator.GenerateUserLockKey("test_user")
	expectedUserKey := "lock:user:test_user"
	if userKey != expectedUserKey {
		t.Fatalf("用户锁key生成错误，期望: %s, 实际: %s", expectedUserKey, userKey)
	}

	// 测试商品锁key
	productKey := keyGenerator.GenerateProductLockKey(456)
	expectedProductKey := "lock:product:456"
	if productKey != expectedProductKey {
		t.Fatalf("商品锁key生成错误，期望: %s, 实际: %s", expectedProductKey, productKey)
	}

	t.Log("锁key生成器测试通过")
}

// TestDistributedLockExpiration 测试分布式锁过期机制
func TestDistributedLockExpiration(t *testing.T) {
	// 加载全局配置并初始化Redis连接
	cfg := config.Load()
	util.InitRedis(cfg.RedisAddr, cfg.RedisPwd)

	ctx := context.Background()
	keyGenerator := util.NewLockKeyGenerator()
	lockKey := keyGenerator.GenerateInventoryLockKey(999)

	// 创建一个短过期时间的锁
	lock := util.NewDistributedLock(util.RedisClient, lockKey, 1*time.Second)

	locked, err := lock.TryLock(ctx)
	if err != nil || !locked {
		t.Fatalf("无法获取锁")
	}

	// 等待锁过期
	time.Sleep(2 * time.Second)

	// 检查锁是否已过期
	isLocked, err := lock.IsLocked(ctx)
	if err != nil {
		t.Fatalf("检查锁状态失败: %v", err)
	}

	if isLocked {
		t.Fatal("锁应该已经过期")
	}

	// 尝试重新获取锁（应该成功）
	locked2, err := lock.TryLock(ctx)
	if err != nil || !locked2 {
		t.Fatal("锁过期后应该能重新获取")
	}

	// 清理
	lock.Unlock(ctx)

	t.Log("分布式锁过期机制测试通过")
}

// BenchmarkDistributedLock 分布式锁性能基准测试
func BenchmarkDistributedLock(b *testing.B) {
	// 加载全局配置并初始化Redis连接
	cfg := config.Load()
	util.InitRedis(cfg.RedisAddr, cfg.RedisPwd)

	ctx := context.Background()
	keyGenerator := util.NewLockKeyGenerator()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lockKey := keyGenerator.GenerateInventoryLockKey(i)
		lock := util.NewDistributedLock(util.RedisClient, lockKey, 1*time.Second)

		locked, err := lock.TryLock(ctx)
		if err != nil {
			b.Fatalf("获取锁失败: %v", err)
		}

		if locked {
			lock.Unlock(ctx)
		}
	}
}

// 运行所有测试的主函数
func TestAllDistributedLock(t *testing.T) {
	fmt.Println("开始运行分布式锁测试...")

	t.Run("Basic", TestDistributedLockBasic)
	t.Run("Concurrency", TestDistributedLockConcurrency)
	t.Run("Retry", TestDistributedLockRetry)
	t.Run("KeyGenerator", TestLockKeyGenerator)
	t.Run("Expiration", TestDistributedLockExpiration)

	fmt.Println("所有分布式锁测试完成")
}
