package service

import (
	"go-ecommerce-app/internal/domain"
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

func (s *TransactionService) CreateTransaction(transaction *domain.Transaction) (*domain.Transaction, error) {
	createdTransaction, err := s.Repo.CreateTransaction(transaction)
	if err != nil {
		return nil, err
	}
	return createdTransaction, nil
}

func (s *TransactionService) GetTransactionsByUserID(userID uint) ([]domain.Transaction, error) {
	transactions, err := s.Repo.FindTransactionsByUserID(userID)
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (s *TransactionService) GetTransactionByID(id uint) (*domain.Transaction, error) {
	transaction, err := s.Repo.FindTransactionByID(id)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

