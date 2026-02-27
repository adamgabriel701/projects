package repository

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"ecommerce-api/internal/domain"
	"time"
)

type CartRepository struct {
	mu    sync.RWMutex
	file  string
	carts map[string]*domain.Cart // userID -> cart
}

func NewCartRepository(dataDir string) (*CartRepository, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	repo := &CartRepository{
		file:  filepath.Join(dataDir, "carts.json"),
		carts: make(map[string]*domain.Cart),
	}

	if err := repo.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return repo, nil
}

func (r *CartRepository) GetOrCreate(userID string) *domain.Cart {
	r.mu.Lock()
	defer r.mu.Unlock()

	if cart, exists := r.carts[userID]; exists {
		return cart
	}

	cart := &domain.Cart{
		ID:        generateID(),
		UserID:    userID,
		Items:     []domain.CartItem{},
		Total:     0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	r.carts[userID] = cart
	r.save()
	return cart
}

func (r *CartRepository) Update(cart *domain.Cart) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	cart.UpdatedAt = time.Now()
	r.carts[cart.UserID] = cart
	return r.save()
}

func (r *CartRepository) Clear(userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if cart, exists := r.carts[userID]; exists {
		cart.Items = []domain.CartItem{}
		cart.Total = 0
		cart.UpdatedAt = time.Now()
		return r.save()
	}
	return errors.New("carrinho não encontrado")
}

func (r *CartRepository) save() error {
	data, err := json.MarshalIndent(r.carts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.file, data, 0644)
}

func (r *CartRepository) load() error {
	data, err := os.ReadFile(r.file)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &r.carts)
}
