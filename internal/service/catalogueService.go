package service

import (
	"errors"
	"go-ecommerce-app/config"
	"go-ecommerce-app/internal/dto"
	"go-ecommerce-app/internal/helper"
	"go-ecommerce-app/internal/repository"
)

type CatalogueService struct {
	Repo   repository.CatalogueRepository
	Auth   helper.Auth
	Config config.AppConfig
}

func NewCatalogueService(repo repository.CatalogueRepository, auth helper.Auth, config config.AppConfig) CatalogueService {
	return CatalogueService{
		Repo:   repo,
		Auth:   auth,
		Config: config,
	}
}

// Category methods - to be implemented
func (s CatalogueService) CreateCategory(sellerID uint, category dto.Category) (interface{}, error) {
	if sellerID == 0 {
		return nil, errors.New("seller ID is required")
	}

	createdCategory, err := s.Repo.CreateCategory(sellerID, category)
	if err != nil {
		return nil, err
	}

	return createdCategory, nil
}

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
		Take:  query.GetLimit(),
		Skip:  query.GetOffset(),
		Total: total,
	}

	return &dto.PaginatedResponse{
		Data:       result,
		Pagination: pagination,
	}, nil
}

func (s CatalogueService) GetCategoryByID(id uint) (interface{}, error) {
	category, err := s.Repo.GetCategoryByID(id)
	if err != nil {
		return nil, err
	}
	return category, nil
}

func (s CatalogueService) UpdateCategory(id uint, category dto.Category) (interface{}, error) {
	updatedCategory, err := s.Repo.UpdateCategory(id, category)
	if err != nil {
		return nil, err
	}
	return updatedCategory, nil
}

func (s CatalogueService) DeleteCategory(id uint) error {
	productCount, err := s.Repo.CountProductsByCategoryID(id)
	if err != nil {
		return err
	}

	if productCount > 0 {
		return errors.New("cannot delete category: category has associated products. Please remove or reassign products before deleting the category")
	}

	return s.Repo.DeleteCategory(id)
}

// Product methods
func (s CatalogueService) CreateProduct(sellerID uint, product dto.Product) (interface{}, error) {
	if sellerID == 0 {
		return nil, errors.New("seller ID is required")
	}

	createdProduct, err := s.Repo.CreateProduct(sellerID, product)
	if err != nil {
		return nil, err
	}
	return createdProduct, nil
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
		Take:  query.GetLimit(),
		Skip:  query.GetOffset(),
		Total: total,
	}

	return &dto.PaginatedResponse{
		Data:       result,
		Pagination: pagination,
	}, nil
}

func (s CatalogueService) GetProductByID(id uint) (interface{}, error) {
	product, err := s.Repo.GetProductByID(id)
	if err != nil {
		return nil, err
	}
	return product, nil
}

func (s CatalogueService) UpdateProduct(productID uint, sellerID uint, product dto.Product) (interface{}, error) {
	existingProduct, err := s.Repo.GetProductByID(productID)
	if err != nil {
		return nil, err
	}

	if existingProduct.SellerID != sellerID {
		return nil, errors.New("unauthorized: you can only update your own products")
	}

	updatedProduct, err := s.Repo.UpdateProduct(productID, product)
	if err != nil {
		return nil, err
	}
	return updatedProduct, nil
}

func (s CatalogueService) DeleteProduct(id uint) error {
	return s.Repo.DeleteProduct(id)
}
