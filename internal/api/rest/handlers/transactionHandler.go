package handlers

import (
	"fmt"

	"go-ecommerce-app/internal/api/rest"
	"go-ecommerce-app/internal/domain"
	"go-ecommerce-app/internal/dto"
	"go-ecommerce-app/internal/helper"
	"go-ecommerce-app/internal/repository"
	"go-ecommerce-app/internal/service"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type TransactionHandler struct {
	transactionService service.TransactionService
	auth               helper.Auth
	userService        service.UserService
}

func initializeTransactionService(db *gorm.DB, auth helper.Auth) service.TransactionService {
	transactionRepo := repository.NewTransactionRepository(db)
	return service.NewTransactionService(transactionRepo, auth)
}

func SetupTransactionRoutes(restHandler *rest.RestHandler, bankService *service.BankService) {
	app := restHandler.App

	transactionService := initializeTransactionService(restHandler.DB, restHandler.Auth)
	userRepo := repository.NewUserRepository(restHandler.DB)
	catalogueRepo := repository.NewCatalogueRepository(restHandler.DB)
	userService := service.NewUserService(userRepo, catalogueRepo, restHandler.Auth, restHandler.Config, bankService)

	handler := TransactionHandler{
		transactionService: transactionService,
		auth:               restHandler.Auth,
		userService:        userService,
	}

	// General authenticated routes
	secRoute := app.Group("/", restHandler.Auth.Authorize)
	secRoute.Get("/payment", handler.MakePayment)

	// Seller-specific authenticated routes
	sellerRoute := app.Group("/seller", restHandler.Auth.AuthorizeSeller(userRepo))
	sellerRoute.Get("/orders", handler.GetOrders)
	sellerRoute.Get("/orders/:id", handler.GetOrderDetails)

	// Private endpoints (authentication required)
	privateRoutes := app.Group("/", restHandler.Auth.Authorize)
	privateRoutes.Post("/transactions", handler.CreateTransaction)
	privateRoutes.Get("/transactions", handler.GetTransactions)
	privateRoutes.Get("/transactions/:id", handler.GetTransactionByID)
}

func NewTransactionHandler(transactionService service.TransactionService, auth helper.Auth) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
		auth:               auth,
	}
}

func (h *TransactionHandler) CreateTransaction(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)

	var transactionInput dto.CreateTransactionInput
	if err := ctx.BodyParser(&transactionInput); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	transaction := &domain.Transaction{
		UserID:         user.ID,
		OrderID:        transactionInput.OrderID,
		Amount:         transactionInput.Amount,
		Status:         "pending",
		PaymentMethod:  transactionInput.PaymentMethod,
		PaymentGateway: transactionInput.PaymentGateway,
		TransactionID:  transactionInput.TransactionID,
		PaymentID:      transactionInput.PaymentID,
		Currency:       transactionInput.Currency,
	}

	if transaction.Currency == "" {
		transaction.Currency = "NGN"
	}

	createdTransaction, err := h.transactionService.CreateTransaction(transaction)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	transactionResponse := fiber.Map{
		"id":              createdTransaction.ID,
		"user_id":         createdTransaction.UserID,
		"order_id":        createdTransaction.OrderID,
		"amount":          createdTransaction.Amount,
		"status":          createdTransaction.Status,
		"payment_method":  createdTransaction.PaymentMethod,
		"payment_gateway": createdTransaction.PaymentGateway,
		"transaction_id":  createdTransaction.TransactionID,
		"payment_id":      createdTransaction.PaymentID,
		"currency":        createdTransaction.Currency,
		"created_at":      createdTransaction.CreatedAt,
		"updated_at":      createdTransaction.UpdatedAt,
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Transaction created successfully",
		"transaction": transactionResponse,
	})
}

func (h *TransactionHandler) GetTransactions(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)

	transactions, err := h.transactionService.GetTransactionsByUserID(user.ID)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	transactionsResponse := make([]fiber.Map, len(transactions))
	for i, transaction := range transactions {
		transactionsResponse[i] = fiber.Map{
			"id":              transaction.ID,
			"user_id":         transaction.UserID,
			"order_id":        transaction.OrderID,
			"amount":          transaction.Amount,
			"status":          transaction.Status,
			"payment_method":  transaction.PaymentMethod,
			"payment_gateway": transaction.PaymentGateway,
			"transaction_id":  transaction.TransactionID,
			"payment_id":      transaction.PaymentID,
			"currency":        transaction.Currency,
			"created_at":      transaction.CreatedAt,
			"updated_at":      transaction.UpdatedAt,
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "Transactions retrieved successfully",
		"transactions": transactionsResponse,
		"count":        len(transactions),
	})
}

func (h *TransactionHandler) GetTransactionByID(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)
	transactionID := ctx.Params("id")

	var id uint
	if _, err := fmt.Sscanf(transactionID, "%d", &id); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid transaction ID",
		})
	}

	transaction, err := h.transactionService.GetTransactionByID(id)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	// Verify that the transaction belongs to the user
	if transaction.UserID != user.ID {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "You don't have access to this transaction",
		})
	}

	transactionResponse := fiber.Map{
		"id":              transaction.ID,
		"user_id":         transaction.UserID,
		"order_id":        transaction.OrderID,
		"amount":          transaction.Amount,
		"status":          transaction.Status,
		"payment_method":  transaction.PaymentMethod,
		"payment_gateway": transaction.PaymentGateway,
		"transaction_id":  transaction.TransactionID,
		"payment_id":      transaction.PaymentID,
		"currency":        transaction.Currency,
		"created_at":      transaction.CreatedAt,
		"updated_at":      transaction.UpdatedAt,
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":     "Transaction retrieved successfully",
		"transaction": transactionResponse,
	})
}

func (h *TransactionHandler) MakePayment(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)

	var paymentInput dto.MakePaymentInput
	if err := ctx.BodyParser(&paymentInput); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Create transaction for payment
	transaction := &domain.Transaction{
		UserID:         user.ID,
		OrderID:        paymentInput.OrderID,
		Amount:         paymentInput.Amount,
		Status:         "pending",
		PaymentMethod:  paymentInput.PaymentMethod,
		PaymentGateway: paymentInput.PaymentGateway,
		TransactionID:  paymentInput.TransactionID,
		PaymentID:      paymentInput.PaymentID,
		Currency:       paymentInput.Currency,
	}

	if transaction.Currency == "" {
		transaction.Currency = "NGN"
	}

	createdTransaction, err := h.transactionService.CreateTransaction(transaction)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	paymentResponse := fiber.Map{
		"id":              createdTransaction.ID,
		"user_id":         createdTransaction.UserID,
		"order_id":        createdTransaction.OrderID,
		"amount":          createdTransaction.Amount,
		"status":          createdTransaction.Status,
		"payment_method":  createdTransaction.PaymentMethod,
		"payment_gateway": createdTransaction.PaymentGateway,
		"transaction_id":  createdTransaction.TransactionID,
		"payment_id":      createdTransaction.PaymentID,
		"currency":        createdTransaction.Currency,
		"created_at":      createdTransaction.CreatedAt,
		"updated_at":      createdTransaction.UpdatedAt,
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Payment processed successfully",
		"payment": paymentResponse,
	})
}

func (h *TransactionHandler) GetOrders(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)

	orders, err := h.userService.GetOrdersBySellerID(user.ID)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	ordersResponse := make([]fiber.Map, len(orders))
	for i, order := range orders {
		// Filter items to only show items belonging to this seller
		sellerItems := make([]fiber.Map, 0)
		for _, item := range order.Items {
			if item.SellerId == user.ID {
				sellerItems = append(sellerItems, fiber.Map{
					"id":         item.ID,
					"product_id": item.ProductID,
					"name":       item.Name,
					"image_url":  item.ImageUrl,
					"seller_id":  item.SellerId,
					"price":      item.Price,
					"quantity":   item.Quantity,
					"created_at": item.CreatedAt,
					"updated_at": item.UpdatedAt,
				})
			}
		}

		ordersResponse[i] = fiber.Map{
			"id":               order.ID,
			"user_id":          order.UserID,
			"status":           order.Status,
			"amount":           order.Amount,
			"transaction_id":   order.TransactionId,
			"order_ref_number": order.OrderRefNumber,
			"payment_id":       order.PaymentId,
			"items":            sellerItems,
			"created_at":       order.CreatedAt,
			"updated_at":       order.UpdatedAt,
		}
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Orders retrieved successfully",
		"orders":  ordersResponse,
		"count":   len(orders),
	})
}

func (h *TransactionHandler) GetOrderDetails(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)
	orderID, err := ctx.ParamsInt("id")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid order ID",
		})
	}

	order, err := h.userService.GetOrderDetailsBySellerID(uint(orderID), user.ID)
	if err != nil {
		if err.Error() == "order not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Order not found",
			})
		}
		return helper.HandleDBError(ctx, err)
	}

	// Filter items to only show items belonging to this seller
	sellerItems := make([]fiber.Map, 0)
	for _, item := range order.Items {
		if item.SellerId == user.ID {
			sellerItems = append(sellerItems, fiber.Map{
				"id":         item.ID,
				"product_id": item.ProductID,
				"name":       item.Name,
				"image_url":  item.ImageUrl,
				"seller_id":  item.SellerId,
				"price":      item.Price,
				"quantity":   item.Quantity,
				"created_at": item.CreatedAt,
				"updated_at": item.UpdatedAt,
			})
		}
	}

	orderResponse := fiber.Map{
		"id":               order.ID,
		"user_id":          order.UserID,
		"status":           order.Status,
		"amount":           order.Amount,
		"transaction_id":   order.TransactionId,
		"order_ref_number": order.OrderRefNumber,
		"payment_id":       order.PaymentId,
		"items":            sellerItems,
		"created_at":       order.CreatedAt,
		"updated_at":       order.UpdatedAt,
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Order details retrieved successfully",
		"order":   orderResponse,
	})
}
