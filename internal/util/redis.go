// internal/util/redis.go（go-redis/v9 用法）
package util

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// InitRedis 初始化Redis客户端
func InitRedis(addr, password string) {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           0,               // 使用默认数据库
		PoolSize:     10,              // 连接池大小
		MinIdleConns: 5,               // 最小空闲连接数
		MaxRetries:   3,               // 最大重试次数
		DialTimeout:  5 * time.Second, // 连接超时
		ReadTimeout:  3 * time.Second, // 读取超时
		WriteTimeout: 3 * time.Second, // 写入超时
		PoolTimeout:  4 * time.Second, // 连接池超时
	})
}

// PingRedis 测试Redis连接
func PingRedis(ctx context.Context) error {
	_, err := RedisClient.Ping(ctx).Result()
	return err
}

// GetOrderCache 从Redis查询订单缓存
// context 上下文 用于传递请求的上下文信息
func GetOrderCache(ctx context.Context, orderID string) (string, error) {
	return RedisClient.Get(ctx, "order:"+orderID).Result()
}

// SetOrderCache 设置订单缓存
func SetOrderCache(ctx context.Context, orderID string, data string, expiration time.Duration) error {
	return RedisClient.Set(ctx, "order:"+orderID, data, expiration).Err()
}

// DelOrderCache 删除订单缓存
func DelOrderCache(ctx context.Context, orderID string) error {
	return RedisClient.Del(ctx, "order:"+orderID).Err()
}

// GetProductCache 从Redis查询商品缓存
func GetProductCache(ctx context.Context, productID int) (string, error) {
	return RedisClient.Get(ctx, "product:"+string(rune(productID))).Result()
}

// SetProductCache 设置商品缓存
func SetProductCache(ctx context.Context, productID int, data string, expiration time.Duration) error {
	return RedisClient.Set(ctx, "product:"+string(rune(productID)), data, expiration).Err()
}

// DelProductCache 删除商品缓存
func DelProductCache(ctx context.Context, productID int) error {
	return RedisClient.Del(ctx, "product:"+string(rune(productID))).Err()
}

// CloseRedis 关闭Redis连接
func CloseRedis() error {
	return RedisClient.Close()
}
