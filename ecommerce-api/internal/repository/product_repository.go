package repository

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"ecommerce-api/internal/domain"
)

type ProductRepository struct {
	mu       sync.RWMutex
	file     string
	products map[string]*domain.Product
}

func NewProductRepository(dataDir string) (*ProductRepository, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	repo := &ProductRepository{
		file:     filepath.Join(dataDir, "products.json"),
		products: make(map[string]*domain.Product),
	}

	if err := repo.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return repo, nil
}

func (r *ProductRepository) Create(product *domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.products[product.ID] = product
	return r.save()
}

func (r *ProductRepository) FindByID(id string) (*domain.Product, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	product, exists := r.products[id]
	if !exists || !product.Active {
		return nil, errors.New("produto não encontrado")
	}

	p := *product
	return &p, nil
}

func (r *ProductRepository) FindAll(filter domain.ProductFilter) ([]domain.Product, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []domain.Product
	for _, p := range r.products {
		if !p.Active {
			continue
		}

		// Aplicar filtros
		if filter.Category != "" && p.Category != filter.Category {
			continue
		}
		if filter.MinPrice > 0 && p.Price < filter.MinPrice {
			continue
		}
		if filter.MaxPrice > 0 && p.Price > filter.MaxPrice {
			continue
		}
		if filter.InStock && p.Stock <= 0 {
			continue
		}
		if filter.Search != "" && !strings.Contains(strings.ToLower(p.Name), strings.ToLower(filter.Search)) {
			continue
		}

		result = append(result, *p)
	}

	total := len(result)

	// Paginação
	if filter.Page > 0 && filter.Limit > 0 {
		start := (filter.Page - 1) * filter.Limit
		end := start + filter.Limit
		
		if start > len(result) {
			return []domain.Product{}, total
		}
		if end > len(result) {
			end = len(result)
		}
		result = result[start:end]
	}

	return result, total
}

func (r *ProductRepository) Update(product *domain.Product) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.products[product.ID]; !exists {
		return errors.New("produto não encontrado")
	}

	r.products[product.ID] = product
	return r.save()
}

func (r *ProductRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if product, exists := r.products[id]; exists {
		product.Active = false
		return r.save()
	}
	return errors.New("produto não encontrado")
}

func (r *ProductRepository) UpdateStock(id string, quantity int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	product, exists := r.products[id]
	if !exists {
		return errors.New("produto não encontrado")
	}

	product.Stock += quantity
	if product.Stock < 0 {
		product.Stock = 0
	}

	return r.save()
}

func (r *ProductRepository) save() error {
	data, err := json.MarshalIndent(r.products, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.file, data, 0644)
}

func (r *ProductRepository) load() error {
	data, err := os.ReadFile(r.file)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &r.products)
}
