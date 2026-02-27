package repository

import (
	"blog-api/internal/domain"
	"blog-api/internal/utils"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

type TagRepository struct {
	mu     sync.RWMutex
	file   string
	tags   map[string]*domain.Tag
	bySlug map[string]string
	byName map[string]string
}

func NewTagRepository(dataDir string) (*TagRepository, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	repo := &TagRepository{
		file:   filepath.Join(dataDir, "tags.json"),
		tags:   make(map[string]*domain.Tag),
		bySlug: make(map[string]string),
		byName: make(map[string]string),
	}

	if err := repo.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return repo, nil
}

func (r *TagRepository) FindOrCreate(name string) (*domain.Tag, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	slug := utils.GenerateSlug(name)

	// Verificar se existe
	if id, exists := r.bySlug[slug]; exists {
		return r.tags[id], nil
	}

	// Criar nova
	tag := &domain.Tag{
		ID:        uuid.New().String(),
		Name:      name,
		Slug:      slug,
		PostCount: 0,
		CreatedAt: time.Now(),
	}

	r.tags[tag.ID] = tag
	r.bySlug[tag.Slug] = tag.ID
	r.byName[strings.ToLower(name)] = tag.ID

	r.save()
	return tag, nil
}

func (r *TagRepository) FindBySlug(slug string) (*domain.Tag, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, exists := r.bySlug[slug]
	if !exists {
		return nil, errors.New("tag não encontrada")
	}

	t := *r.tags[id]
	return &t, nil
}

func (r *TagRepository) FindAll() []domain.Tag {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var tags []domain.Tag
	for _, t := range r.tags {
		tags = append(tags, *t)
	}

	// Ordenar por post count (mais usadas primeiro)
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].PostCount > tags[j].PostCount
	})

	return tags
}

func (r *TagRepository) IncrementPostCount(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if tag, exists := r.tags[id]; exists {
		tag.PostCount++
		return r.save()
	}
	return errors.New("tag não encontrada")
}

func (r *TagRepository) save() error {
	data, err := json.MarshalIndent(r.tags, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.file, data, 0644)
}

func (r *TagRepository) load() error {
	data, err := os.ReadFile(r.file)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &r.tags)
}
