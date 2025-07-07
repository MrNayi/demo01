// internal/model/order.go
package model

import "time"

// Order 订单模型
type Order struct {
	ID          string    `gorm:"type:varchar(32);primaryKey" json:"id"`
	UserID      string    `gorm:"type:varchar(32)" json:"user_id"`
	Items       string    `gorm:"type:text" json:"items"` // 存储JSON格式的商品列表
	TotalAmount float64   `json:"total_amount"`
	Status      string    `gorm:"size:20;default:'pending'" json:"status"` // pending, paid, shipped, completed, cancelled
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Inventory 库存模型
type Inventory struct {
	ProductID int `gorm:"primaryKey" json:"product_id"`
	Stock     int `gorm:"type:int" json:"stock"`
	Version   int `gorm:"type:int" json:"version"` // 乐观锁版本号
}

// OrderItem 订单商品项
type OrderItem struct {
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}
