package helper

import (
	"encoding/json"
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

	// Check for foreign key constraint violations
	if strings.Contains(errMsg, "violates foreign key constraint") {
		// Check if it's a product creation/update with invalid category_id
		if strings.Contains(errMsg, "fk_categories_products") &&
			(strings.Contains(errMsg, "insert or update on table \"products\"") ||
				strings.Contains(errMsg, "insert or update on table 'products'")) {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message":    "Invalid category",
				"error":      "The specified category does not exist. Please provide a valid category ID.",
				"error_full": errMsg,
			})
		}
		// Check if it's a category deletion with associated products
		if strings.Contains(errMsg, "fk_categories_products") &&
			strings.Contains(errMsg, "delete on table \"categories\"") {
			return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
				"message":    "Cannot delete category",
				"error":      "This category has associated products. Please remove or reassign products before deleting the category.",
				"error_full": errMsg,
			})
		}
		// Generic foreign key violation
		return ctx.Status(fiber.StatusConflict).JSON(fiber.Map{
			"message":    "Operation not allowed",
			"error":      "This operation cannot be completed due to existing relationships in the database.",
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

// HandleBodyParserError processes JSON parsing errors and returns specific field errors
func HandleBodyParserError(ctx *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}

	errMsg := err.Error()

	// Check for JSON syntax errors
	var jsonErr *json.SyntaxError
	if errors.As(err, &jsonErr) {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid JSON format",
			"error":   "The request body contains invalid JSON. Please check your JSON syntax.",
			"details": errMsg,
		})
	}

	// Check for JSON unmarshal type errors (field type mismatches)
	var jsonTypeErr *json.UnmarshalTypeError
	if errors.As(err, &jsonTypeErr) {
		fieldName := jsonTypeErr.Field
		expectedType := jsonTypeErr.Type.String()
		actualType := "unknown"

		// Convert Go type names to user-friendly names
		switch expectedType {
		case "int", "int64":
			expectedType = "integer"
			actualType = "number (decimal)"
		case "float64":
			expectedType = "number"
			actualType = "string or other type"
		case "string":
			expectedType = "string"
			actualType = "number or other type"
		case "uint":
			expectedType = "positive integer"
			actualType = "negative number or decimal"
		}

		// Convert field name from JSON tag to readable format
		readableFieldName := formatFieldName(fieldName)

		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid field type",
			"error":   "Field type mismatch",
			"field":   readableFieldName,
			"details": strings.Title(readableFieldName) + " must be a " + expectedType + " but received " + actualType + ".",
		})
	}

	// Check for common error patterns in error message
	lowerErrMsg := strings.ToLower(errMsg)

	// Pattern: "cannot unmarshal X into Y"
	if strings.Contains(lowerErrMsg, "cannot unmarshal") {
		// Try to extract field name from error
		if strings.Contains(lowerErrMsg, "stock") {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid field type",
				"error":   "Field type mismatch",
				"field":   "stock",
				"details": "Stock must be an integer (whole number), not a decimal.",
			})
		}
		if strings.Contains(lowerErrMsg, "category_id") {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid field type",
				"error":   "Field type mismatch",
				"field":   "category_id",
				"details": "Category ID must be a positive integer, not a decimal or string.",
			})
		}
		if strings.Contains(lowerErrMsg, "price") {
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"message": "Invalid field type",
				"error":   "Field type mismatch",
				"field":   "price",
				"details": "Price must be a number (integer or decimal).",
			})
		}
	}

	// Generic error response
	return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
		"message": "Invalid request body",
		"error":   "The request body could not be parsed. Please check that all field types are correct.",
		"details": errMsg,
	})
}

// formatFieldName converts JSON field names to readable format
func formatFieldName(fieldName string) string {
	// Map common field names to readable versions
	fieldMap := map[string]string{
		"stock":       "stock",
		"category_id": "category_id",
		"categoryID":  "category_id",
		"price":       "price",
		"name":        "name",
		"description": "description",
		"image_url":   "image_url",
		"imageURL":    "image_url",
	}

	if readable, ok := fieldMap[strings.ToLower(fieldName)]; ok {
		return readable
	}

	// Default: return as-is
	return fieldName
}
