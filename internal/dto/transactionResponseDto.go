package dto

import "time"

type TransactionResponse struct {
	ID             uint      `json:"id"`
	UserID         uint      `json:"user_id"`
	OrderID        uint      `json:"order_id"`
	Amount         float64   `json:"amount"`
	Status         string    `json:"status"`
	PaymentMethod  string    `json:"payment_method"`
	PaymentGateway string    `json:"payment_gateway"`
	TransactionID  string    `json:"transaction_id"`
	PaymentID      string    `json:"payment_id"`
	Currency       string    `json:"currency"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type OrderItemResponse struct {
	ID        uint      `json:"id"`
	ProductID uint      `json:"product_id"`
	Name      string    `json:"name"`
	ImageURL  string    `json:"image_url"`
	SellerID  uint      `json:"seller_id"`
	Price     float64   `json:"price"`
	Quantity  uint      `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrderResponse struct {
	ID             uint                `json:"id"`
	UserID         uint                `json:"user_id"`
	Status         string              `json:"status"`
	Amount         float64             `json:"amount"`
	TransactionID  string              `json:"transaction_id"`
	OrderRefNumber uint                `json:"order_ref_number"`
	PaymentID      string              `json:"payment_id"`
	Items          []OrderItemResponse `json:"items"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

type SellerOrderResponse struct {
	ID             uint                `json:"id"`
	UserID         uint                `json:"user_id"`
	Status         string              `json:"status"`
	Amount         float64             `json:"amount"`
	TransactionID  string              `json:"transaction_id"`
	OrderRefNumber uint                `json:"order_ref_number"`
	PaymentID      string              `json:"payment_id"`
	Items          []OrderItemResponse `json:"items"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
}

type SellerOrderDetails struct {
	OrderRefNumber  uint   `json:"order_ref_number"`
	OrderStatus     string `json:"order_status"`
	CreatedAt       string `json:"created_at"`
	OrderItemId     uint   `json:"order_item_id"`
	ProductId       uint   `json:"product_id"`
	Name            string `json:"name"`
	ImageUrl        string `json:"image_url"`
	Price           string `json:"price"`
	Qty             uint   `json:"qty"`
	CustomerName    string `json:"customer_name"`
	CustomerEmail   string `json:"customer_email"`
	CustomerPhone   string `json:"customer_phone"`
	CustomerAddress string `json:"customer_address"`
}
