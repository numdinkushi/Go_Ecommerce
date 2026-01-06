package repository

import (
	"go-ecommerce-app/internal/domain"

	"gorm.io/gorm"
)

type TransactionRepository interface {
	CreateTransaction(transaction *domain.Transaction) (*domain.Transaction, error)
	FindTransactionByID(id uint) (*domain.Transaction, error)
	FindTransactionsByUserID(userID uint) ([]domain.Transaction, error)
	FindTransactionByOrderID(orderID uint) (*domain.Transaction, error)
	UpdateTransaction(transaction *domain.Transaction) (*domain.Transaction, error)
}

type transactionRepository struct {
	DB *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) TransactionRepository {
	return &transactionRepository{DB: db}
}

func (r *transactionRepository) CreateTransaction(transaction *domain.Transaction) (*domain.Transaction, error) {
	err := r.DB.Create(transaction).Error
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

func (r *transactionRepository) FindTransactionByID(id uint) (*domain.Transaction, error) {
	var transaction domain.Transaction
	err := r.DB.First(&transaction, id).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) FindTransactionsByUserID(userID uint) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	err := r.DB.Where("user_id = ?", userID).Find(&transactions).Error
	if err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *transactionRepository) FindTransactionByOrderID(orderID uint) (*domain.Transaction, error) {
	var transaction domain.Transaction
	err := r.DB.Where("order_id = ?", orderID).First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) UpdateTransaction(transaction *domain.Transaction) (*domain.Transaction, error) {
	err := r.DB.Save(transaction).Error
	if err != nil {
		return nil, err
	}
	return transaction, nil
}
