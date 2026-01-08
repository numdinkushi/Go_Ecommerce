package domain

import "time"

type Payment struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UserId        uint      `json:"user_id"`
	CaptureMethod string    `json:"capture_method"`
	Amount        float64   `json:"amount"`
	TransactionId uint      `json:"transaction_id"` // References Transaction.ID
	CustomerId    string    `json:"customer_id"`    // stripe customer if
	PaymentId     string    `json:"payment_id"`     // payment id
	Status        string    `json:"status"`         // initial, success, failed
	Response      string    `json:"response"`
	CreatedAt     time.Time `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}

