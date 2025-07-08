package config

import (
	"os"
)

// Config 应用配置
type Config struct {
	MySQLDSN  string // MySQL连接字符串
	RedisAddr string // Redis地址
	RedisPwd  string // Redis密码
	Port      string // 服务端口
}

// Load 加载配置
func Load() *Config {
	return &Config{
		MySQLDSN:  getEnv("MYSQL_DSN", "root:Azspigot1996@tcp(14.103.163.34:3306)/nayidemo?charset=utf8mb4&parseTime=True&loc=Local"),
		RedisAddr: getEnv("REDIS_ADDR", "14.103.163.34:6379"),
		RedisPwd:  getEnv("REDIS_PWD", "Azspigot1996"),
		Port:      getEnv("PORT", "8080"),
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
