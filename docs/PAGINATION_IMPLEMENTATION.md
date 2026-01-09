# Professional Pagination, Search & Filtering Implementation Guide

## Architecture Overview

This implementation follows **Clean Architecture**, **DRY**, **SOLID** principles:

1. **DTO Layer** (`dto/`) - Query parameters and response structures
2. **Repository Layer** (`repository/`) - Database queries with filtering
3. **Service Layer** (`service/`) - Business logic and pagination metadata
4. **Handler Layer** (`handlers/`) - HTTP request parsing

## 1. Repository Layer Implementation

### Update Repository Interface

```go
// internal/repository/catalogueRepository.go

type CatalogueRepository interface {
    // Category methods
    CreateCategory(sellerID uint, category dto.Category) (*domain.Category, error)
    GetCategories(query dto.CategoryQuery) ([]domain.Category, int64, error)
    GetCategoryByID(id uint) (*domain.Category, error)
    UpdateCategory(id uint, category dto.Category) (*domain.Category, error)
    DeleteCategory(id uint) error

    // Product methods
    CreateProduct(sellerID uint, product dto.Product) (*domain.Product, error)
    GetProducts(query dto.ProductQuery) ([]domain.Product, int64, error)
    GetProductByID(id uint) (*domain.Product, error)
    UpdateProduct(id uint, product dto.Product) (*domain.Product, error)
    DeleteProduct(id uint) error
}
```

### Implement GetCategories with Filtering

```go
func (r *catalogueRepository) GetCategories(query dto.CategoryQuery) ([]domain.Category, int64, error) {
    var categories []domain.Category
    var total int64

    // Build query
    db := r.DB.Model(&domain.Category{})

    // Search filter
    if query.Search != "" {
        searchPattern := "%" + query.Search + "%"
        db = db.Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
    }

    // Parent ID filter
    if query.ParentID != nil {
        db = db.Where("parent_id = ?", *query.ParentID)
    } else {
        // If not specified, you might want only top-level categories
        // db = db.Where("parent_id IS NULL")
    }

    // Count total (before pagination)
    if err := db.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // Sorting
    sortBy := query.GetSortBy()
    sortOrder := query.GetSortOrder()
    db = db.Order(sortBy + " " + sortOrder)

    // Pagination
    offset := query.GetOffset()
    limit := query.GetLimit()
    err := db.Offset(offset).Limit(limit).Find(&categories).Error

    if err != nil {
        return nil, 0, err
    }

    return categories, total, nil
}
```

### Implement GetProducts with Filtering

```go
func (r *catalogueRepository) GetProducts(query dto.ProductQuery) ([]domain.Product, int64, error) {
    var products []domain.Product
    var total int64

    // Build query
    db := r.DB.Model(&domain.Product{})

    // Search filter
    if query.Search != "" {
        searchPattern := "%" + query.Search + "%"
        db = db.Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
    }

    // Category filter
    if query.CategoryID != nil {
        db = db.Where("category_id = ?", *query.CategoryID)
    }

    // Price range filter
    if query.MinPrice != nil {
        db = db.Where("price >= ?", *query.MinPrice)
    }
    if query.MaxPrice != nil {
        db = db.Where("price <= ?", *query.MaxPrice)
    }

    // Count total (before pagination)
    if err := db.Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // Sorting
    sortBy := query.GetSortBy()
    sortOrder := query.GetSortOrder()
    db = db.Order(sortBy + " " + sortOrder)

    // Pagination
    offset := query.GetOffset()
    limit := query.GetLimit()
    err := db.Offset(offset).Limit(limit).Find(&products).Error

    if err != nil {
        return nil, 0, err
    }

    return products, total, nil
}
```

## 2. Service Layer Implementation

### Update Service Methods

```go
// internal/service/catalogueService.go

func (s CatalogueService) GetCategories(query dto.CategoryQuery) (*dto.PaginatedResponse, error) {
    categories, total, err := s.Repo.GetCategories(query)
    if err != nil {
        return nil, err
    }

    result := make([]interface{}, len(categories))
    for i, category := range categories {
        result[i] = category
    }

    pagination := dto.PaginationMeta{
        Page:       query.Page,
        Limit:      query.GetLimit(),
        Total:      total,
        TotalPages: int((total + int64(query.GetLimit()) - 1) / int64(query.GetLimit())),
    }

    return &dto.PaginatedResponse{
        Data:       result,
        Pagination: pagination,
    }, nil
}

func (s CatalogueService) GetProducts(query dto.ProductQuery) (*dto.PaginatedResponse, error) {
    products, total, err := s.Repo.GetProducts(query)
    if err != nil {
        return nil, err
    }

    result := make([]interface{}, len(products))
    for i, product := range products {
        result[i] = product
    }

    pagination := dto.PaginationMeta{
        Page:       query.Page,
        Limit:      query.GetLimit(),
        Total:      total,
        TotalPages: int((total + int64(query.GetLimit()) - 1) / int64(query.GetLimit())),
    }

    return &dto.PaginatedResponse{
        Data:       result,
        Pagination: pagination,
    }, nil
}
```

## 3. Handler Layer Implementation

### Update Handlers

```go
// internal/api/rest/handlers/catalogueHandler.go

func (h *CatalogueHandler) GetCategories(ctx *fiber.Ctx) error {
    query := dto.CategoryQuery{}
    
    // Parse query parameters
    if err := ctx.QueryParser(&query); err != nil {
        return helper.HandleValidationError(ctx, "Invalid query parameters")
    }

    // Set defaults
    if query.Page < 1 {
        query.Page = 1
    }
    if query.Limit < 1 {
        query.Limit = 10
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

func (h *CatalogueHandler) GetProducts(ctx *fiber.Ctx) error {
    query := dto.ProductQuery{}
    
    // Parse query parameters
    if err := ctx.QueryParser(&query); err != nil {
        return helper.HandleValidationError(ctx, "Invalid query parameters")
    }

    // Set defaults
    if query.Page < 1 {
        query.Page = 1
    }
    if query.Limit < 1 {
        query.Limit = 10
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
```

## 4. Usage Examples

### Products Endpoint

```bash
# Basic pagination
GET /products?page=1&limit=10

# Search
GET /products?search=laptop&page=1&limit=10

# Filter by category
GET /products?category_id=5&page=1&limit=10

# Price range filter
GET /products?min_price=100&max_price=500&page=1&limit=10

# Sorting
GET /products?sort_by=price&sort_order=asc&page=1&limit=10

# Combined filters
GET /products?search=laptop&category_id=5&min_price=100&max_price=1000&sort_by=price&sort_order=asc&page=1&limit=20
```

### Categories Endpoint

```bash
# Basic pagination
GET /categories?page=1&limit=10

# Search
GET /categories?search=electronics&page=1&limit=10

# Filter by parent (get subcategories)
GET /categories?parent_id=1&page=1&limit=10

# Top-level categories only (parent_id is null)
GET /categories?page=1&limit=10

# Sorting
GET /categories?sort_by=display_order&sort_order=asc&page=1&limit=10
```

### Response Format

```json
{
  "message": "Products retrieved successfully",
  "data": [
    {
      "id": 1,
      "name": "Laptop",
      "price": 999.99,
      ...
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 150,
    "total_pages": 15
  }
}
```

## Key Benefits

1. **DRY**: Reusable pagination structures
2. **SOLID**: Single Responsibility - each layer handles its concern
3. **Modular**: Query DTOs can be reused/extended
4. **Type-Safe**: Compile-time validation
5. **Testable**: Each layer can be tested independently
6. **Scalable**: Easy to add new filters or sorting options

