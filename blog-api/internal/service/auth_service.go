package service

import (
	"errors"
	"time"
	"blog-api/internal/config"
	"blog-api/internal/domain"
	"blog-api/internal/repository"
	"blog-api/internal/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repository.UserRepository
	config   *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		config:   cfg,
	}
}

func (s *AuthService) Register(req domain.RegisterRequest) (*domain.User, string, error) {
	// Verificar username único
	if _, err := s.userRepo.FindByUsername(req.Username); err == nil {
		return nil, "", errors.New("username já existe")
	}

	// Verificar email único
	if _, err := s.userRepo.FindByEmail(req.Email); err == nil {
		return nil, "", errors.New("email já cadastrado")
	}

	// Hash senha
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", err
	}

	user := &domain.User{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Password:  string(hashedPassword),
		Username:  req.Username,
		Name:      req.Name,
		Role:      domain.RoleReader,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", err
	}

	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role), s.config.JWTSecret, s.config.TokenExpiry)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) Login(req domain.LoginRequest) (*domain.User, string, error) {
	var user *domain.User
	var err error

	// Login por email ou username
	if req.Email != "" {
		user, err = s.userRepo.FindByEmail(req.Email)
	} else if req.Username != "" {
		user, err = s.userRepo.FindByUsername(req.Username)
	} else {
		return nil, "", errors.New("email ou username obrigatório")
	}

	if err != nil {
		return nil, "", errors.New("credenciais inválidas")
	}

	if !user.Active {
		return nil, "", errors.New("conta desativada")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, "", errors.New("credenciais inválidas")
	}

	token, err := utils.GenerateToken(user.ID, user.Email, string(user.Role), s.config.JWTSecret, s.config.TokenExpiry)
	if err != nil {
		return nil, "", err
	}

	return user, token, nil
}

func (s *AuthService) GetProfile(userID string) (*domain.User, error) {
	return s.userRepo.FindByID(userID)
}

func (s *AuthService) UpdateProfile(userID string, req domain.UpdateProfileRequest) (*domain.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Bio != "" {
		user.Bio = req.Bio
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) ChangeUserRole(adminID, targetUserID string, newRole domain.UserRole) error {
	admin, err := s.userRepo.FindByID(adminID)
	if err != nil || admin.Role != domain.RoleAdmin {
		return errors.New("não autorizado")
	}

	user, err := s.userRepo.FindByID(targetUserID)
	if err != nil {
		return err
	}

	user.Role = newRole
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(user)
}
