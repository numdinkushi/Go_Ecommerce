package domain

import "time"

type Order struct {
	ID             uint        `json:"id" gorm:"primaryKey"`
	UserID         uint        `json:"user_id" gorm:"not null"`
	Status         string      `json:"status" gorm:"default:pending"`
	Amount         float64     `json:"amount" gorm:"not null"`
	TransactionId  string      `json:"transaction_id"`
	OrderRefNumber uint        `json:"order_ref_number" gorm:"not null;unique"`
	PaymentId      string      `json:"payment_id"`
	Items          []OrderItem `json:"items" gorm:"foreignKey:OrderID"`
	CreatedAt      time.Time   `json:"created_at" gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt      time.Time   `json:"updated_at" gorm:"default:CURRENT_TIMESTAMP"`
}
