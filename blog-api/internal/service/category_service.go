package service

import (
	"errors"
	"time"
	"blog-api/internal/domain"
	"blog-api/internal/repository"
	"blog-api/internal/utils"

	"github.com/google/uuid"
)

type CategoryService struct {
	categoryRepo *repository.CategoryRepository
}

func NewCategoryService(categoryRepo *repository.CategoryRepository) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo}
}

func (s *CategoryService) Create(req domain.CreateCategoryRequest) (*domain.Category, error) {
	slug := utils.GenerateSlug(req.Name)
	
	// Verificar se slug existe
	if _, err := s.categoryRepo.FindBySlug(slug); err == nil {
		return nil, errors.New("categoria com nome similar já existe")
	}

	category := &domain.Category{
		ID:          uuid.New().String(),
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		PostCount:   0,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.categoryRepo.Create(category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *CategoryService) GetAll() []domain.Category {
	return s.categoryRepo.FindAll()
}

func (s *CategoryService) GetBySlug(slug string) (*domain.Category, error) {
	return s.categoryRepo.FindBySlug(slug)
}

func (s *CategoryService) Update(id string, req domain.UpdateCategoryRequest) (*domain.Category, error) {
	category, err := s.categoryRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		category.Name = req.Name
		category.Slug = utils.GenerateSlug(req.Name)
	}
	if req.Description != "" {
		category.Description = req.Description
	}

	category.UpdatedAt = time.Now()

	if err := s.categoryRepo.Update(category); err != nil {
		return nil, err
	}

	return category, nil
}

func (s *CategoryService) Delete(id string) error {
	return s.categoryRepo.Delete(id)
}
