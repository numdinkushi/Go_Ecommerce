package handlers

import (
	"fmt"

	"go-ecommerce-app/config"
	"go-ecommerce-app/internal/api/rest"
	"go-ecommerce-app/internal/dto"
	"go-ecommerce-app/internal/helper"
	"go-ecommerce-app/internal/repository"
	"go-ecommerce-app/internal/service"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type TransactionHandler struct {
	transactionService *service.TransactionService
	auth               helper.Auth
	userService        service.UserService
	userRepo           repository.UserRepository
}

func initializeTransactionService(db *gorm.DB, auth helper.Auth, config config.AppConfig) *service.TransactionService {
	transactionRepo := repository.NewTransactionRepository(db)
	return service.NewTransactionService(transactionRepo, auth, config)
}

func SetupTransactionRoutes(restHandler *rest.RestHandler, bankService *service.BankService) {
	app := restHandler.App

	transactionService := initializeTransactionService(restHandler.DB, restHandler.Auth, restHandler.Config)
	userRepo := repository.NewUserRepository(restHandler.DB)
	catalogueRepo := repository.NewCatalogueRepository(restHandler.DB)
	userService := service.NewUserService(userRepo, catalogueRepo, restHandler.Auth, restHandler.Config, bankService)

	handler := TransactionHandler{
		transactionService: transactionService,
		auth:               restHandler.Auth,
		userService:        userService,
		userRepo:           userRepo,
	}

	// General authenticated routes
	secRoute := app.Group("/", restHandler.Auth.Authorize)
	secRoute.Post("/payment", handler.MakePayment)

	// Seller-specific authenticated routes
	sellerRoute := app.Group("/seller", restHandler.Auth.AuthorizeSeller(userRepo))
	sellerRoute.Get("/orders", handler.GetOrders)
	sellerRoute.Get("/orders/:id", handler.GetOrderDetails)

	// Private endpoints (authentication required)
	privateRoutes := app.Group("/", restHandler.Auth.Authorize)
	privateRoutes.Post("/transactions", handler.CreateTransaction)
	privateRoutes.Get("/transactions", handler.GetTransactions)
	privateRoutes.Get("/transactions/:id", handler.GetTransactionByID)
	privateRoutes.Post("/transactions/:id/verify", handler.VerifyPaymentStatus)

	// Webhook endpoints (no authentication required, verified by signature)
	app.Post("/webhooks/payment/:gateway", handler.HandlePaymentWebhook)
}

func NewTransactionHandler(transactionService *service.TransactionService, auth helper.Auth) *TransactionHandler {
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

	transaction, err := h.transactionService.CreateTransaction(user.ID, transactionInput)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Transaction created successfully",
		"transaction": transaction,
	})
}

func (h *TransactionHandler) GetTransactions(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)

	transactions, err := h.transactionService.GetTransactionsByUserID(user.ID)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":      "Transactions retrieved successfully",
		"transactions": transactions,
		"count":        len(transactions),
	})
}

func (h *TransactionHandler) GetTransactionByID(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)
	transactionID, err := ctx.ParamsInt("id")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid transaction ID",
		})
	}

	transaction, err := h.transactionService.GetTransactionByIDAndUserID(uint(transactionID), user.ID)
	if err != nil {
		if err.Error() == "transaction not found" {
			return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "You don't have access to this transaction",
			})
		}
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":     "Transaction retrieved successfully",
		"transaction": transaction,
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

	// 1. Dual-source amount calculation: Try cart first, fallback to order if cart is empty
	var calculatedAmount float64

	// First, try to get amount from cart (new payment flow)
	cartItems, err := h.userRepo.FindCartByUserID(user.ID)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to retrieve cart items",
			"error":   err.Error(),
		})
	}

	if len(cartItems) > 0 {
		// Cart has items - use cart for amount calculation (new payment)
		for _, cartItem := range cartItems {
			calculatedAmount += cartItem.Price * float64(cartItem.Quantity)
		}
	} else {
		// Cart is empty - check if order_id is provided for retry payment
		if paymentInput.OrderID == 0 {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Cart is empty and no order_id provided. Please add items to cart or provide a valid order_id for payment retry",
			})
		}

		// Fetch order from database
		order, err := h.userRepo.FindOrderByID(paymentInput.OrderID)
		if err != nil {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Order not found",
				"error":   err.Error(),
			})
		}

		// Validate order ownership
		if order.UserID != user.ID {
			return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"message": "You don't have permission to pay for this order",
			})
		}

		// Validate order status - only allow payment for pending orders
		if order.Status != "pending" {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": fmt.Sprintf("Cannot create payment for order with status '%s'. Only pending orders can be paid", order.Status),
			})
		}

		// Use order amount for payment (retry payment flow)
		calculatedAmount = order.Amount
	}

	// Validate URLs are provided
	if paymentInput.SuccessURL == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "success_url is required",
		})
	}
	if paymentInput.CancelURL == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "cancel_url is required",
		})
	}

	// Create payment session using server-calculated amount (never trust client-provided amounts)
	payment, err := h.transactionService.ProcessPayment(user.ID, paymentInput, calculatedAmount, paymentInput.SuccessURL, paymentInput.CancelURL)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Payment processed successfully",
		"payment": payment,
	})
}

func (h *TransactionHandler) GetOrders(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)

	orders, err := h.userService.GetOrdersBySellerID(user.ID)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Orders retrieved successfully",
		"orders":  orders,
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

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Order details retrieved successfully",
		"order":   order,
	})
}

func (h *TransactionHandler) VerifyPaymentStatus(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)
	transactionID, err := ctx.ParamsInt("id")
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid transaction ID",
		})
	}

	gateway := ctx.Query("gateway")
	if gateway == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Payment gateway is required",
		})
	}

	transaction, err := h.transactionService.VerifyPaymentStatus(uint(transactionID), gateway)
	if err != nil {
		if err.Error() == "transaction not found" {
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"message": "Transaction not found",
			})
		}
		return helper.HandleDBError(ctx, err)
	}

	if transaction.UserID != user.ID {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "You don't have access to this transaction",
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":     "Payment status verified successfully",
		"transaction": transaction,
	})
}

func (h *TransactionHandler) HandlePaymentWebhook(ctx *fiber.Ctx) error {
	gateway := ctx.Params("gateway")
	if gateway == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Payment gateway is required",
		})
	}

	payload := ctx.Body()
	signature := ctx.Get("X-Signature")
	if signature == "" {
		signature = ctx.Get("stripe-signature")
	}
	if signature == "" {
		signature = ctx.Get("verif-hash")
	}

	if signature == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Webhook signature is required",
		})
	}

	err := h.transactionService.ProcessWebhook(gateway, payload, signature)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Failed to process webhook",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Webhook processed successfully",
	})
}
