package domain

import "time"

type OrderItem struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	OrderID   uint      `json:"order_id" gorm:"not null"`
	ProductID uint      `json:"product_id" gorm:"not null"`
	Name      string    `json:"name" gorm:"not null"`
	ImageUrl  string    `json:"image_url"`
	SellerId  uint      `json:"seller_id" gorm:"not null"`
	Price     float64   `json:"price" gorm:"not null"`
	Quantity  uint      `json:"quantity" gorm:"column:qty;not null"`
	CreatedAt time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"` 
}
