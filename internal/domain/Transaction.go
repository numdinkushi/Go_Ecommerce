package domain

import "time"

type Transaction struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	UserID         uint      `json:"user_id" gorm:"not null"`
	OrderID        uint      `json:"order_id" gorm:"not null"`
	Amount         float64   `json:"amount" gorm:"not null"`
	Status         string    `json:"status" gorm:"default:pending"` // pending, completed, failed, refunded
	PaymentMethod  string    `json:"payment_method"`                // card, bank_transfer, etc.
	PaymentGateway string    `json:"payment_gateway"`               // flutterwave, stripe, etc.
	TransactionID  string    `json:"transaction_id"`                // Gateway transaction ID
	PaymentID      string    `json:"payment_id"`                    // Gateway payment ID
	Currency       string    `json:"currency" gorm:"default:NGN"`
	CreatedAt      time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

