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

type CommentRepository struct {
	mu       sync.RWMutex
	file     string
	comments map[string]*domain.Comment
	byPost   map[string][]string // postID -> []commentIDs
}

func NewCommentRepository(dataDir string) (*CommentRepository, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	repo := &CommentRepository{
		file:     filepath.Join(dataDir, "comments.json"),
		comments: make(map[string]*domain.Comment),
		byPost:   make(map[string][]string),
	}

	if err := repo.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return repo, nil
}

func (r *CommentRepository) Create(comment *domain.Comment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.comments[comment.ID] = comment
	r.byPost[comment.PostID] = append(r.byPost[comment.PostID], comment.ID)

	return r.save()
}

func (r *CommentRepository) FindByPost(postID string, status domain.CommentStatus) ([]domain.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var comments []domain.Comment
	for _, id := range r.byPost[postID] {
		if comment, exists := r.comments[id]; exists {
			if status == "" || comment.Status == status {
				comments = append(comments, *comment)
			}
		}
	}

	// Ordenar por data
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.Before(comments[j].CreatedAt)
	})

	// Organizar em threads (comentários pai com replies)
	return r.organizeThreads(comments), nil
}

func (r *CommentRepository) organizeThreads(comments []domain.Comment) []domain.Comment {
	commentMap := make(map[string]*domain.Comment)
	var roots []domain.Comment

	for i := range comments {
		commentMap[comments[i].ID] = &comments[i]
	}

	for i := range comments {
		if comments[i].ParentID != "" {
			if parent, exists := commentMap[comments[i].ParentID]; exists {
				parent.Replies = append(parent.Replies, comments[i])
			}
		} else {
			roots = append(roots, comments[i])
		}
	}

	return roots
}

func (r *CommentRepository) FindByID(id string) (*domain.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	comment, exists := r.comments[id]
	if !exists {
		return nil, errors.New("comentário não encontrado")
	}

	c := *comment
	return &c, nil
}

func (r *CommentRepository) Update(comment *domain.Comment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.comments[comment.ID]; !exists {
		return errors.New("comentário não encontrado")
	}

	r.comments[comment.ID] = comment
	return r.save()
}

func (r *CommentRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.comments[id]; !exists {
		return errors.New("comentário não encontrado")
	}

	delete(r.comments, id)
	// Remover do índice byPost seria necessário em produção
	
	return r.save()
}

func (r *CommentRepository) CountByPost(postID string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.byPost[postID])
}

func (r *CommentRepository) save() error {
	data, err := json.MarshalIndent(r.comments, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.file, data, 0644)
}

func (r *CommentRepository) load() error {
	data, err := os.ReadFile(r.file)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &r.comments)
}
