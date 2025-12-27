package domain

import "time"

type User struct {
	ID        uint    `json:"id" gorm:"primaryKey"`
	FirstName string `json:"first_name" gorm:"not null"`
	LastName  string `json:"last_name" gorm:"not null"`
	Email     string `json:"email" gorm:"not null;unique"`
	Phone     string `json:"phone" gorm:"index;not null;unique"`
	Code      int `json:"code" gorm:"not null"`
	Expiry    time.Time `json:"expiry"`
	Verified  bool `json:"verified" gorm:"default:false"`
	Password  string `json:"password"`
	UserType  string `json:"user_type" gorm:"default:buyer"`
	CreatedAt time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

