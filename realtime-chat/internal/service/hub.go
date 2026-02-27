package service

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"realtime-chat/internal/domain"
	"realtime-chat/internal/repository"
	"sync"
	"time"
)

type Hub struct {
	rooms      map[string]*domain.Room
	roomsMu    sync.RWMutex
	msgRepo    *repository.MessageRepository
	register   chan *domain.Client
	unregister chan *domain.Client
	broadcast  chan domain.Message
}

func NewHub(msgRepo *repository.MessageRepository) *Hub {
	return &Hub{
		rooms:      make(map[string]*domain.Room),
		msgRepo:    msgRepo,
		register:   make(chan *domain.Client),
		unregister: make(chan *domain.Client),
		broadcast:  make(chan domain.Message, 256),
	}
}

func (h *Hub) Run() {
	// Criar sala geral por padrão
	h.CreateRoom("general", "Geral", "Sala de bate-papo geral")

	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)

		case client := <-h.unregister:
			h.handleUnregister(client)

		case message := <-h.broadcast:
			h.handleBroadcast(message)
		}
	}
}

func (h *Hub) handleRegister(client *domain.Client) {
	room := h.GetRoom(client.RoomID)
	if room == nil {
		close(client.Send)
		return
	}

	room.AddClient(client)
	
	// Notificar entrada
	joinMsg := domain.Message{
		ID:        generateID(),
		RoomID:    client.RoomID,
		Username:  "Sistema",
		Content:   client.Username + " entrou na sala",
		Type:      "join",
		CreatedAt: time.Now(),
	}
	
	h.broadcast <- joinMsg
	
	log.Printf("👤 %s entrou na sala %s", client.Username, client.RoomID)
}

func (h *Hub) handleUnregister(client *domain.Client) {
	room := h.GetRoom(client.RoomID)
	if room == nil {
		return
	}

	// Verificar se cliente existe na sala
	clients := room.GetClients()
	found := false
	for _, c := range clients {
		if c.ID == client.ID {
			found = true
			break
		}
	}

	if found {
		room.RemoveClient(client)
		
		// Notificar saída
		leaveMsg := domain.Message{
			ID:        generateID(),
			RoomID:    client.RoomID,
			Username:  "Sistema",
			Content:   client.Username + " saiu da sala",
			Type:      "leave",
			CreatedAt: time.Now(),
		}
		
		h.broadcast <- leaveMsg
		
		log.Printf("👋 %s saiu da sala %s", client.Username, client.RoomID)
	}
}

func (h *Hub) handleBroadcast(message domain.Message) {
	// Persistir mensagem
	if message.Type == "text" {
		go h.msgRepo.Save(message.RoomID, message)
	}

	room := h.GetRoom(message.RoomID)
	if room == nil {
		return
	}

	clients := room.GetClients()
	for _, client := range clients {
		select {
		case client.Send <- message:
		default:
			// Cliente lento, remover
			room.RemoveClient(client)
		}
	}
}

func (h *Hub) GetRoom(roomID string) *domain.Room {
	h.roomsMu.RLock()
	defer h.roomsMu.RUnlock()
	return h.rooms[roomID]
}

func (h *Hub) CreateRoom(id, name, description string) *domain.Room {
	h.roomsMu.Lock()
	defer h.roomsMu.Unlock()

	if room, exists := h.rooms[id]; exists {
		return room
	}

	room := domain.NewRoom(id, name, description)
	h.rooms[id] = room
	
	// Iniciar goroutine para broadcast da sala
	go h.runRoomBroadcast(room)
	
	log.Printf("🏠 Sala criada: %s (%s)", name, id)
	return room
}

func (h *Hub) runRoomBroadcast(room *domain.Room) {
	for msg := range room.GetBroadcastChan() {
		h.broadcast <- msg
	}
}

func (h *Hub) GetRooms() []domain.RoomInfo {
	h.roomsMu.RLock()
	defer h.roomsMu.RUnlock()

	rooms := make([]domain.RoomInfo, 0, len(h.rooms))
	for _, room := range h.rooms {
		rooms = append(rooms, domain.RoomInfo{
			ID:          room.ID,
			Name:        room.Name,
			Description: room.Description,
			UserCount:   room.GetUserCount(),
		})
	}
	return rooms
}

func (h *Hub) GetRegisterChan() chan<- *domain.Client {
	return h.register
}

func (h *Hub) GetUnregisterChan() chan<- *domain.Client {
	return h.unregister
}

func (h *Hub) GetBroadcastChan() chan<- domain.Message {
	return h.broadcast
}

func generateID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}