package domain

import "time"

type Address struct {
	ID           uint `gorm:"primaryKey"`
	UserID       uint
	AddressLine1 string
	AddressLine2 string
	City         string
	State        string
	Country      string
	PostalCode   string
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}
