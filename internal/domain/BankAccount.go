package domain

import "time"

type BankAccount struct {
	ID                uint      `json:"id" gorm:"primaryKey"`
	UserId            uint      `json:"user_id" gorm:"not null"`
	BankName          string    `json:"bank_name" gorm:"not null"`
	BankAccountNumber string    `json:"bank_account_number" gorm:"not null"`
	BankCode          string    `json:"bank_code" gorm:"not null"`
	CreatedAt         time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt         time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}
