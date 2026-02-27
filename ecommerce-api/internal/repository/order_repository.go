package repository

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"ecommerce-api/internal/domain"
	"github.com/google/uuid"
)

type OrderRepository struct {
	mu     sync.RWMutex
	file   string
	orders map[string]*domain.Order
	byUser map[string][]string // userID -> []orderIDs
}

func generateID() string {
	return uuid.New().String()
}

func NewOrderRepository(dataDir string) (*OrderRepository, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	repo := &OrderRepository{
		file:   filepath.Join(dataDir, "orders.json"),
		orders: make(map[string]*domain.Order),
		byUser: make(map[string][]string),
	}

	if err := repo.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return repo, nil
}

func (r *OrderRepository) Create(order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.orders[order.ID] = order
	r.byUser[order.UserID] = append(r.byUser[order.UserID], order.ID)

	return r.save()
}

func (r *OrderRepository) FindByID(id string) (*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	order, exists := r.orders[id]
	if !exists {
		return nil, errors.New("pedido não encontrado")
	}

	o := *order
	return &o, nil
}

func (r *OrderRepository) FindByUser(userID string) []domain.Order {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var orders []domain.Order
	for _, id := range r.byUser[userID] {
		if order, exists := r.orders[id]; exists {
			orders = append(orders, *order)
		}
	}

	// Ordenar por data decrescente
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].CreatedAt.After(orders[j].CreatedAt)
	})

	return orders
}

func (r *OrderRepository) FindAll(status domain.OrderStatus, page, limit int) ([]domain.Order, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var orders []domain.Order
	for _, order := range r.orders {
		if status == "" || order.Status == status {
			orders = append(orders, *order)
		}
	}

	total := len(orders)

	// Ordenar por data decrescente
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].CreatedAt.After(orders[j].CreatedAt)
	})

	// Paginação
	if page > 0 && limit > 0 {
		start := (page - 1) * limit
		end := start + limit
		
		if start > len(orders) {
			return []domain.Order{}, total
		}
		if end > len(orders) {
			end = len(orders)
		}
		orders = orders[start:end]
	}

	return orders, total
}

func (r *OrderRepository) Update(order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.orders[order.ID]; !exists {
		return errors.New("pedido não encontrado")
	}

	r.orders[order.ID] = order
	return r.save()
}

func (r *OrderRepository) save() error {
	data, err := json.MarshalIndent(r.orders, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.file, data, 0644)
}

func (r *OrderRepository) load() error {
	data, err := os.ReadFile(r.file)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &r.orders)
}
