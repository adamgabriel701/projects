package repository

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"blog-api/internal/domain"
)

type CategoryRepository struct {
	mu         sync.RWMutex
	file       string
	categories map[string]*domain.Category
	bySlug     map[string]string
}

func NewCategoryRepository(dataDir string) (*CategoryRepository, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	repo := &CategoryRepository{
		file:       filepath.Join(dataDir, "categories.json"),
		categories: make(map[string]*domain.Category),
		bySlug:     make(map[string]string),
	}

	if err := repo.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return repo, nil
}

func (r *CategoryRepository) Create(category *domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.bySlug[category.Slug]; exists {
		return errors.New("slug já existe")
	}

	r.categories[category.ID] = category
	r.bySlug[category.Slug] = category.ID

	return r.save()
}

func (r *CategoryRepository) FindByID(id string) (*domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	category, exists := r.categories[id]
	if !exists {
		return nil, errors.New("categoria não encontrada")
	}

	c := *category
	return &c, nil
}

func (r *CategoryRepository) FindBySlug(slug string) (*domain.Category, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.bySlug[slug]
	if !exists {
		return nil, errors.New("categoria não encontrada")
	}

	c := *r.categories[id]
	return &c, nil
}

func (r *CategoryRepository) FindAll() []domain.Category {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var categories []domain.Category
	for _, c := range r.categories {
		categories = append(categories, *c)
	}

	// Ordenar por nome
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].Name < categories[j].Name
	})

	return categories
}

func (r *CategoryRepository) Update(category *domain.Category) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.categories[category.ID]; !exists {
		return errors.New("categoria não encontrada")
	}

	// Atualizar slug no índice se mudou
	oldCategory := r.categories[category.ID]
	if oldCategory.Slug != category.Slug {
		delete(r.bySlug, oldCategory.Slug)
		r.bySlug[category.Slug] = category.ID
	}

	r.categories[category.ID] = category
	return r.save()
}

func (r *CategoryRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	category, exists := r.categories[id]
	if !exists {
		return errors.New("categoria não encontrada")
	}

	delete(r.categories, id)
	delete(r.bySlug, category.Slug)

	return r.save()
}

func (r *CategoryRepository) IncrementPostCount(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if category, exists := r.categories[id]; exists {
		category.PostCount++
		return r.save()
	}
	return errors.New("categoria não encontrada")
}

func (r *CategoryRepository) save() error {
	data, err := json.MarshalIndent(r.categories, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.file, data, 0644)
}

func (r *CategoryRepository) load() error {
	data, err := os.ReadFile(r.file)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &r.categories)
}
