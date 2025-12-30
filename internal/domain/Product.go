package domain

import "time"

type Product struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	Price       float64   `json:"price" gorm:"not null"`
	CategoryID  uint      `json:"category_id" gorm:"not null"`
	Stock       int       `json:"stock" gorm:"default:0"`
	ImageURL    string    `json:"image_url"`
	SellerID    uint      `json:"seller_id" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}
