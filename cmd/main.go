// cmd/main.go
package main

import (
	"demo01/config"
	"demo01/internal/handler"
	"demo01/internal/model"
	"demo01/internal/repository"
	"demo01/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 初始化配置
	cfg := config.Load()

	// 初始化GORM
	db, err := gorm.Open(mysql.Open(cfg.MySQLDSN), &gorm.Config{})
	if err != nil {
		panic("连接数据库失败: " + err.Error())
	}

	// 初始化Redis（暂时禁用）
	// util.InitRedis(cfg.RedisAddr)

	// 初始化Kafka Producer（暂时禁用）
	// kafkaProducer, err := kafka.NewProducer(&kafka.ConfigMap{
	// 	"bootstrap.servers": cfg.KafkaAddr,
	// })
	// if err != nil {
	// 	panic("初始化Kafka失败: " + err.Error())
	// }
	// defer kafkaProducer.Close()

	// 自动迁移数据库表结构
	if err := db.AutoMigrate(&model.Order{}, &model.Inventory{}); err != nil {
		panic("数据库迁移失败: " + err.Error())
	}

	// 添加测试数据
	var count int64
	db.Model(&model.Inventory{}).Count(&count)
	if count == 0 {
		// 插入测试库存数据
		testInventories := []model.Inventory{
			{ProductID: 1, Stock: 100, Version: 0},
			{ProductID: 2, Stock: 50, Version: 0},
			{ProductID: 3, Stock: 200, Version: 0},
		}
		for _, inventory := range testInventories {
			db.Create(&inventory)
		}
	}

	// 依赖注入（暂时传入 nil Kafka redis功能暂时禁用）
	orderRepo := repository.NewOrderRepo(db)
	inventoryRepo := repository.NewInventoryRepo(db)
	orderService := service.NewOrderService(orderRepo, inventoryRepo, nil)
	orderHandler := handler.NewOrderHandler(orderService)

	// 初始化Gin
	r := gin.Default()
	// 路由注册
	r.POST("/orders", orderHandler.CreateOrderHandler)
	r.GET("/orders", orderHandler.GetAllOrdersHandler)
	r.GET("/orders/:id", orderHandler.GetOrderHandler)

	// 启动服务（绑定到所有网络接口）
	if err := r.Run("0.0.0.0:" + cfg.Port); err != nil {
		panic("启动服务失败: " + err.Error())
	}
}
