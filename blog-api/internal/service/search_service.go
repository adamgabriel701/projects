package service

import (
	"blog-api/internal/domain"
	"blog-api/internal/repository"
	"strings"
)

type SearchService struct {
	postRepo *repository.PostRepository
}

func NewSearchService(postRepo *repository.PostRepository) *SearchService {
	return &SearchService{postRepo: postRepo}
}

type SearchResult struct {
	Posts      []domain.Post `json:"posts"`
	Total      int           `json:"total"`
	Query      string        `json:"query"`
	Suggestion string        `json:"suggestion,omitempty"`
}

func (s *SearchService) Search(query string, page, limit int) (*SearchResult, error) {
	// Normalizar query
	query = strings.ToLower(strings.TrimSpace(query))

	if query == "" {
		return &SearchResult{
			Posts: []domain.Post{},
			Total: 0,
			Query: query,
		}, nil
	}

	posts, total := s.postRepo.FindAll(domain.PostFilter{
		Status: domain.StatusPublished,
		Search: query,
		Page:   page,
		Limit:  limit,
	})

	// Sugerir correção simples (placeholder para algo mais sofisticado)
	suggestion := ""
	if total == 0 {
		suggestion = s.didYouMean(query)
	}

	return &SearchResult{
		Posts:      posts,
		Total:      total,
		Query:      query,
		Suggestion: suggestion,
	}, nil
}

func (s *SearchService) didYouMean(query string) string {
	// Implementação simples - em produção usar algoritmo de distância de edição
	commonTypos := map[string]string{
		"javascrit": "javascript",
		"pyton":     "python",
		"goland":    "golang",
	}

	if correction, exists := commonTypos[query]; exists {
		return correction
	}
	return ""
}

func (s *SearchService) GetRelatedPosts(postID string, limit int) ([]domain.Post, error) {
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return nil, err
	}

	// Buscar posts da mesma categoria ou com tags similares
	var categoryID string
	if len(post.Tags) > 0 {
		// Usar primeira tag para busca relacionada
		// Simplificado - em produção usar mais critérios
		categoryID = post.CategoryID
	}

	posts, _ := s.postRepo.FindAll(domain.PostFilter{
		Status:     domain.StatusPublished,
		CategoryID: categoryID,
		Limit:      limit + 1, // +1 para remover o próprio post
	})

	// Remover o post atual da lista
	var related []domain.Post
	for _, p := range posts {
		if p.ID != postID {
			related = append(related, p)
			if len(related) >= limit {
				break
			}
		}
	}

	return related, nil
}
