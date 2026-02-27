package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"realtime-chat/internal/domain"
)

type MessageRepository struct {
	mu       sync.RWMutex
	dataDir  string
	messages map[string][]domain.Message // roomID -> messages
}

func NewMessageRepository(dataDir string) (*MessageRepository, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	repo := &MessageRepository{
		dataDir:  dataDir,
		messages: make(map[string][]domain.Message),
	}

	// Carregar mensagens existentes
	files, err := os.ReadDir(dataDir)
	if err == nil {
		for _, file := range files {
			if filepath.Ext(file.Name()) == ".json" {
				roomID := file.Name()[:len(file.Name())-5]
				repo.loadRoomMessages(roomID)
			}
		}
	}

	return repo, nil
}

func (r *MessageRepository) Save(roomID string, msg domain.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.messages[roomID] = append(r.messages[roomID], msg)
	
	// Manter apenas últimas 100 mensagens na memória
	if len(r.messages[roomID]) > 100 {
		r.messages[roomID] = r.messages[roomID][len(r.messages[roomID])-100:]
	}

	return r.persist(roomID)
}

func (r *MessageRepository) GetRecent(roomID string, limit int) ([]domain.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	msgs := r.messages[roomID]
	if len(msgs) == 0 {
		// Tentar carregar do disco
		r.mu.RUnlock()
		r.loadRoomMessages(roomID)
		r.mu.RLock()
		msgs = r.messages[roomID]
	}

	// Retornar as mais recentes
	if len(msgs) <= limit {
		result := make([]domain.Message, len(msgs))
		copy(result, msgs)
		return result, nil
	}

	result := make([]domain.Message, limit)
	copy(result, msgs[len(msgs)-limit:])
	
	// Ordenar por data
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})
	
	return result, nil
}

func (r *MessageRepository) persist(roomID string) error {
	filePath := filepath.Join(r.dataDir, roomID+".json")
	data, err := json.MarshalIndent(r.messages[roomID], "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, data, 0644)
}

func (r *MessageRepository) loadRoomMessages(roomID string) error {
	filePath := filepath.Join(r.dataDir, roomID+".json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var msgs []domain.Message
	if err := json.Unmarshal(data, &msgs); err != nil {
		return err
	}

	r.messages[roomID] = msgs
	return nil
}

