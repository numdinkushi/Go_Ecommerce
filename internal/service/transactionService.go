package service

import (
	"errors"
	"go-ecommerce-app/config"
	"go-ecommerce-app/internal/domain"
	"go-ecommerce-app/internal/dto"
	"go-ecommerce-app/internal/helper"
	"go-ecommerce-app/internal/repository"
	"go-ecommerce-app/pkg/payment"
	"log"
	"strconv"
	"sync"
	"time"
)

type TransactionService struct {
	Repo      repository.TransactionRepository
	OrderRepo repository.UserRepository
	Auth      helper.Auth
	Config    config.AppConfig
}

func NewTransactionService(repo repository.TransactionRepository, orderRepo repository.UserRepository, auth helper.Auth, config config.AppConfig) *TransactionService {
	return &TransactionService{
		Repo:      repo,
		OrderRepo: orderRepo,
		Auth:      auth,
		Config:    config,
	}
}

func (s *TransactionService) CreateTransaction(userID uint, input dto.CreateTransactionInput) (*dto.TransactionResponse, error) {
	currency := input.Currency
	if currency == "" {
		currency = "NGN"
	}

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

	return s.toTransactionResponse(createdTransaction), nil
}

func (s *TransactionService) ProcessPayment(userID uint, input dto.MakePaymentInput, calculatedAmount float64, successURL, cancelURL string) (*dto.TransactionResponse, error) {
	if input.PaymentGateway == "" {
		return nil, errors.New("payment gateway is required")
	}

	existingTransactions, err := s.Repo.FindTransactionsByOrderID(input.OrderID)
	if err != nil {
		return nil, err
	}

	for _, existing := range existingTransactions {
		if existing.UserID != userID {
			continue
		}

		if existing.Status == "succeeded" {
			return nil, errors.New("payment already completed for this order")
		}

		if existing.Status == "pending" {
			provider, providerErr := payment.GetProvider(existing.PaymentGateway)
			if providerErr == nil && existing.TransactionID != "" {
				statusResp, statusErr := provider.GetPaymentStatus(existing.TransactionID)
				if statusErr == nil {
					if statusResp.Status != existing.Status {
						s.updateTransactionFromProviderStatus(&existing, statusResp)
						updated, fetchErr := s.Repo.FindTransactionByID(existing.ID)
						if fetchErr == nil && updated != nil {
							existing = *updated
						}
					}

					if statusResp.Status == "pending" {
						response := s.toTransactionResponse(&existing)
						if existing.CheckoutURL != "" {
							response.CheckoutURL = existing.CheckoutURL
						}
						return response, nil
					}

					if statusResp.Status == "succeeded" {
						return nil, errors.New("payment already completed for this order")
					}
					if statusResp.Status == "failed" {
						continue
					}
				}
			}

			response := s.toTransactionResponse(&existing)
			if existing.CheckoutURL != "" {
				response.CheckoutURL = existing.CheckoutURL
			}
			return response, nil
		}

		if existing.Status != "failed" {
			return nil, errors.New("cannot create new payment: existing transaction status is '" + existing.Status + "'. Only failed transactions allow new payment attempts")
		}
	}

	provider, err := payment.GetProvider(input.PaymentGateway)
	if err != nil {
		return nil, errors.New("payment provider not available: " + input.PaymentGateway)
	}

	currency := input.Currency
	if currency == "" {
		currency = "NGN"
	}

	sessionReq := payment.CreatePaymentSessionRequest{
		Amount:     calculatedAmount,
		Currency:   currency,
		UserID:     userID,
		OrderID:    input.OrderID,
		SuccessURL: successURL,
		CancelURL:  cancelURL,
		Metadata:   make(map[string]string),
	}

	sessionResp, err := provider.CreatePaymentSession(sessionReq)
	if err != nil {
		log.Printf("Failed to create payment session: %v", err)
		return nil, err
	}

	transaction := &domain.Transaction{
		UserID:         userID,
		OrderID:        input.OrderID,
		Amount:         calculatedAmount,
		Status:         "pending",
		PaymentMethod:  input.PaymentMethod,
		PaymentGateway: input.PaymentGateway,
		TransactionID:  sessionResp.SessionID,
		PaymentID:      sessionResp.PaymentID,
		CheckoutURL:    sessionResp.CheckoutURL,
		Currency:       currency,
	}

	createdTransaction, err := s.Repo.CreateTransaction(transaction)
	if err != nil {
		log.Printf("Failed to create transaction in database: %v", err)
		return nil, err
	}

	go s.startPaymentStatusPolling(createdTransaction.ID, input.PaymentGateway)

	response := s.toTransactionResponse(createdTransaction)

	return response, nil
}

func (s *TransactionService) GetTransactionsByUserID(userID uint) ([]dto.TransactionResponse, error) {
	transactions, err := s.Repo.FindTransactionsByUserID(userID)
	if err != nil {
		return nil, err
	}

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

	if transaction.UserID != userID {
		return nil, errors.New("transaction not found")
	}

	return s.toTransactionResponse(transaction), nil
}

// VerifyPaymentStatus verifies the payment status with the payment provider
func (s *TransactionService) VerifyPaymentStatus(transactionID uint, gateway string) (*dto.TransactionResponse, error) {
	provider, err := payment.GetProvider(gateway)
	if err != nil {
		return nil, errors.New("payment provider not available: " + gateway)
	}

	transaction, err := s.Repo.FindTransactionByID(transactionID)
	if err != nil {
		return nil, err
	}

	statusResp, err := provider.GetPaymentStatus(transaction.TransactionID)
	if err != nil {
		return nil, err
	}

	s.updateTransactionFromProviderStatus(transaction, statusResp)

	// Fetch updated transaction
	updatedTransaction, err := s.Repo.FindTransactionByID(transactionID)
	if err != nil {
		return nil, err
	}

	return s.toTransactionResponse(updatedTransaction), nil
}

// updateOrderWithPaymentDetails updates the order with payment details and status
// when a transaction succeeds
func (s *TransactionService) updateOrderWithPaymentDetails(transaction *domain.Transaction) error {
	order, err := s.OrderRepo.FindOrderByID(transaction.OrderID)
	if err != nil {
		return err
	}

	// Update payment details if transaction succeeded
	if transaction.Status == "succeeded" {
		// Update order payment_id and transaction_id with actual values from transaction
		if transaction.PaymentID != "" {
			order.PaymentId = transaction.PaymentID
		}
		if transaction.TransactionID != "" {
			order.TransactionId = transaction.TransactionID
		}

		// Update order status if still pending
		if order.Status == "pending" {
			order.Status = "confirmed"
		}
	} else {
		// For non-succeeded transactions, still update payment details if available
		// (in case payment_id was set but status is still pending)
		if transaction.PaymentID != "" && order.PaymentId == "" {
			order.PaymentId = transaction.PaymentID
		}
		if transaction.TransactionID != "" && order.TransactionId == "" {
			order.TransactionId = transaction.TransactionID
		}
	}

	updatedOrder, err := s.OrderRepo.UpdateOrder(order)
	if err != nil {
		return err
	}

	if transaction.Status == "succeeded" && order.Status == "confirmed" {
		log.Printf("Order %d updated - Status: confirmed, PaymentID: %s, TransactionID: %s",
			order.ID, updatedOrder.PaymentId, updatedOrder.TransactionId)
	}

	return nil
}

// updateOrderStatusIfTransactionSucceeded checks if any transaction for an order has succeeded
// and updates the order status to "confirmed" if so (used by VerifyPaymentStatus)
func (s *TransactionService) updateOrderStatusIfTransactionSucceeded(orderID uint) error {
	transactions, err := s.Repo.FindTransactionsByOrderID(orderID)
	if err != nil {
		return err
	}

	hasSucceeded := false
	for _, txn := range transactions {
		if txn.Status == "succeeded" {
			hasSucceeded = true
			break
		}
	}

	if !hasSucceeded {
		return nil
	}

	order, err := s.OrderRepo.FindOrderByID(orderID)
	if err != nil {
		return err
	}

	if order.Status != "pending" {
		return nil
	}

	order.Status = "confirmed"
	updatedOrder, err := s.OrderRepo.UpdateOrder(order)
	if err != nil {
		return err
	}

	log.Printf("Order %d status updated from 'pending' to 'confirmed' after successful transaction", orderID)
	_ = updatedOrder // Acknowledge but don't use

	return nil
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

	// Match transactions using priority: PaymentID, SessionID, then metadata order_id
	var transaction *domain.Transaction
	isTestWebhook := false

	if webhookEvent.PaymentID != "" {
		transaction, err = s.Repo.FindTransactionByPaymentID(webhookEvent.PaymentID)
		if err != nil {
			transaction = nil
		}
	}

	if transaction == nil && webhookEvent.SessionID != "" {
		transaction, err = s.Repo.FindTransactionByTransactionID(webhookEvent.SessionID)
		if err != nil {
			transaction = nil
		}
	}

	// Fetch metadata from Stripe if empty (payment_intent.succeeded events may have empty metadata initially)
	if transaction == nil && len(webhookEvent.Metadata) == 0 && gateway == "stripe" {
		if webhookEvent.PaymentID != "" {
			if stripeProvider, ok := provider.(interface {
				GetPaymentIntentMetadata(paymentIntentID string) (map[string]string, error)
			}); ok {
				metadata, metaErr := stripeProvider.GetPaymentIntentMetadata(webhookEvent.PaymentID)
				if metaErr == nil && len(metadata) > 0 {
					webhookEvent.Metadata = metadata
				} else if metaErr != nil {
					isTestWebhook = true
				}
			}
		}
		if len(webhookEvent.Metadata) == 0 && webhookEvent.SessionID != "" {
			statusResp, statusErr := provider.GetPaymentStatus(webhookEvent.SessionID)
			if statusErr == nil && statusResp.Metadata != nil {
				webhookEvent.Metadata = statusResp.Metadata
			}
		}
	}

	// Fallback: match by metadata order_id
	if transaction == nil {
		orderIDStr := webhookEvent.Metadata["order_id"]
		if orderIDStr != "" {
			orderID, parseErr := strconv.ParseUint(orderIDStr, 10, 32)
			if parseErr == nil {
				transactions, findErr := s.Repo.FindTransactionsByOrderID(uint(orderID))
				if findErr == nil && len(transactions) > 0 {
					for i := range transactions {
						if transactions[i].Status == "pending" {
							transaction = &transactions[i]
							break
						}
					}
					if transaction == nil {
						transaction = &transactions[0]
					}
				}
			}
		}
	}

	if transaction == nil {
		if isTestWebhook {
			return nil
		}

		log.Printf("Transaction not found for webhook - EventType: %s, PaymentID: %s, SessionID: %s",
			webhookEvent.EventType, webhookEvent.PaymentID, webhookEvent.SessionID)
		return errors.New("transaction not found - webhook cannot be matched to any transaction")
	}

	// Skip processing if transaction already succeeded (idempotency check)
	if transaction.Status == "succeeded" && webhookEvent.Status == "succeeded" {
		return nil
	}

	statusResp := &payment.PaymentStatusResponse{
		Status:    webhookEvent.Status,
		PaymentID: webhookEvent.PaymentID,
		Amount:    webhookEvent.Amount,
		Currency:  webhookEvent.Currency,
		Metadata:  webhookEvent.Metadata,
	}

	if webhookEvent.SessionID != "" && transaction.TransactionID == "" {
		transaction.TransactionID = webhookEvent.SessionID
		s.Repo.UpdateTransaction(transaction)
	}

	s.updateTransactionFromProviderStatus(transaction, statusResp)

	return nil
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
		CheckoutURL:    transaction.CheckoutURL,
		Currency:       transaction.Currency,
		CreatedAt:      transaction.CreatedAt,
		UpdatedAt:      transaction.UpdatedAt,
	}
}

// PaymentStatusPollingManager manages background polling for payment status
// Prevents duplicate polling for the same transaction
type PaymentStatusPollingManager struct {
	activePolls map[uint]*sync.Mutex
	mu          sync.RWMutex
}

var pollingManager = &PaymentStatusPollingManager{
	activePolls: make(map[uint]*sync.Mutex),
}

// startPaymentStatusPolling starts background polling for a transaction's payment status
// This ensures we update transaction/order status even if webhooks fail
// Polls at intervals: 10s, 20s, 30s, 60s, 120s (total ~5 minutes max)
func (s *TransactionService) startPaymentStatusPolling(transactionID uint, gateway string) {
	pollingManager.mu.Lock()
	if _, exists := pollingManager.activePolls[transactionID]; exists {
		// Polling already running for this transaction
		pollingManager.mu.Unlock()
		return
	}
	pollMutex := &sync.Mutex{}
	pollingManager.activePolls[transactionID] = pollMutex
	pollingManager.mu.Unlock()

	go func() {
		defer func() {
			pollingManager.mu.Lock()
			delete(pollingManager.activePolls, transactionID)
			pollingManager.mu.Unlock()
		}()

		// Polling intervals: progressively longer delays
		intervals := []time.Duration{10 * time.Second, 20 * time.Second, 30 * time.Second, 60 * time.Second, 120 * time.Second}
		maxAttempts := len(intervals)

		for attempt := 0; attempt < maxAttempts; attempt++ {
			time.Sleep(intervals[attempt])

			pollMutex.Lock()
			transaction, err := s.Repo.FindTransactionByID(transactionID)
			if err != nil {
				pollMutex.Unlock()
				return
			}

			if transaction.Status == "succeeded" || transaction.Status == "failed" {
				pollMutex.Unlock()
				return
			}

			provider, err := payment.GetProvider(gateway)
			if err != nil {
				pollMutex.Unlock()
				continue
			}

			statusResp, err := provider.GetPaymentStatus(transaction.TransactionID)
			if err != nil {
				pollMutex.Unlock()
				continue
			}

			if statusResp.Status != transaction.Status {
				log.Printf("Polling: Transaction %d status changed: %s → %s", transactionID, transaction.Status, statusResp.Status)
				s.updateTransactionFromProviderStatus(transaction, statusResp)
			}
			pollMutex.Unlock()

			if statusResp.Status == "succeeded" || statusResp.Status == "failed" {
				return
			}
		}

		log.Printf("Polling: Reached max attempts for transaction %d", transactionID)
	}()
}

// updateTransactionFromProviderStatus updates transaction and order status from provider response
func (s *TransactionService) updateTransactionFromProviderStatus(transaction *domain.Transaction, statusResp *payment.PaymentStatusResponse) {
	oldStatus := transaction.Status
	transaction.Status = statusResp.Status

	if statusResp.PaymentID != "" && transaction.PaymentID == "" {
		transaction.PaymentID = statusResp.PaymentID
	}

	updatedTransaction, err := s.Repo.UpdateTransaction(transaction)
	if err != nil {
		log.Printf("Failed to update transaction %d: %v", transaction.ID, err)
		return
	}

	if oldStatus != updatedTransaction.Status {
		log.Printf("Transaction updated - ID: %d, Status: %s → %s", updatedTransaction.ID, oldStatus, updatedTransaction.Status)
	}

	if updatedTransaction.Status == "succeeded" {
		if err := s.updateOrderWithPaymentDetails(updatedTransaction); err != nil {
			log.Printf("Failed to update order after transaction status update: %v", err)
		}
	}
}
