package dto

type VerifyAccountInput struct {
	AccountNumber string `json:"account_number" validate:"required"`
	BankCode      string `json:"bank_code" validate:"required"`
}
