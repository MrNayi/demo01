package model

import (
	"time"
)

// Product 商品模型
type Product struct {
	ID          int       `json:"id" gorm:"primaryKey;autoIncrement"`
	Name        string    `json:"name" gorm:"size:100;not null;comment:商品名称"`
	Description string    `json:"description" gorm:"size:500;comment:商品描述"`
	Price       float64   `json:"price" gorm:"type:decimal(10,2);not null;comment:商品价格"`
	Stock       int       `json:"stock" gorm:"not null;default:0;comment:库存数量"`
	Category    string    `json:"category" gorm:"size:50;comment:商品分类"`
	Status      string    `json:"status" gorm:"size:20;default:'active';comment:商品状态"`
	Version     int       `json:"version" gorm:"default:0;comment:乐观锁版本号"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName 指定表名
func (Product) TableName() string {
	return "products"
}
