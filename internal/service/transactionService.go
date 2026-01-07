package service

import (
	"errors"
	"go-ecommerce-app/internal/domain"
	"go-ecommerce-app/internal/dto"
	"go-ecommerce-app/internal/helper"
	"go-ecommerce-app/internal/repository"
	"go-ecommerce-app/pkg/payment"
	"os"
)

type TransactionService struct {
	Repo repository.TransactionRepository
	Auth helper.Auth
}

func NewTransactionService(repo repository.TransactionRepository, auth helper.Auth) *TransactionService {
	return &TransactionService{
		Repo: repo,
		Auth: auth,
	}
}

func (s *TransactionService) CreateTransaction(userID uint, input dto.CreateTransactionInput) (*dto.TransactionResponse, error) {
	// Business logic: Set default currency
	currency := input.Currency
	if currency == "" {
		currency = "NGN"
	}

	// Business logic: Create transaction domain model
	transaction := &domain.Transaction{
		UserID:         userID,
		OrderID:        input.OrderID,
		Amount:         input.Amount,
		Status:         "pending",
		PaymentMethod:  input.PaymentMethod,
		PaymentGateway: input.PaymentGateway,
		TransactionID:  input.TransactionID,
		PaymentID:      input.PaymentID,
		Currency:       currency,
	}

	createdTransaction, err := s.Repo.CreateTransaction(transaction)
	if err != nil {
		return nil, err
	}

	// Business logic: Transform to DTO
	return s.toTransactionResponse(createdTransaction), nil
}

func (s *TransactionService) ProcessPayment(userID uint, input dto.MakePaymentInput) (*dto.TransactionResponse, error) {
	// Validate payment gateway
	if input.PaymentGateway == "" {
		return nil, errors.New("payment gateway is required")
	}

	// Get payment provider
	provider, err := payment.GetProvider(input.PaymentGateway)
	if err != nil {
		return nil, errors.New("payment provider not available: " + input.PaymentGateway)
	}

	// Set default currency
	currency := input.Currency
	if currency == "" {
		currency = "NGN"
	}

	// Get success and failure URLs from environment
	successURL := os.Getenv("PAYMENT_SUCCESS_URL")
	if successURL == "" {
		successURL = "http://localhost:3000/payment/success"
	}

	failureURL := os.Getenv("PAYMENT_FAILURE_URL")
	if failureURL == "" {
		failureURL = "http://localhost:3000/payment/failure"
	}

	// Create payment session using provider
	sessionReq := payment.CreatePaymentSessionRequest{
		Amount:     input.Amount,
		Currency:   currency,
		UserID:     userID,
		OrderID:    input.OrderID,
		SuccessURL: successURL,
		FailureURL: failureURL,
		Metadata:   make(map[string]string),
	}

	sessionResp, err := provider.CreatePaymentSession(sessionReq)
	if err != nil {
		return nil, err
	}

	// Create transaction record
	transaction := &domain.Transaction{
		UserID:         userID,
		OrderID:        input.OrderID,
		Amount:         input.Amount,
		Status:         "pending",
		PaymentMethod:  input.PaymentMethod,
		PaymentGateway: input.PaymentGateway,
		TransactionID:  sessionResp.SessionID,
		PaymentID:      sessionResp.PaymentID,
		Currency:       currency,
	}

	createdTransaction, err := s.Repo.CreateTransaction(transaction)
	if err != nil {
		return nil, err
	}

	// Transform to DTO and include checkout URL
	response := s.toTransactionResponse(createdTransaction)
	response.CheckoutURL = sessionResp.CheckoutURL

	return response, nil
}

func (s *TransactionService) GetTransactionsByUserID(userID uint) ([]dto.TransactionResponse, error) {
	transactions, err := s.Repo.FindTransactionsByUserID(userID)
	if err != nil {
		return nil, err
	}

	// Business logic: Transform to DTOs
	responses := make([]dto.TransactionResponse, len(transactions))
	for i, transaction := range transactions {
		responses[i] = *s.toTransactionResponse(&transaction)
	}

	return responses, nil
}

func (s *TransactionService) GetTransactionByIDAndUserID(id uint, userID uint) (*dto.TransactionResponse, error) {
	transaction, err := s.Repo.FindTransactionByID(id)
	if err != nil {
		return nil, err
	}

	// Business logic: Verify ownership
	if transaction.UserID != userID {
		return nil, errors.New("transaction not found")
	}

	// Business logic: Transform to DTO
	return s.toTransactionResponse(transaction), nil
}

// VerifyPaymentStatus verifies the payment status with the payment provider
func (s *TransactionService) VerifyPaymentStatus(transactionID uint, gateway string) (*dto.TransactionResponse, error) {
	// Get payment provider
	provider, err := payment.GetProvider(gateway)
	if err != nil {
		return nil, errors.New("payment provider not available: " + gateway)
	}

	// Get transaction from DB
	transaction, err := s.Repo.FindTransactionByID(transactionID)
	if err != nil {
		return nil, err
	}

	// Query provider for current status
	statusResp, err := provider.GetPaymentStatus(transaction.TransactionID)
	if err != nil {
		return nil, err
	}

	// Update transaction status in DB
	transaction.Status = statusResp.Status
	if statusResp.PaymentID != "" {
		transaction.PaymentID = statusResp.PaymentID
	}

	updatedTransaction, err := s.Repo.UpdateTransaction(transaction)
	if err != nil {
		return nil, err
	}

	return s.toTransactionResponse(updatedTransaction), nil
}

// ProcessWebhook processes a webhook event from a payment provider
func (s *TransactionService) ProcessWebhook(gateway string, payload []byte, signature string) error {
	// Get payment provider
	provider, err := payment.GetProvider(gateway)
	if err != nil {
		return errors.New("payment provider not available: " + gateway)
	}

	// Process webhook
	webhookEvent, err := provider.ProcessWebhook(payload, signature)
	if err != nil {
		return err
	}

	// Find transaction by session ID or payment ID
	var transaction *domain.Transaction
	if webhookEvent.SessionID != "" {
		transaction, err = s.Repo.FindTransactionByTransactionID(webhookEvent.SessionID)
		if err != nil {
			transaction = nil
		}
	}

	if transaction == nil && webhookEvent.PaymentID != "" {
		transaction, err = s.Repo.FindTransactionByPaymentID(webhookEvent.PaymentID)
		if err != nil {
			transaction = nil
		}
	}

	if transaction == nil {
		return errors.New("transaction not found for webhook event")
	}

	// Update transaction status
	transaction.Status = webhookEvent.Status
	if webhookEvent.PaymentID != "" {
		transaction.PaymentID = webhookEvent.PaymentID
	}

	_, err = s.Repo.UpdateTransaction(transaction)
	return err
}

func (s *TransactionService) toTransactionResponse(transaction *domain.Transaction) *dto.TransactionResponse {
	return &dto.TransactionResponse{
		ID:             transaction.ID,
		UserID:         transaction.UserID,
		OrderID:        transaction.OrderID,
		Amount:         transaction.Amount,
		Status:         transaction.Status,
		PaymentMethod:  transaction.PaymentMethod,
		PaymentGateway: transaction.PaymentGateway,
		TransactionID:  transaction.TransactionID,
		PaymentID:      transaction.PaymentID,
		Currency:       transaction.Currency,
		CreatedAt:      transaction.CreatedAt,
		UpdatedAt:      transaction.UpdatedAt,
	}
}
