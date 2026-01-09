package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

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
	userRepo := repository.NewUserRepository(db)
	return service.NewTransactionService(transactionRepo, userRepo, auth, config)
}

// SetupWebhookRoutes registers webhook endpoints as public routes (no auth required)
// This MUST be called before any other route setup to ensure webhooks are not caught by auth middleware
func SetupWebhookRoutes(restHandler *rest.RestHandler) {
	app := restHandler.App

	transactionService := initializeTransactionService(restHandler.DB, restHandler.Auth, restHandler.Config)
	userRepo := repository.NewUserRepository(restHandler.DB)
	catalogueRepo := repository.NewCatalogueRepository(restHandler.DB)
	var bankService *service.BankService // Webhooks don't need bank service
	userService := service.NewUserService(userRepo, catalogueRepo, restHandler.Auth, restHandler.Config, bankService)

	handler := TransactionHandler{
		transactionService: transactionService,
		auth:               restHandler.Auth,
		userService:        userService,
		userRepo:           userRepo,
	}

	app.Post("/webhooks/payment/:gateway", func(c *fiber.Ctx) error {
		gateway := c.Params("gateway")
		if gateway == "test" {
			return c.Status(fiber.StatusMethodNotAllowed).JSON(fiber.Map{
				"message": "Use GET /webhooks/payment/test for testing",
			})
		}
		return handler.HandlePaymentWebhook(c)
	})

	log.Printf("Webhook endpoint registered: POST /webhooks/payment/:gateway")

	app.Get("/webhooks/payment/test", func(c *fiber.Ctx) error {
		return c.Status(200).JSON(fiber.Map{
			"message":  "Webhook endpoint is reachable",
			"endpoint": "/webhooks/payment/stripe",
			"method":   "POST",
			"note":     "Configure this URL in Stripe Dashboard → Developers → Webhooks",
		})
	})

	app.Get("/webhooks/payment/info", func(c *fiber.Ctx) error {
		webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
		hasSecret := webhookSecret != ""

		ngrokURL := ""
		ngrokInfoURL := "http://localhost:4040/api/tunnels"
		info := fiber.Map{
			"webhook_endpoint":          "/webhooks/payment/stripe",
			"method":                    "POST",
			"webhook_secret_configured": hasSecret,
			"required_events": []string{
				"checkout.session.completed",
				"payment_intent.succeeded",
				"charge.succeeded",
			},
		}

		resp, err := http.Get(ngrokInfoURL)
		if err == nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			re := regexp.MustCompile(`https://[^"]*\.ngrok[^"]*`)
			matches := re.FindString(string(body))
			if matches != "" {
				ngrokURL = matches
				info["ngrok_url"] = ngrokURL
				info["webhook_url"] = ngrokURL + "/webhooks/payment/stripe"
				info["status"] = "Configure this URL in Stripe Dashboard → Developers → Webhooks"
			}
		}

		if ngrokURL == "" {
			info["status"] = "ngrok not detected. Start ngrok with: ngrok http 9000"
		}

		if !hasSecret {
			info["warning"] = "STRIPE_WEBHOOK_SECRET is not set in environment variables"
		}

		return c.Status(200).JSON(info)
	})
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

	secRoute := app.Group("/", restHandler.Auth.Authorize)
	secRoute.Post("/payment", handler.MakePayment)

	sellerRoute := app.Group("/seller", restHandler.Auth.AuthorizeSeller(userRepo))
	sellerRoute.Get("/orders", handler.GetOrders)
	sellerRoute.Get("/orders/:id", handler.GetOrderDetails)

	privateRoutes := app.Group("/", restHandler.Auth.Authorize)
	privateRoutes.Post("/transactions", handler.CreateTransaction)
	privateRoutes.Get("/transactions", handler.GetTransactions)
	privateRoutes.Get("/transactions/:id", handler.GetTransactionByID)
	privateRoutes.Post("/transactions/:id/verify", handler.VerifyPaymentStatus)
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

	if paymentInput.OrderID == 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "order_id is required. Please create an order first, then provide the order_id for payment",
		})
	}

	order, err := h.userRepo.FindOrderByID(paymentInput.OrderID)
	if err != nil {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Order not found",
			"error":   err.Error(),
		})
	}

	if order.UserID != user.ID {
		return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "You don't have permission to pay for this order",
		})
	}

	if order.Status != "pending" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": fmt.Sprintf("Cannot create payment for order with status '%s'. Only pending orders can be paid", order.Status),
		})
	}

	calculatedAmount := order.Amount

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

	var verifyInput dto.VerifyPaymentInput
	if err := ctx.BodyParser(&verifyInput); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if verifyInput.Gateway == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Payment gateway is required",
		})
	}

	transaction, err := h.transactionService.VerifyPaymentStatus(uint(transactionID), verifyInput.Gateway)
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
			"path":    ctx.Path(),
		})
	}

	payload := ctx.Body()

	signature := ctx.Get("Stripe-Signature")
	if signature == "" {
		signature = ctx.Get("stripe-signature")
	}
	if signature == "" {
		signature = ctx.Get("X-Signature")
	}

	if signature == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Webhook signature is required",
			"error":   "Missing Stripe-Signature header",
		})
	}

	err := h.transactionService.ProcessWebhook(gateway, payload, signature)
	if err != nil {
		errMsg := err.Error()

		if errMsg == "STRIPE_WEBHOOK_SECRET not configured" ||
			errMsg == "invalid webhook signature" ||
			errMsg == "signature verification failed" ||
			strings.Contains(errMsg, "signature") {
			log.Printf("Webhook signature verification failed: %v", err)
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Webhook signature verification failed",
				"message": "Invalid webhook signature. Check STRIPE_WEBHOOK_SECRET configuration.",
			})
		}

		if strings.Contains(errMsg, "not actually succeeded") ||
			strings.Contains(errMsg, "requires_confirmation") ||
			strings.Contains(errMsg, "not finalized") {
			return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
				"message": "Webhook received but payment not finalized",
				"error":   err.Error(),
				"note":    "Payment is still processing. Stripe will send another webhook when finalized.",
			})
		}

		log.Printf("Webhook processing failed: %v", err)
		return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Webhook received but processing failed",
			"error":   err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Webhook processed successfully",
	})
}
