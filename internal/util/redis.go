// internal/util/redis.go（go-redis 用法）
package util

import (
	"context"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

// InitRedis 初始化Redis
func InitRedis(addr string) {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

// GetOrderCache 从Redis查询订单缓存
// context 上下文 用于传递请求的上下文信息
func GetOrderCache(ctx context.Context, orderID string) (string, error) {
	return RedisClient.Get(ctx, "order:"+orderID).Result()
}
