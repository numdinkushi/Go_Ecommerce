package domain

import "time"

type Category struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Name         string    `json:"name" gorm:"not null"`
	ParentID     *uint     `json:"parent_id"`
	ImageURL     string    `json:"image_url" gorm:"not null"`
	SellerID     uint      `json:"seller_id" gorm:"not null"`
	Products     []Product `json:"products" gorm:"foreignKey:CategoryID"`
	DisplayOrder int       `json:"display_order" gorm:"default:0"`
	Description  string    `json:"description" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}
