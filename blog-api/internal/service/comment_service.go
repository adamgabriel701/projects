package service

import (
	"errors"
	"time"
	"blog-api/internal/domain"
	"blog-api/internal/repository"

	"github.com/google/uuid"
)

type CommentService struct {
	commentRepo *repository.CommentRepository
	postRepo    *repository.PostRepository
}

func NewCommentService(commentRepo *repository.CommentRepository, postRepo *repository.PostRepository) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
	}
}

func (s *CommentService) Create(userID string, req domain.CreateCommentRequest) (*domain.Comment, error) {
	// Verificar se post existe e está publicado
	post, err := s.postRepo.FindByID(req.PostID)
	if err != nil {
		return nil, errors.New("post não encontrado")
	}

	if post.Status != domain.StatusPublished {
		return nil, errors.New("não é possível comentar em posts não publicados")
	}

	comment := &domain.Comment{
		ID:         uuid.New().String(),
		PostID:     req.PostID,
		AuthorID:   userID,
		ParentID:   req.ParentID,
		AuthorName: req.AuthorName,
		Email:      req.Email,
		Content:    req.Content,
		Status:     domain.CommentStatusPending, // requer moderação
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if userID != "" {
		// Usuário logado - aprovar automaticamente
		comment.Status = domain.CommentStatusApproved
	}

	if err := s.commentRepo.Create(comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *CommentService) GetByPost(postID string, includePending bool) ([]domain.Comment, error) {
	status := domain.CommentStatusApproved
	if includePending {
		status = ""
	}
	return s.commentRepo.FindByPost(postID, status)
}

func (s *CommentService) Moderate(userID string, userRole domain.UserRole, commentID string, status domain.CommentStatus) error {
	if userRole != domain.RoleEditor && userRole != domain.RoleAdmin {
		return errors.New("não autorizado")
	}

	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		return err
	}

	comment.Status = status
	comment.UpdatedAt = time.Now()

	return s.commentRepo.Update(comment)
}

func (s *CommentService) Delete(userID string, userRole domain.UserRole, commentID string) error {
	comment, err := s.commentRepo.FindByID(commentID)
	if err != nil {
		return err
	}

	// Autor, editor ou admin pode deletar
	if comment.AuthorID != userID && userRole != domain.RoleEditor && userRole != domain.RoleAdmin {
		return errors.New("não autorizado")
	}

	return s.commentRepo.Delete(commentID)
}
