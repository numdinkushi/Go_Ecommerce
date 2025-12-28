package handlers

import (
	"go-ecommerce-app/internal/api/rest"
	"go-ecommerce-app/internal/dto"
	"go-ecommerce-app/internal/service"

	"github.com/gofiber/fiber/v2"
)

type BankHandler struct {
	bankService *service.BankService
}

func SetupBankRoutes(restHandler *rest.RestHandler, bankService *service.BankService) {
	app := restHandler.App
	handler := BankHandler{
		bankService: bankService,
	}

	// Public endpoint
	app.Get("/banks", handler.GetBanks)

	// Private endpoint (requires authentication)
	privateRoutes := app.Group("/", restHandler.Auth.Authorize)
	privateRoutes.Post("/banks/verify", handler.VerifyAccount)
}

func (h *BankHandler) GetBanks(ctx *fiber.Ctx) error {
	if h.bankService == nil {
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"message": "Bank service is not available",
			"error":   "FLUTTERWAVE_SECRET_KEY is not configured",
		})
	}

	banks, err := h.bankService.GetBanks()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success":    false,
			"message":    "Failed to fetch banks",
			"error":      err.Error(),
			"error_full": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Banks retrieved successfully",
		"data":    banks,
	})
}

func (h *BankHandler) VerifyAccount(ctx *fiber.Ctx) error {
	if h.bankService == nil {
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"message": "Bank service is not available",
			"error":   "FLUTTERWAVE_SECRET_KEY is not configured",
		})
	}

	verifyInput := dto.VerifyAccountInput{}
	if err := ctx.BodyParser(&verifyInput); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if verifyInput.AccountNumber == "" || verifyInput.BankCode == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "account_number and bank_code are required",
		})
	}

	verificationResult, err := h.bankService.VerifyAccount(verifyInput.AccountNumber, verifyInput.BankCode)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success":    false,
			"message":    "Bank account verification failed",
			"error":      err.Error(),
			"error_full": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"message": "Bank account verified successfully",
		"data":    verificationResult,
	})
}
