package repository

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"blog-api/internal/domain"
)

type UserRepository struct {
	mu       sync.RWMutex
	file     string
	users    map[string]*domain.User // id -> user
	byEmail  map[string]string       // email -> id
	byUsername map[string]string     // username -> id
}

func NewUserRepository(dataDir string) (*UserRepository, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	repo := &UserRepository{
		file:       filepath.Join(dataDir, "users.json"),
		users:      make(map[string]*domain.User),
		byEmail:    make(map[string]string),
		byUsername: make(map[string]string),
	}

	if err := repo.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return repo, nil
}

func (r *UserRepository) Create(user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byEmail[user.Email]; exists {
		return errors.New("email já cadastrado")
	}
	if _, exists := r.byUsername[strings.ToLower(user.Username)]; exists {
		return errors.New("username já existe")
	}

	r.users[user.ID] = user
	r.byEmail[user.Email] = user.ID
	r.byUsername[strings.ToLower(user.Username)] = user.ID

	return r.save()
}

func (r *UserRepository) FindByID(id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, errors.New("usuário não encontrado")
	}

	u := *user
	return &u, nil
}

func (r *UserRepository) FindByEmail(email string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.byEmail[email]
	if !exists {
		return nil, errors.New("usuário não encontrado")
	}

	u := *r.users[id]
	return &u, nil
}

func (r *UserRepository) FindByUsername(username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.byUsername[strings.ToLower(username)]
	if !exists {
		return nil, errors.New("usuário não encontrado")
	}

	u := *r.users[id]
	return &u, nil
}

func (r *UserRepository) Update(user *domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; !exists {
		return errors.New("usuário não encontrado")
	}

	r.users[user.ID] = user
	return r.save()
}

func (r *UserRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id]
	if !exists {
		return errors.New("usuário não encontrado")
	}

	delete(r.users, id)
	delete(r.byEmail, user.Email)
	delete(r.byUsername, strings.ToLower(user.Username))

	return r.save()
}

func (r *UserRepository) List(page, limit int) ([]domain.User, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var users []domain.User
	for _, u := range r.users {
		users = append(users, *u)
	}

	total := len(users)

	// Paginação simples
	if page > 0 && limit > 0 {
		start := (page - 1) * limit
		end := start + limit
		
		if start > len(users) {
			return []domain.User{}, total
		}
		if end > len(users) {
			end = len(users)
		}
		users = users[start:end]
	}

	return users, total
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
