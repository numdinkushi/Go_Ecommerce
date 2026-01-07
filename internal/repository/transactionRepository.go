package repository

import (
	"fmt"
	"go-ecommerce-app/internal/domain"
	"go-ecommerce-app/internal/dto"

	"gorm.io/gorm"
)

type TransactionRepository interface {
	CreateTransaction(transaction *domain.Transaction) (*domain.Transaction, error)
	FindTransactionByID(id uint) (*domain.Transaction, error)
	FindTransactionsByUserID(userID uint) ([]domain.Transaction, error)
	FindTransactionByOrderID(orderID uint) (*domain.Transaction, error)
	FindTransactionByTransactionID(transactionID string) (*domain.Transaction, error)
	FindTransactionByPaymentID(paymentID string) (*domain.Transaction, error)
	UpdateTransaction(transaction *domain.Transaction) (*domain.Transaction, error)
	FindOrderById(uId uint, id uint) (dto.SellerOrderDetails, error)
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

func (r *transactionRepository) FindTransactionByTransactionID(transactionID string) (*domain.Transaction, error) {
	var transaction domain.Transaction
	err := r.DB.Where("transaction_id = ?", transactionID).First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (r *transactionRepository) FindTransactionByPaymentID(paymentID string) (*domain.Transaction, error) {
	var transaction domain.Transaction
	err := r.DB.Where("payment_id = ?", paymentID).First(&transaction).Error
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

func (r *transactionRepository) FindOrderById(uId uint, id uint) (dto.SellerOrderDetails, error) {
	var orderItem domain.OrderItem
	var result dto.SellerOrderDetails

	err := r.DB.Where("order_id = ? AND seller_id = ?", id, uId).
		Preload("Order").
		Preload("Order.User").
		Preload("Order.User.Address").
		First(&orderItem).Error

	if err != nil {
		return result, err
	}

	// Build customer address string
	customerAddress := ""
	if orderItem.Order.User.Address.ID != 0 {
		addr := orderItem.Order.User.Address
		addressParts := []string{}
		if addr.AddressLine1 != "" {
			addressParts = append(addressParts, addr.AddressLine1)
		}
		if addr.AddressLine2 != "" {
			addressParts = append(addressParts, addr.AddressLine2)
		}
		if addr.City != "" {
			addressParts = append(addressParts, addr.City)
		}
		if addr.State != "" {
			addressParts = append(addressParts, addr.State)
		}
		if addr.Country != "" {
			addressParts = append(addressParts, addr.Country)
		}
		if addr.PostalCode != "" {
			addressParts = append(addressParts, addr.PostalCode)
		}
		customerAddress = ""
		for i, part := range addressParts {
			if i > 0 {
				customerAddress += ", "
			}
			customerAddress += part
		}
	}

	result = dto.SellerOrderDetails{
		OrderRefNumber:  orderItem.Order.OrderRefNumber,
		OrderStatus:     orderItem.Order.Status,
		CreatedAt:       orderItem.Order.CreatedAt.Format("2006-01-02 15:04:05"),
		OrderItemId:     orderItem.ID,
		ProductId:       orderItem.ProductID,
		Name:            orderItem.Name,
		ImageUrl:        orderItem.ImageUrl,
		Price:           fmt.Sprintf("%.2f", orderItem.Price),
		Qty:             orderItem.Quantity,
		CustomerName:    orderItem.Order.User.FirstName + " " + orderItem.Order.User.LastName,
		CustomerEmail:   orderItem.Order.User.Email,
		CustomerPhone:   orderItem.Order.User.Phone,
		CustomerAddress: customerAddress,
	}

	return result, nil
}
