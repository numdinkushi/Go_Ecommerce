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

func (s *TransactionService) ProcessPayment(userID uint, input dto.MakePaymentInput, calculatedAmount float64, successURL, cancelURL string) (*dto.TransactionResponse, error) {
	// Validate payment gateway
	if input.PaymentGateway == "" {
		return nil, errors.New("payment gateway is required")
	}

	// Amount is always calculated server-side from cart items for security
	// No client-provided amount is accepted or validated

	// DRY: Check existing transactions for this order
	existingTransactions, err := s.Repo.FindTransactionsByOrderID(input.OrderID)
	if err != nil {
		return nil, err
	}

	// SOLID: Prevent duplicate payments - only allow new transaction if all previous ones are failed
	for _, existing := range existingTransactions {
		// Only check transactions for the same user
		if existing.UserID != userID {
			continue
		}

		// If transaction is succeeded, prevent new payment
		if existing.Status == "succeeded" {
			return nil, errors.New("payment already completed for this order")
		}

		// If transaction is pending, check with provider to verify current status
		if existing.Status == "pending" {
			provider, providerErr := payment.GetProvider(existing.PaymentGateway)
			if providerErr == nil && existing.TransactionID != "" {
				statusResp, statusErr := provider.GetPaymentStatus(existing.TransactionID)
				if statusErr == nil {
					// Status changed on provider - update local transaction
					if statusResp.Status != existing.Status {
						log.Printf("🔄 Updating existing transaction %d status from provider: %s → %s",
							existing.ID, existing.Status, statusResp.Status)
						s.updateTransactionFromProviderStatus(&existing, statusResp)
						// Re-fetch to get updated status
						updated, fetchErr := s.Repo.FindTransactionByID(existing.ID)
						if fetchErr == nil && updated != nil {
							existing = *updated
						}
					}

					// If still pending after update, return existing transaction
					if statusResp.Status == "pending" {
						log.Printf("♻️  Reusing existing pending transaction - ID: %d, TransactionID: %s",
							existing.ID, existing.TransactionID)
						response := s.toTransactionResponse(&existing)
						if existing.CheckoutURL != "" {
							response.CheckoutURL = existing.CheckoutURL
						}
						return response, nil
					}

					// If now succeeded, prevent new payment
					if statusResp.Status == "succeeded" {
						return nil, errors.New("payment already completed for this order")
					}
					// If failed, allow new payment attempt
					if statusResp.Status == "failed" {
						continue
					}
				}
			}

			// If still pending after provider check (or provider check failed), return existing
			log.Printf("♻️  Reusing existing pending transaction - ID: %d, TransactionID: %s",
				existing.ID, existing.TransactionID)
			response := s.toTransactionResponse(&existing)
			if existing.CheckoutURL != "" {
				response.CheckoutURL = existing.CheckoutURL
			}
			return response, nil
		}

		// If not failed and not succeeded and not pending, prevent new payment
		if existing.Status != "failed" {
			return nil, errors.New("cannot create new payment: existing transaction status is '" + existing.Status + "'. Only failed transactions allow new payment attempts")
		}
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

	// Create payment session using provider
	// Use calculatedAmount (server-side) instead of client-provided amount for security
	// URLs are provided by frontend for flexibility
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
		log.Printf("❌ Failed to create payment session: %v", err)
		return nil, err
	}

	log.Printf("✅ Payment session created - SessionID: %s, PaymentID: %s, CheckoutURL: %s",
		sessionResp.SessionID, sessionResp.PaymentID, sessionResp.CheckoutURL)

	// Note: PaymentID may be empty initially - Stripe creates PaymentIntent when user starts checkout
	// The webhook will contain the PaymentID and we'll update the transaction then

	// Create transaction record
	// Use calculatedAmount (server-side) instead of client-provided amount for security
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

	log.Printf("💾 Creating transaction record - UserID: %d, OrderID: %d, TransactionID: %s, PaymentID: %s",
		userID, input.OrderID, transaction.TransactionID, transaction.PaymentID)

	createdTransaction, err := s.Repo.CreateTransaction(transaction)
	if err != nil {
		log.Printf("❌ Failed to create transaction in database: %v", err)
		return nil, err
	}

	log.Printf("✅ Transaction created successfully - ID: %d, OrderID: %d, TransactionID: %s, PaymentID: %s",
		createdTransaction.ID, createdTransaction.OrderID, createdTransaction.TransactionID, createdTransaction.PaymentID)
	log.Printf("📝 IMPORTANT: Webhook must match by PaymentID (%s) or TransactionID (%s)",
		createdTransaction.PaymentID, createdTransaction.TransactionID)

	// Start background polling for payment status (retry logic)
	go s.startPaymentStatusPolling(createdTransaction.ID, input.PaymentGateway)

	// Transform to DTO (checkout URL is already stored in transaction)
	response := s.toTransactionResponse(createdTransaction)

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
// DRY: Uses updateTransactionFromProviderStatus for consistency
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

	// DRY: Use shared update method
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
		log.Printf("✅ Order %d updated - Status: 'confirmed', PaymentID: %s, TransactionID: %s",
			order.ID, updatedOrder.PaymentId, updatedOrder.TransactionId)
	} else {
		log.Printf("✅ Order %d payment details updated - PaymentID: %s, TransactionID: %s",
			order.ID, updatedOrder.PaymentId, updatedOrder.TransactionId)
	}

	return nil
}

// updateOrderStatusIfTransactionSucceeded checks if any transaction for an order has succeeded
// and updates the order status to "confirmed" if so (used by VerifyPaymentStatus)
func (s *TransactionService) updateOrderStatusIfTransactionSucceeded(orderID uint) error {
	// Get all transactions for this order
	transactions, err := s.Repo.FindTransactionsByOrderID(orderID)
	if err != nil {
		return err
	}

	// Check if any transaction has succeeded
	hasSucceeded := false
	for _, txn := range transactions {
		if txn.Status == "succeeded" {
			hasSucceeded = true
			break
		}
	}

	// If no transaction succeeded, no need to update order
	if !hasSucceeded {
		return nil
	}

	// Get the order
	order, err := s.OrderRepo.FindOrderByID(orderID)
	if err != nil {
		return err
	}

	// Only update if order is still pending
	if order.Status != "pending" {
		return nil
	}

	// Update order status to confirmed
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

	// BEST PRACTICE: Match transactions using priority order:
	// 1. PaymentIntent ID (for payment_intent.succeeded, charge.succeeded)
	// 2. SessionID (for checkout.session.completed)
	// 3. Metadata order_id (ultimate fallback for all event types)
	var transaction *domain.Transaction
	isTestWebhook := false

	// Step 1: Try matching by PaymentIntent ID first (most direct for payment_intent events)
	if webhookEvent.PaymentID != "" {
		log.Printf("🔍 [Step 1/3] Searching for transaction by PaymentID: %s", webhookEvent.PaymentID)
		transaction, err = s.Repo.FindTransactionByPaymentID(webhookEvent.PaymentID)
		if err != nil {
			transaction = nil
			log.Printf("   ❌ Not found by PaymentID")
		} else {
			log.Printf("   ✅ Found transaction by PaymentID: transaction_id=%d", transaction.ID)
		}
	}

	// Step 2: Try matching by SessionID (for checkout.session.completed events)
	if transaction == nil && webhookEvent.SessionID != "" {
		log.Printf("🔍 [Step 2/3] Searching for transaction by SessionID: %s", webhookEvent.SessionID)
		transaction, err = s.Repo.FindTransactionByTransactionID(webhookEvent.SessionID)
		if err != nil {
			transaction = nil
			log.Printf("   ❌ Not found by SessionID")
		} else {
			log.Printf("   ✅ Found transaction by SessionID: transaction_id=%d", transaction.ID)
		}
	}

	// Step 3: If metadata is empty, try to fetch from Stripe
	// This is critical for payment_intent.succeeded events which may have empty metadata initially
	if transaction == nil && len(webhookEvent.Metadata) == 0 && gateway == "stripe" {
		if webhookEvent.PaymentID != "" {
			log.Printf("⚠️  Metadata empty, fetching payment intent from Stripe: %s", webhookEvent.PaymentID)
			if stripeProvider, ok := provider.(interface {
				GetPaymentIntentMetadata(paymentIntentID string) (map[string]string, error)
			}); ok {
				metadata, metaErr := stripeProvider.GetPaymentIntentMetadata(webhookEvent.PaymentID)
				if metaErr == nil && len(metadata) > 0 {
					webhookEvent.Metadata = metadata
					log.Printf("✅ Fetched metadata from Stripe payment intent: %v", metadata)
				} else if metaErr != nil {
					// Payment intent doesn't exist (404) - likely a test webhook with dummy ID
					log.Printf("⚠️  Payment intent not found (likely test webhook): %v", metaErr)
					isTestWebhook = true
				}
			}
		}
		// If still no metadata, try fetching by SessionID
		if len(webhookEvent.Metadata) == 0 && webhookEvent.SessionID != "" {
			log.Printf("⚠️  Metadata still empty, fetching session from Stripe: %s", webhookEvent.SessionID)
			statusResp, statusErr := provider.GetPaymentStatus(webhookEvent.SessionID)
			if statusErr == nil && statusResp.Metadata != nil {
				webhookEvent.Metadata = statusResp.Metadata
				log.Printf("✅ Fetched metadata from Stripe session: %v", statusResp.Metadata)
			}
		}
	}

	// Step 4: Try matching by metadata order_id (ultimate fallback)
	// This ensures we can match even if IDs don't match (due to edge cases or data inconsistencies)
	if transaction == nil {
		orderIDStr := webhookEvent.Metadata["order_id"]
		if orderIDStr != "" {
			orderID, parseErr := strconv.ParseUint(orderIDStr, 10, 32)
			if parseErr == nil {
				log.Printf("🔍 [Step 3/3] Searching for transaction by order_id from metadata: %d", orderID)
				transactions, findErr := s.Repo.FindTransactionsByOrderID(uint(orderID))
				if findErr == nil && len(transactions) > 0 {
					// Prefer pending transactions first
					for i := range transactions {
						if transactions[i].Status == "pending" {
							transaction = &transactions[i]
							log.Printf("✅ Found pending transaction by order_id: order_id=%d, transaction_id=%d, TransactionID=%s",
								orderID, transaction.ID, transaction.TransactionID)
							break
						}
					}
					// If no pending found, use the most recent one
					if transaction == nil {
						transaction = &transactions[0]
						log.Printf("✅ Found transaction by order_id (not pending): order_id=%d, transaction_id=%d, TransactionID=%s",
							orderID, transaction.ID, transaction.TransactionID)
					}
				} else {
					log.Printf("   ❌ No transactions found for order_id=%d", orderID)
				}
			}
		}
	}

	if transaction == nil {
		if isTestWebhook {
			log.Printf("⚠️  Test webhook ignored (PaymentID not found in Stripe): %s", webhookEvent.PaymentID)
			log.Printf("   EventType: %s", webhookEvent.EventType)
			// Return success to prevent Stripe from retrying test webhooks
			return nil
		}

		log.Printf("❌ Transaction not found for webhook")
		log.Printf("   EventType: %s", webhookEvent.EventType)
		log.Printf("   PaymentID: %s", webhookEvent.PaymentID)
		log.Printf("   SessionID: %s", webhookEvent.SessionID)
		log.Printf("   Metadata: %v", webhookEvent.Metadata)
		log.Printf("   Note: This webhook cannot be matched to any transaction in the database")

		// Return error but don't do expensive debugging - just log and return
		return errors.New("transaction not found - webhook cannot be matched to any transaction")
	}

	// IDEMPOTENCY CHECK: If transaction is already succeeded, skip processing
	// This prevents duplicate processing if webhook is received multiple times
	if transaction.Status == "succeeded" && webhookEvent.Status == "succeeded" {
		log.Printf("✅ Transaction already succeeded (idempotency check) - ID: %d, Status: %s", transaction.ID, transaction.Status)
		log.Printf("   Webhook event type: %s - skipping duplicate processing", webhookEvent.EventType)
		return nil
	}

	// DRY: Use shared update method - convert webhook event to provider status response format
	statusResp := &payment.PaymentStatusResponse{
		Status:    webhookEvent.Status,
		PaymentID: webhookEvent.PaymentID,
		Amount:    webhookEvent.Amount,
		Currency:  webhookEvent.Currency,
		Metadata:  webhookEvent.Metadata,
	}

	// Update TransactionID if webhook provides it and transaction doesn't have it
	if webhookEvent.SessionID != "" && transaction.TransactionID == "" {
		transaction.TransactionID = webhookEvent.SessionID
		log.Printf("💾 Updating transaction TransactionID to: %s", webhookEvent.SessionID)
		s.Repo.UpdateTransaction(transaction)
	}

	// DRY: Use shared update method for status and PaymentID
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
			// Check transaction status
			transaction, err := s.Repo.FindTransactionByID(transactionID)
			if err != nil {
				pollMutex.Unlock()
				log.Printf("⚠️  Polling: Transaction %d not found, stopping poll", transactionID)
				return
			}

			// If already succeeded or failed, stop polling
			if transaction.Status == "succeeded" || transaction.Status == "failed" {
				pollMutex.Unlock()
				log.Printf("✅ Polling: Transaction %d status is %s, stopping poll", transactionID, transaction.Status)
				return
			}

			// Check status with provider (provider-agnostic: works with Stripe, Flutterwave, etc.)
			provider, err := payment.GetProvider(gateway)
			if err != nil {
				pollMutex.Unlock()
				log.Printf("⚠️  Polling: Provider %s not available for transaction %d", gateway, transactionID)
				continue
			}

			statusResp, err := provider.GetPaymentStatus(transaction.TransactionID)
			if err != nil {
				pollMutex.Unlock()
				log.Printf("⚠️  Polling: Failed to check status for transaction %d: %v", transactionID, err)
				continue
			}

			// Update if status changed
			if statusResp.Status != transaction.Status {
				log.Printf("🔄 Polling: Transaction %d status changed: %s → %s", transactionID, transaction.Status, statusResp.Status)
				s.updateTransactionFromProviderStatus(transaction, statusResp)
			}
			pollMutex.Unlock()

			// If succeeded or failed, stop polling
			if statusResp.Status == "succeeded" || statusResp.Status == "failed" {
				return
			}
		}

		log.Printf("⏰ Polling: Reached max attempts for transaction %d, stopping background poll", transactionID)
	}()
}

// updateTransactionFromProviderStatus updates transaction and order status from provider response
// DRY: Reusable method for updating transaction status (used by polling, verification, and webhooks)
// SOLID: Single responsibility - only updates status, doesn't handle provider-specific logic
func (s *TransactionService) updateTransactionFromProviderStatus(transaction *domain.Transaction, statusResp *payment.PaymentStatusResponse) {
	oldStatus := transaction.Status
	transaction.Status = statusResp.Status

	// Update PaymentID if provider provides it and transaction doesn't have it
	if statusResp.PaymentID != "" && transaction.PaymentID == "" {
		transaction.PaymentID = statusResp.PaymentID
		log.Printf("💾 Updating transaction PaymentID to: %s", statusResp.PaymentID)
	}

	// Update transaction in database
	updatedTransaction, err := s.Repo.UpdateTransaction(transaction)
	if err != nil {
		log.Printf("❌ Failed to update transaction %d: %v", transaction.ID, err)
		return
	}

	if oldStatus != updatedTransaction.Status {
		log.Printf("✅ Transaction updated via polling/verification - ID: %d, Status: %s → %s",
			updatedTransaction.ID, oldStatus, updatedTransaction.Status)
	}

	// Update order status if transaction succeeded
	if updatedTransaction.Status == "succeeded" {
		if err := s.updateOrderWithPaymentDetails(updatedTransaction); err != nil {
			log.Printf("⚠️  Failed to update order after transaction status update: %v", err)
		}
	}
}
