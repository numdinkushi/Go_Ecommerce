package handlers

import (
	"encoding/json"
	"go-ecommerce-app/config"
	"go-ecommerce-app/internal/api/rest"
	"go-ecommerce-app/internal/dto"
	"go-ecommerce-app/internal/helper"
	"go-ecommerce-app/internal/repository"
	"go-ecommerce-app/internal/service"
	"time"

	"github.com/gofiber/fiber/v2"
)

type CatalogueHandler struct {
	catalogueService service.CatalogueService
	auth             helper.Auth
	config           config.AppConfig
}

func SetupCatalogueRoutes(restHandler *rest.RestHandler, bankService *service.BankService) {
	app := restHandler.App

	catalogueRepo := repository.NewCatalogueRepository(restHandler.DB)
	userRepo := repository.NewUserRepository(restHandler.DB)
	catalogueService := service.NewCatalogueService(catalogueRepo, restHandler.Auth, restHandler.Config)
	handler := CatalogueHandler{
		catalogueService: catalogueService,
		auth:             restHandler.Auth,
		config:           restHandler.Config,
	}

	// Public endpoints (no authentication required)
	app.Get("/products", handler.GetProducts)
	app.Get("/products/:id", handler.GetProductByID)
	app.Get("/categories", handler.GetCategories)
	app.Get("/categories/:id", handler.GetCategoryByID)

	// Private endpoints (authentication required - seller only)
	sellerPrivateRoutes := app.Group("/seller", restHandler.Auth.AuthorizeSeller(userRepo))
	sellerPrivateRoutes.Post("/categories", handler.CreateCategory)
	sellerPrivateRoutes.Patch("/categories/:id", handler.UpdateCategory)
	sellerPrivateRoutes.Delete("/categories/:id", handler.DeleteCategory)
	sellerPrivateRoutes.Get("/categories/:id", handler.GetCategoryByID)

	sellerPrivateRoutes.Post("/products", handler.CreateProduct)
	sellerPrivateRoutes.Get("/products", handler.GetProducts)
	sellerPrivateRoutes.Get("/products/:id", handler.GetProductByID)
	sellerPrivateRoutes.Put("/products/:id", handler.UpdateProduct)
	sellerPrivateRoutes.Patch("/products/:id", handler.PatchProduct)
	sellerPrivateRoutes.Delete("/products/:id", handler.DeleteProduct)
}

// Category Handlers

func (h *CatalogueHandler) CreateCategory(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)
	if user.ID == 0 {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	category := dto.Category{}
	err := ctx.BodyParser(&category)
	if err != nil {
		return helper.HandleBodyParserError(ctx, err)
	}

	if category.Name == "" {
		return helper.HandleValidationError(ctx, "Field 'name' is required")
	}

	createdCategory, err := h.catalogueService.CreateCategory(user.ID, category)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":  "Category created successfully",
		"category": createdCategory,
	})
}

func (h *CatalogueHandler) GetCategories(ctx *fiber.Ctx) error {
	query := dto.CategoryQuery{}

	if err := ctx.QueryParser(&query); err != nil {
		return helper.HandleValidationError(ctx, "Invalid query parameters")
	}

	if query.Take < 1 {
		query.Take = 10
	}
	if query.Skip < 0 {
		query.Skip = 0
	}

	if beginningStr := ctx.Query("beginning"); beginningStr != "" {
		beginning, err := time.Parse(time.RFC3339, beginningStr)
		if err != nil {
			return helper.HandleValidationError(ctx, "Invalid beginning date format. Use ISO 8601 format (e.g., 2024-01-01T00:00:00Z)")
		}
		query.Beginning = &beginning
	}

	if endingStr := ctx.Query("ending"); endingStr != "" {
		ending, err := time.Parse(time.RFC3339, endingStr)
		if err != nil {
			return helper.HandleValidationError(ctx, "Invalid ending date format. Use ISO 8601 format (e.g., 2024-02-01T00:00:00Z)")
		}
		query.Ending = &ending
	}

	result, err := h.catalogueService.GetCategories(query)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":    "Categories retrieved successfully",
		"data":       result.Data,
		"pagination": result.Pagination,
	})
}

func (h *CatalogueHandler) GetCategoryByID(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return helper.HandleValidationError(ctx, "Invalid category ID")
	}

	category, err := h.catalogueService.GetCategoryByID(uint(id))
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Category retrieved successfully",
		"category": category,
	})
}

func (h *CatalogueHandler) UpdateCategory(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return helper.HandleValidationError(ctx, "Invalid category ID")
	}

	category := dto.Category{}
	if err := ctx.BodyParser(&category); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	updatedCategory, err := h.catalogueService.UpdateCategory(uint(id), category)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":  "Category updated successfully",
		"category": updatedCategory,
	})
}

func (h *CatalogueHandler) DeleteCategory(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return helper.HandleValidationError(ctx, "Invalid category ID")
	}

	err = h.catalogueService.DeleteCategory(uint(id))
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Category deleted successfully",
	})
}

// Product Handlers

func (h *CatalogueHandler) CreateProduct(ctx *fiber.Ctx) error {
	user := h.auth.GetCurrentUser(ctx)
	if user.ID == 0 {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	product := dto.Product{}
	err := ctx.BodyParser(&product)
	if err != nil {
		return helper.HandleBodyParserError(ctx, err)
	}

	// Validate required fields
	if product.Name == "" {
		return helper.HandleValidationError(ctx, "Field 'name' is required")
	}
	if product.Price <= 0 {
		return helper.HandleValidationError(ctx, "Field 'price' must be greater than 0")
	}
	if product.CategoryID == 0 {
		return helper.HandleValidationError(ctx, "Field 'category_id' is required and must be a valid category ID")
	}
	if product.Stock < 0 {
		return helper.HandleValidationError(ctx, "Field 'stock' cannot be negative")
	}

	createdProduct, err := h.catalogueService.CreateProduct(user.ID, product)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Product created successfully",
		"product": createdProduct,
	})
}

func (h *CatalogueHandler) GetProducts(ctx *fiber.Ctx) error {
	query := dto.ProductQuery{}

	if err := ctx.QueryParser(&query); err != nil {
		return helper.HandleValidationError(ctx, "Invalid query parameters")
	}

	if query.Take < 1 {
		query.Take = 10
	}
	if query.Skip < 0 {
		query.Skip = 0
	}

	if beginningStr := ctx.Query("beginning"); beginningStr != "" {
		beginning, err := time.Parse(time.RFC3339, beginningStr)
		if err != nil {
			return helper.HandleValidationError(ctx, "Invalid beginning date format. Use ISO 8601 format (e.g., 2024-01-01T00:00:00Z)")
		}
		query.Beginning = &beginning
	}

	if endingStr := ctx.Query("ending"); endingStr != "" {
		ending, err := time.Parse(time.RFC3339, endingStr)
		if err != nil {
			return helper.HandleValidationError(ctx, "Invalid ending date format. Use ISO 8601 format (e.g., 2024-02-01T00:00:00Z)")
		}
		query.Ending = &ending
	}

	result, err := h.catalogueService.GetProducts(query)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message":    "Products retrieved successfully",
		"data":       result.Data,
		"pagination": result.Pagination,
	})
}

func (h *CatalogueHandler) GetProductByID(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return helper.HandleValidationError(ctx, "Invalid product ID")
	}

	product, err := h.catalogueService.GetProductByID(uint(id))
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Product retrieved successfully",
		"product": product,
	})
}

func (h *CatalogueHandler) UpdateProduct(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return helper.HandleValidationError(ctx, "Invalid product ID")
	}

	user := h.auth.GetCurrentUser(ctx)
	if user.ID == 0 {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	product := dto.Product{}
	if err := ctx.BodyParser(&product); err != nil {
		return helper.HandleBodyParserError(ctx, err)
	}

	// Validate required fields
	if product.Name == "" {
		return helper.HandleValidationError(ctx, "Field 'name' is required")
	}
	if product.Price <= 0 {
		return helper.HandleValidationError(ctx, "Field 'price' must be greater than 0")
	}
	if product.CategoryID == 0 {
		return helper.HandleValidationError(ctx, "Field 'category_id' is required and must be a valid category ID")
	}
	if product.Stock < 0 {
		return helper.HandleValidationError(ctx, "Field 'stock' cannot be negative")
	}

	updatedProduct, err := h.catalogueService.UpdateProduct(uint(id), user.ID, product)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Product updated successfully",
		"product": updatedProduct,
	})
}

func (h *CatalogueHandler) PatchProduct(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return helper.HandleValidationError(ctx, "Invalid product ID")
	}

	user := h.auth.GetCurrentUser(ctx)
	if user.ID == 0 {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Unauthorized",
		})
	}

	// Read body bytes once
	bodyBytes := ctx.Body()
	if len(bodyBytes) == 0 {
		return helper.HandleValidationError(ctx, "Request body cannot be empty. Provide at least one field to update")
	}

	// Parse body as map to detect which fields are provided
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &bodyMap); err != nil {
		return helper.HandleBodyParserError(ctx, err)
	}

	if len(bodyMap) == 0 {
		return helper.HandleValidationError(ctx, "Request body cannot be empty. Provide at least one field to update")
	}

	// Parse body into product DTO
	product := dto.Product{}
	if err := json.Unmarshal(bodyBytes, &product); err != nil {
		return helper.HandleBodyParserError(ctx, err)
	}

	// For PATCH, only validate fields that are actually provided in the request
	if _, provided := bodyMap["name"]; provided {
		if product.Name == "" {
			return helper.HandleValidationError(ctx, "Field 'name' cannot be empty")
		}
	}
	if _, provided := bodyMap["price"]; provided {
		if product.Price <= 0 {
			return helper.HandleValidationError(ctx, "Field 'price' must be greater than 0")
		}
	}
	if _, provided := bodyMap["category_id"]; provided {
		if product.CategoryID == 0 {
			return helper.HandleValidationError(ctx, "Field 'category_id' must be a valid category ID")
		}
	}
	if _, provided := bodyMap["stock"]; provided {
		if product.Stock < 0 {
			return helper.HandleValidationError(ctx, "Field 'stock' cannot be negative")
		}
	}

	updatedProduct, err := h.catalogueService.UpdateProduct(uint(id), user.ID, product)
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Product updated successfully",
		"product": updatedProduct,
	})
}

func (h *CatalogueHandler) DeleteProduct(ctx *fiber.Ctx) error {
	id, err := ctx.ParamsInt("id")
	if err != nil {
		return helper.HandleValidationError(ctx, "Invalid product ID")
	}

	err = h.catalogueService.DeleteProduct(uint(id))
	if err != nil {
		return helper.HandleDBError(ctx, err)
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Product deleted successfully",
	})
}
