// cmd/main.go
package main

import (
	"demo01/config"
	"demo01/internal/database"
	"demo01/internal/handler"
	"demo01/internal/repository"
	"demo01/internal/service"
	"demo01/internal/util"

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

	// 初始化Redis
	util.InitRedis(cfg.RedisAddr, cfg.RedisPwd)

	// 初始化数据库
	if err := database.InitDatabase(db); err != nil {
		panic("数据库初始化失败: " + err.Error())
	}

	// 依赖注入
	orderRepo := repository.NewOrderRepo(db, util.RedisClient)
	inventoryRepo := repository.NewInventoryRepo(db)
	productRepo := repository.NewProductRepo(db)

	orderService := service.NewOrderService(orderRepo, inventoryRepo)
	productService := service.NewProductService(productRepo)

	orderHandler := handler.NewOrderHandler(orderService)
	productHandler := handler.NewProductHandler(productService)
	healthHandler := handler.NewHealthHandler()

	// 初始化Gin
	r := gin.Default()

	// 健康检查路由
	r.GET("/health", healthHandler.HealthCheck)

	// 路由注册
	// 订单相关路由 orders
	{
		r.POST("/orders", orderHandler.CreateOrderHandler)
		r.GET("/orders", orderHandler.GetAllOrdersHandler)
		r.GET("/orders/:id", orderHandler.GetOrderHandler)
	}

	// 商品相关路由 products
	{
		r.POST("/products", productHandler.CreateProductHandler)
		r.GET("/products", productHandler.GetAllProductsHandler)
		r.GET("/products/:id", productHandler.GetProductHandler)
		r.GET("/products/:id/stock", productHandler.GetStockHandler)
		r.GET("/products/recommend", productHandler.RecommendProductsHandler)
		r.GET("/products/recommend_serial", productHandler.RecommendProductsSerialHandler)
	}

	// TODO:用户相关路由 users
	{

	}

	// 启动服务（绑定到所有网络接口）
	if err := r.Run("0.0.0.0:" + cfg.Port); err != nil {
		panic("启动服务失败: " + err.Error())
	}
}
