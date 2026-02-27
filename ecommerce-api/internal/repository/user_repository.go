package repository

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"ecommerce-api/internal/domain"
)

type UserRepository struct {
	mu    sync.RWMutex
	file  string
	users map[string]*domain.User // email -> user
	byID  map[string]string       // id -> email
}

func NewUserRepository(dataDir string) (*UserRepository, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	repo := &UserRepository{
		file:  filepath.Join(dataDir, "users.json"),
		users: make(map[string]*domain.User),
		byID:  make(map[string]string),
	}

	if err := repo.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return repo, nil
}

func (r *UserRepository) Create(user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.Email]; exists {
		return errors.New("email já cadastrado")
	}

	r.users[user.Email] = user
	r.byID[user.ID] = user.Email

	return r.save()
}

func (r *UserRepository) FindByEmail(email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[email]
	if !exists {
		return nil, errors.New("usuário não encontrado")
	}

	// Retornar cópia
	u := *user
	return &u, nil
}

func (r *UserRepository) FindByID(id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	email, exists := r.byID[id]
	if !exists {
		return nil, errors.New("usuário não encontrado")
	}

	u := *r.users[email]
	return &u, nil
}

func (r *UserRepository) Update(user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	email, exists := r.byID[user.ID]
	if !exists {
		return errors.New("usuário não encontrado")
	}

	delete(r.users, email)
	r.users[user.Email] = user
	r.byID[user.ID] = user.Email

	return r.save()
}

func (r *UserRepository) save() error {
	data, err := json.MarshalIndent(r.users, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.file, data, 0644)
}

func (r *UserRepository) load() error {
	data, err := os.ReadFile(r.file)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &r.users)
}
