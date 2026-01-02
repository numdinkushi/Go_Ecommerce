package domain

import "time"

type Cart struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint
	SellerID  uint
	Name      string
	ImageURL  string
	Price     float64
	Quantity  int
	ProductID uint
	CreatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}
