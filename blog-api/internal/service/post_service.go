package service

import (
	"blog-api/internal/domain"
	"blog-api/internal/repository"
	"blog-api/internal/utils"
	"blog-api/pkg/markdown"
	"errors"
	"time"

	"github.com/google/uuid"
)

type PostService struct {
	postRepo     *repository.PostRepository
	categoryRepo *repository.CategoryRepository
	tagRepo      *repository.TagRepository
}

func NewPostService(postRepo *repository.PostRepository, categoryRepo *repository.CategoryRepository, tagRepo *repository.TagRepository) *PostService {
	return &PostService{
		postRepo:     postRepo,
		categoryRepo: categoryRepo,
		tagRepo:      tagRepo,
	}
}

func (s *PostService) Create(authorID string, req domain.CreatePostRequest) (*domain.Post, error) {
	// Validar categoria
	if _, err := s.categoryRepo.FindByID(req.CategoryID); err != nil {
		return nil, errors.New("categoria não encontrada")
	}

	// Gerar slug único
	slug := utils.GenerateSlug(req.Title)
	baseSlug := slug
	counter := 1
	for {
		if _, err := s.postRepo.FindBySlug(slug); err != nil {
			break // slug disponível
		}
		slug = baseSlug + "-" + string(rune('0'+counter))
		counter++
	}

	// Processar markdown para HTML
	htmlContent := markdown.ToHTML(req.Content)

	// Gerar excerpt
	excerpt := req.Content
	if len(excerpt) > 200 {
		excerpt = excerpt[:200] + "..."
	}

	// Processar tags
	var tags []domain.Tag
	for _, tagName := range req.Tags {
		tag, err := s.tagRepo.FindOrCreate(tagName)
		if err == nil {
			tags = append(tags, *tag)
			s.tagRepo.IncrementPostCount(tag.ID)
		}
	}

	// Definir status e data de publicação
	status := req.Status
	if status == "" {
		status = domain.StatusDraft
	}

	var publishedAt *time.Time
	if status == domain.StatusPublished {
		now := time.Now()
		publishedAt = &now
		s.categoryRepo.IncrementPostCount(req.CategoryID)
	}

	post := &domain.Post{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Slug:        slug,
		Excerpt:     excerpt,
		Content:     req.Content,
		HTMLContent: htmlContent,
		AuthorID:    authorID,
		CategoryID:  req.CategoryID,
		Tags:        tags,
		Status:      status,
		Featured:    req.Featured,
		Views:       0,
		Likes:       0,
		LikedBy:     []string{},
		CoverImage:  req.CoverImage,
		PublishedAt: publishedAt,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.postRepo.Create(post); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostService) GetBySlug(slug string, incrementViews bool) (*domain.Post, error) {
	post, err := s.postRepo.FindBySlug(slug)
	if err != nil {
		return nil, err
	}

	if incrementViews {
		s.postRepo.IncrementViews(post.ID)
		post.Views++
	}

	return post, nil
}

func (s *PostService) GetByID(id string) (*domain.Post, error) {
	return s.postRepo.FindByID(id)
}

func (s *PostService) List(filter domain.PostFilter) ([]domain.Post, int, error) {
	posts, total := s.postRepo.FindAll(filter)
	return posts, total, nil
}

func (s *PostService) Update(userID string, userRole domain.UserRole, postID string, req domain.UpdatePostRequest) (*domain.Post, error) {
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return nil, err
	}

	// Verificar permissão
	if post.AuthorID != userID && userRole != domain.RoleEditor && userRole != domain.RoleAdmin {
		return nil, errors.New("não autorizado")
	}

	// Atualizar campos
	if req.Title != "" {
		post.Title = req.Title
		post.Excerpt = req.Content
		if len(post.Excerpt) > 200 {
			post.Excerpt = post.Excerpt[:200] + "..."
		}
	}
	if req.Content != "" {
		post.Content = req.Content
		post.HTMLContent = markdown.ToHTML(req.Content)
	}
	if req.CategoryID != "" {
		post.CategoryID = req.CategoryID
	}
	if req.Tags != nil {
		var tags []domain.Tag
		for _, tagName := range req.Tags {
			tag, _ := s.tagRepo.FindOrCreate(tagName)
			tags = append(tags, *tag)
		}
		post.Tags = tags
	}
	if req.Featured != nil {
		post.Featured = *req.Featured
	}
	if req.CoverImage != "" {
		post.CoverImage = req.CoverImage
	}
	if req.Status != "" {
		oldStatus := post.Status
		post.Status = req.Status

		if oldStatus != domain.StatusPublished && req.Status == domain.StatusPublished {
			now := time.Now()
			post.PublishedAt = &now
		}
	}

	post.UpdatedAt = time.Now()

	if err := s.postRepo.Update(post); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostService) Delete(userID string, userRole domain.UserRole, postID string) error {
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return err
	}

	if post.AuthorID != userID && userRole != domain.RoleAdmin {
		return errors.New("não autorizado")
	}

	return s.postRepo.Delete(postID)
}

func (s *PostService) ToggleLike(postID, userID string) (bool, error) {
	return s.postRepo.ToggleLike(postID, userID)
}

func (s *PostService) GetFeatured() ([]domain.Post, error) {
	posts, _ := s.postRepo.FindAll(domain.PostFilter{
		Status:   domain.StatusPublished,
		Featured: true,
		Limit:    5,
	})
	return posts, nil
}
