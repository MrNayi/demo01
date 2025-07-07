package database

import (
	"demo01/internal/model"
	"log"

	"gorm.io/gorm"
)

// InitDatabase 初始化数据库
func InitDatabase(db *gorm.DB) error {
	// 1. 自动迁移数据库表结构
	if err := db.AutoMigrate(&model.Order{}, &model.Inventory{}, &model.Product{}); err != nil {
		return err
	}

	// 2. 初始化测试数据
	if err := initTestData(db); err != nil {
		return err
	}

	log.Println("数据库初始化完成")
	return nil
}

// initTestData 初始化测试数据
func initTestData(db *gorm.DB) error {
	// 检查是否已有数据
	var productCount int64
	db.Model(&model.Product{}).Count(&productCount)
	if productCount == 0 {
		log.Println("开始初始化测试商品数据...")

		// 插入测试商品数据
		testProducts := []model.Product{
			{
				Name:        "iPhone 15",
				Description: "苹果最新手机，搭载A17 Pro芯片",
				Price:       5999.00,
				Stock:       100,
				Category:    "手机",
				Status:      "active",
			},
			{
				Name:        "MacBook Pro",
				Description: "专业级笔记本电脑，M3芯片",
				Price:       12999.00,
				Stock:       50,
				Category:    "电脑",
				Status:      "active",
			},
			{
				Name:        "AirPods Pro",
				Description: "无线降噪耳机，空间音频",
				Price:       1999.00,
				Stock:       200,
				Category:    "配件",
				Status:      "active",
			},
			{
				Name:        "iPad Air",
				Description: "轻薄平板电脑，M2芯片",
				Price:       4399.00,
				Stock:       80,
				Category:    "平板",
				Status:      "active",
			},
			{
				Name:        "Apple Watch",
				Description: "智能手表，健康监测",
				Price:       2999.00,
				Stock:       150,
				Category:    "配件",
				Status:      "active",
			},
		}

		for _, product := range testProducts {
			if err := db.Create(&product).Error; err != nil {
				return err
			}
		}

		log.Printf("成功初始化 %d 个测试商品", len(testProducts))
	}

	// 检查库存数据
	var inventoryCount int64
	db.Model(&model.Inventory{}).Count(&inventoryCount)
	if inventoryCount == 0 {
		log.Println("开始初始化测试库存数据...")

		// 插入测试库存数据（与商品数据保持一致）
		testInventories := []model.Inventory{
			{ProductID: 1, Stock: 100, Version: 0},
			{ProductID: 2, Stock: 50, Version: 0},
			{ProductID: 3, Stock: 200, Version: 0},
			{ProductID: 4, Stock: 80, Version: 0},
			{ProductID: 5, Stock: 150, Version: 0},
		}

		for _, inventory := range testInventories {
			if err := db.Create(&inventory).Error; err != nil {
				return err
			}
		}

		log.Printf("成功初始化 %d 个测试库存记录", len(testInventories))
	}

	return nil
}
