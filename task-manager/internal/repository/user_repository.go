package repository

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"task-manager/internal/domain"
)

type UserRepository interface {
	Create(user *domain.User) error
	FindByEmail(email string) (*domain.User, error)
	FindByID(id string) (*domain.User, error)
}

type JSONUserRepository struct {
	mu       sync.RWMutex
	filePath string
	users    map[string]*domain.User
}

func NewUserRepository(dataDir string) (*JSONUserRepository, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	repo := &JSONUserRepository{
		filePath: filepath.Join(dataDir, "users.json"),
		users:    make(map[string]*domain.User),
	}

	if err := repo.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return repo, nil
}

func (r *JSONUserRepository) Create(user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.Email]; exists {
		return errors.New("email já cadastrado")
	}

	r.users[user.Email] = user
	r.users[user.ID] = user // indexar por ID também
	
	return r.save()
}

func (r *JSONUserRepository) FindByEmail(email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[email]
	if !exists {
		return nil, errors.New("usuário não encontrado")
	}
	
	// Retornar cópia para evitar mutação externa
	userCopy := *user
	return &userCopy, nil
}

func (r *JSONUserRepository) FindByID(id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, errors.New("usuário não encontrado")
	}
	
	userCopy := *user
	return &userCopy, nil
}

func (r *JSONUserRepository) save() error {
	data, err := json.MarshalIndent(r.users, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.filePath, data, 0644)
}

func (r *JSONUserRepository) load() error {
	data, err := os.ReadFile(r.filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &r.users)
}
