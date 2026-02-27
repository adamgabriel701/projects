package service

import (
	"errors"
	"time"
	"ecommerce-api/internal/domain"
	"ecommerce-api/internal/repository"

	"github.com/google/uuid"
)

type ProductService struct {
	productRepo *repository.ProductRepository
}

func NewProductService(productRepo *repository.ProductRepository) *ProductService {
	return &ProductService{productRepo: productRepo}
}

func (s *ProductService) Create(req domain.CreateProductRequest) (*domain.Product, error) {
	product := &domain.Product{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		Category:    req.Category,
		ImageURL:    req.ImageURL,
		Active:      true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.productRepo.Create(product); err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) GetByID(id string) (*domain.Product, error) {
	return s.productRepo.FindByID(id)
}

func (s *ProductService) List(filter domain.ProductFilter) ([]domain.Product, int, error) {
	products, total := s.productRepo.FindAll(filter)
	return products, total, nil
}

func (s *ProductService) Update(id string, req domain.UpdateProductRequest) (*domain.Product, error) {
	product, err := s.productRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.Price > 0 {
		product.Price = req.Price
	}
	if req.Stock >= 0 {
		product.Stock = req.Stock
	}
	if req.Category != "" {
		product.Category = req.Category
	}
	if req.ImageURL != "" {
		product.ImageURL = req.ImageURL
	}
	if req.Active != nil {
		product.Active = *req.Active
	}

	product.UpdatedAt = time.Now()

	if err := s.productRepo.Update(product); err != nil {
		return nil, err
	}

	return product, nil
}

func (s *ProductService) Delete(id string) error {
	return s.productRepo.Delete(id)
}

func (s *ProductService) CheckStock(productID string, quantity int) error {
	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		return err
	}

	if product.Stock < quantity {
		return errors.New("estoque insuficiente")
	}

	return nil
}

func (s *ProductService) DecreaseStock(productID string, quantity int) error {
	return s.productRepo.UpdateStock(productID, -quantity)
}
