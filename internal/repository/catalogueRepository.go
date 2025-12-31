package repository

import (
	"go-ecommerce-app/internal/domain"
	"go-ecommerce-app/internal/dto"
	"log"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CatalogueRepository interface {
	// Category methods
	CreateCategory(sellerID uint, category dto.Category) (*domain.Category, error)
	GetCategories(query dto.CategoryQuery) ([]domain.Category, int64, error)
	GetCategoryByID(id uint) (*domain.Category, error)
	UpdateCategory(id uint, category dto.Category) (*domain.Category, error)
	CountProductsByCategoryID(categoryID uint) (int64, error)
	DeleteCategory(id uint) error

	// Product methods
	CreateProduct(sellerID uint, product dto.Product) (*domain.Product, error)
	GetProducts(query dto.ProductQuery) ([]domain.Product, int64, error)
	GetProductByID(id uint) (*domain.Product, error)
	UpdateProduct(id uint, product dto.Product) (*domain.Product, error)
	DeleteProduct(id uint) error
}

type catalogueRepository struct {
	DB *gorm.DB
}

func NewCatalogueRepository(db *gorm.DB) CatalogueRepository {
	return &catalogueRepository{
		DB: db,
	}
}

// Category methods

func (r *catalogueRepository) CreateCategory(sellerID uint, category dto.Category) (*domain.Category, error) {
	categoryDomain := domain.Category{
		Name:         category.Name,
		Description:  category.Description,
		SellerID:     sellerID,
		ParentID:     category.ParentID,
		ImageURL:     category.ImageURL,
		DisplayOrder: category.DisplayOrder,
	}

	err := r.DB.Create(&categoryDomain).Error
	if err != nil {
		log.Printf("Failed to create category: %v", err)
		return nil, err
	}

	log.Println("Category created successfully")
	return &categoryDomain, nil
}

func (r *catalogueRepository) GetCategories(query dto.CategoryQuery) ([]domain.Category, int64, error) {
	var categories []domain.Category
	var total int64

	db := r.DB.Model(&domain.Category{})

	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		db = db.Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
	}

	if query.Beginning != nil {
		db = db.Where("created_at >= ?", *query.Beginning)
	}

	if query.Ending != nil {
		db = db.Where("created_at <= ?", *query.Ending)
	}

	if query.ParentID != nil {
		db = db.Where("parent_id = ?", *query.ParentID)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	db = db.Order("display_order ASC, created_at DESC")

	offset := query.GetOffset()
	limit := query.GetLimit()
	err := db.Offset(offset).Limit(limit).Find(&categories).Error

	if err != nil {
		return nil, 0, err
	}

	return categories, total, nil
}

func (r *catalogueRepository) GetCategoryByID(id uint) (*domain.Category, error) {
	var category domain.Category
	err := r.DB.First(&category, id).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *catalogueRepository) UpdateCategory(id uint, category dto.Category) (*domain.Category, error) {
	var categoryDomain domain.Category
	err := r.DB.First(&categoryDomain, id).Error
	if err != nil {
		return nil, err
	}

	categoryDomain.Name = category.Name
	if category.Description != "" {
		categoryDomain.Description = category.Description
	}
	if category.ParentID != nil {
		categoryDomain.ParentID = category.ParentID
	}
	if category.ImageURL != "" {
		categoryDomain.ImageURL = category.ImageURL
	}
	categoryDomain.DisplayOrder = category.DisplayOrder

	err = r.DB.Model(&categoryDomain).Clauses(clause.Returning{}).Updates(categoryDomain).Error
	if err != nil {
		log.Printf("Error updating category: %v", err)
		return nil, err
	}

	return &categoryDomain, nil
}

func (r *catalogueRepository) CountProductsByCategoryID(categoryID uint) (int64, error) {
	var count int64
	err := r.DB.Model(&domain.Product{}).Where("category_id = ?", categoryID).Count(&count).Error
	return count, err
}

func (r *catalogueRepository) DeleteCategory(id uint) error {
	return r.DB.Delete(&domain.Category{}, id).Error
}

// Product methods

func (r *catalogueRepository) CreateProduct(sellerID uint, product dto.Product) (*domain.Product, error) {
	productDomain := domain.Product{
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		CategoryID:  product.CategoryID,
		Stock:       product.Stock,
		ImageURL:    product.ImageURL,
		SellerID:    sellerID,
	}

	err := r.DB.Create(&productDomain).Error
	if err != nil {
		log.Printf("Failed to create product: %v", err)
		return nil, err
	}

	log.Println("Product created successfully")
	return &productDomain, nil
}

func (r *catalogueRepository) GetProducts(query dto.ProductQuery) ([]domain.Product, int64, error) {
	var products []domain.Product
	var total int64

	db := r.DB.Model(&domain.Product{})

	if query.Search != "" {
		searchPattern := "%" + query.Search + "%"
		db = db.Where("name ILIKE ? OR description ILIKE ?", searchPattern, searchPattern)
	}

	if query.Beginning != nil {
		db = db.Where("created_at >= ?", *query.Beginning)
	}

	if query.Ending != nil {
		db = db.Where("created_at <= ?", *query.Ending)
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	db = db.Order("created_at DESC")

	offset := query.GetOffset()
	limit := query.GetLimit()
	err := db.Offset(offset).Limit(limit).Find(&products).Error

	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (r *catalogueRepository) GetProductByID(id uint) (*domain.Product, error) {
	var product domain.Product
	err := r.DB.First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *catalogueRepository) UpdateProduct(id uint, product dto.Product) (*domain.Product, error) {
	var productDomain domain.Product
	err := r.DB.First(&productDomain, id).Error
	if err != nil {
		return nil, err
	}

	productDomain.Name = product.Name
	if product.Description != "" {
		productDomain.Description = product.Description
	}
	productDomain.Price = product.Price
	productDomain.CategoryID = product.CategoryID
	productDomain.Stock = product.Stock
	if product.ImageURL != "" {
		productDomain.ImageURL = product.ImageURL
	}

	err = r.DB.Model(&productDomain).Clauses(clause.Returning{}).Updates(productDomain).Error
	if err != nil {
		log.Printf("Error updating product: %v", err)
		return nil, err
	}

	return &productDomain, nil
}

func (r *catalogueRepository) DeleteProduct(id uint) error {
	return r.DB.Delete(&domain.Product{}, id).Error
}
