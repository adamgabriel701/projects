package repository

import (
	"errors"
	"sync"
	"url-shortener/internal/models"
)

type MemoryStore struct {
	mu   sync.RWMutex
	urls map[string]*models.URL // shortCode -> URL
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		urls: make(map[string]*models.URL),
	}
}

func (s *MemoryStore) Save(url *models.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	s.urls[url.ShortCode] = url
	return nil
}

func (s *MemoryStore) FindByShortCode(code string) (*models.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	url, exists := s.urls[code]
	if !exists {
		return nil, errors.New("URL não encontrada")
	}
	
	return url, nil
}

func (s *MemoryStore) IncrementClicks(code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	url, exists := s.urls[code]
	if !exists {
		return errors.New("URL não encontrada")
	}
	
	url.Clicks++
	return nil
}

func (s *MemoryStore) GetAll() []*models.URL {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make([]*models.URL, 0, len(s.urls))
	for _, url := range s.urls {
		result = append(result, url)
	}
	
	return result
}
