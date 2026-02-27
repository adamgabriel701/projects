package repository

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"blog-api/internal/domain"
	"blog-api/internal/utils"
)

type PostRepository struct {
	mu       sync.RWMutex
	file     string
	posts    map[string]*domain.Post
	index    map[string][]string // tag -> []postIDs
	search   map[string][]string // palavra -> []postIDs (índice invertido simples)
}

func NewPostRepository(dataDir string) (*PostRepository, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	repo := &PostRepository{
		file:   filepath.Join(dataDir, "posts.json"),
		posts:  make(map[string]*domain.Post),
		index:  make(map[string][]string),
		search: make(map[string][]string),
	}

	if err := repo.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return repo, nil
}

func (r *PostRepository) Create(post *domain.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.posts[post.ID] = post
	r.updateIndex(post)

	return r.save()
}

func (r *PostRepository) FindByID(id string) (*domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	post, exists := r.posts[id]
	if !exists {
		return nil, errors.New("post não encontrado")
	}

	p := *post
	return &p, nil
}

func (r *PostRepository) FindBySlug(slug string) (*domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, post := range r.posts {
		if post.Slug == slug && post.Status == domain.StatusPublished {
			p := *post
			return &p, nil
		}
	}

	return nil, errors.New("post não encontrado")
}

func (r *PostRepository) FindAll(filter domain.PostFilter) ([]domain.Post, int) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var posts []domain.Post
	for _, p := range r.posts {
		// Aplicar filtros
		if filter.Status != "" && p.Status != filter.Status {
			continue
		}
		if filter.CategoryID != "" && p.CategoryID != filter.CategoryID {
			continue
		}
		if filter.AuthorID != "" && p.AuthorID != filter.AuthorID {
			continue
		}
		if filter.Featured && !p.Featured {
			continue
		}

		// Filtro por tag
		if filter.Tag != "" {
			hasTag := false
			for _, tag := range p.Tags {
				if tag.Slug == filter.Tag {
					hasTag = true
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		// Busca textual simples
		if filter.Search != "" {
			searchLower := strings.ToLower(filter.Search)
			content := strings.ToLower(p.Title + " " + p.Content + " " + p.Excerpt)
			if !strings.Contains(content, searchLower) {
				continue
			}
		}

		posts = append(posts, *p)
	}

	// Ordenar por data (mais recente primeiro)
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].CreatedAt.After(posts[j].CreatedAt)
	})

	total := len(posts)

	// Paginação
	if filter.Page > 0 && filter.Limit > 0 {
		start := (filter.Page - 1) * filter.Limit
		end := start + filter.Limit
		
		if start > len(posts) {
			return []domain.Post{}, total
		}
		if end > len(posts) {
			end = len(posts)
		}
		posts = posts[start:end]
	}

	return posts, total
}

func (r *PostRepository) Update(post *domain.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.posts[post.ID]; !exists {
		return errors.New("post não encontrado")
	}

	r.posts[post.ID] = post
	r.updateIndex(post)

	return r.save()
}

func (r *PostRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	post, exists := r.posts[id]
	if !exists {
		return errors.New("post não encontrado")
	}

	// Remover dos índices
	for _, tag := range post.Tags {
		r.removeFromIndex(tag.Slug, id)
	}

	delete(r.posts, id)
	return r.save()
}

func (r *PostRepository) IncrementViews(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if post, exists := r.posts[id]; exists {
		post.Views++
		return r.save()
	}
	return errors.New("post não encontrado")
}

func (r *PostRepository) ToggleLike(postID, userID string) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	post, exists := r.posts[postID]
	if !exists {
		return false, errors.New("post não encontrado")
	}

	// Verificar se já deu like
	for i, uid := range post.LikedBy {
		if uid == userID {
			// Remover like
			post.LikedBy = append(post.LikedBy[:i], post.LikedBy[i+1:]...)
			post.Likes--
			r.save()
			return false, nil
		}
	}

	// Adicionar like
	post.LikedBy = append(post.LikedBy, userID)
	post.Likes++
	r.save()
	return true, nil
}

func (r *PostRepository) updateIndex(post *domain.Post) {
	// Indexar tags
	for _, tag := range post.Tags {
		found := false
		for _, id := range r.index[tag.Slug] {
			if id == post.ID {
				found = true
				break
			}
		}
		if !found {
			r.index[tag.Slug] = append(r.index[tag.Slug], post.ID)
		}
	}

	// Indexar palavras para busca (simplificado)
	words := utils.ExtractKeywords(post.Title + " " + post.Content)
	for _, word := range words {
		found := false
		for _, id := range r.search[word] {
			if id == post.ID {
				found = true
				break
			}
		}
		if !found {
			r.search[word] = append(r.search[word], post.ID)
		}
	}
}

func (r *PostRepository) removeFromIndex(tagSlug, postID string) {
	if ids, exists := r.index[tagSlug]; exists {
		var newIDs []string
		for _, id := range ids {
			if id != postID {
				newIDs = append(newIDs, id)
			}
		}
		r.index[tagSlug] = newIDs
	}
}

func (r *PostRepository) save() error {
	data, err := json.MarshalIndent(r.posts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.file, data, 0644)
}

func (r *PostRepository) load() error {
	data, err := os.ReadFile(r.file)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &r.posts)
}
