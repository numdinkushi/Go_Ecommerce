package service

import (
	"errors"
	"go-ecommerce-app/internal/domain"
	"go-ecommerce-app/internal/dto"
	"go-ecommerce-app/internal/helper"
	"go-ecommerce-app/internal/repository"
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
	// Business logic: Set default currency
	currency := input.Currency
	if currency == "" {
		currency = "NGN"
	}

	// Business logic: Create transaction for payment
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
