package helper

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// HandleDBError processes database errors and returns appropriate HTTP responses
func HandleDBError(ctx *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// Check for record not found error (both type check and string check)
	if errors.Is(err, gorm.ErrRecordNotFound) || strings.Contains(strings.ToLower(errMsg), "record not found") {
		return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message":    "Resource not found",
			"error":      "The requested resource does not exist. Please check the ID and try again.",
			"error_full": errMsg,
		})
	}

	// Check for duplicate email error (multiple patterns to catch different error formats)
	if (strings.Contains(errMsg, "duplicate key value violates unique constraint") ||
		strings.Contains(errMsg, "duplicate key") ||
		strings.Contains(errMsg, "email already exists")) &&
		(strings.Contains(errMsg, "uni_users_email") || strings.Contains(errMsg, "email")) {
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message":    "Email already exists",
			"error":      "This email address is already registered to another user. Please use a different email address.",
			"error_full": errMsg,
		})
	}

	// Check for duplicate phone error
	if strings.Contains(errMsg, "duplicate key value violates unique constraint") &&
		strings.Contains(errMsg, "uni_users_phone") {
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message":    "Phone number already exists",
			"error":      "This phone number is already registered. Please use a different phone number.",
			"error_full": errMsg,
		})
	}

	// Check for connection errors
	if strings.Contains(errMsg, "connection refused") || strings.Contains(errMsg, "no connection") {
		return ctx.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"message":    "Database connection failed",
			"error":      "Unable to connect to the database. Please try again later.",
			"error_full": errMsg,
		})
	}

	// Default error response
	return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"message":    "An error occurred while processing your request",
		"error":      errMsg,
		"error_full": errMsg,
	})
}

// HandleValidationError handles validation errors
func HandleValidationError(ctx *fiber.Ctx, message string) error {
	return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"message": message,
		"error":   "Validation failed",
	})
}
