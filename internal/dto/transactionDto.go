package dto

type CreateTransactionInput struct {
	OrderID        uint    `json:"order_id" binding:"required"`
	Amount         float64 `json:"amount" binding:"required"`
	PaymentMethod  string  `json:"payment_method"`
	PaymentGateway string  `json:"payment_gateway"`
	TransactionID  string  `json:"transaction_id"`
	PaymentID      string  `json:"payment_id"`
	Currency       string  `json:"currency"`
}

type MakePaymentInput struct {
	OrderID        uint    `json:"order_id" binding:"required"`
	Amount         float64 `json:"amount" binding:"required"`
	PaymentMethod  string  `json:"payment_method"`
	PaymentGateway string  `json:"payment_gateway"`
	TransactionID  string  `json:"transaction_id"`
	PaymentID      string  `json:"payment_id"`
	Currency       string  `json:"currency"`
}
