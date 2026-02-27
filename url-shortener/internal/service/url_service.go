package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"
	"url-shortener/internal/models"
	"url-shortener/internal/repository"
)

type URLService struct {
	store  *repository.MemoryStore
	domain string
}

func NewURLService(store *repository.MemoryStore, domain string) *URLService {
	return &URLService{
		store:  store,
		domain: domain,
	}
}

func (s *URLService) CreateShortURL(originalURL string) (*models.CreateURLResponse, error) {
	// Gerar código curto único
	shortCode := generateShortCode(6)
	
	// Verificar se já existe (colisão)
	for {
		_, err := s.store.FindByShortCode(shortCode)
		if err != nil {
			break // Não existe, podemos usar
		}
		shortCode = generateShortCode(6)
	}
	
	url := &models.URL{
		ID:          generateShortCode(8),
		OriginalURL: originalURL,
		ShortCode:   shortCode,
		CreatedAt:   time.Now(),
		Clicks:      0,
	}
	
	if err := s.store.Save(url); err != nil {
		return nil, err
	}
	
	return &models.CreateURLResponse{
		ShortURL:    fmt.Sprintf("%s/%s", s.domain, shortCode),
		OriginalURL: originalURL,
	}, nil
}

func (s *URLService) GetOriginalURL(shortCode string) (string, error) {
	url, err := s.store.FindByShortCode(shortCode)
	if err != nil {
		return "", err
	}
	
	// Incrementar contador de cliques (assíncrono)
	go s.store.IncrementClicks(shortCode)
	
	return url.OriginalURL, nil
}

func (s *URLService) GetURLStats(shortCode string) (*models.URL, error) {
	return s.store.FindByShortCode(shortCode)
}

func generateShortCode(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}
