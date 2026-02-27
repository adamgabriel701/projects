package domain

import (
	"sync"
	"time"
)

type RoomStatus string

const (
	RoomStatusActive   RoomStatus = "active"
	RoomStatusArchived RoomStatus = "archived"
)

type Room struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	CreatedAt   string           `json:"created_at"`
	clients     map[*Client]bool // não exportado - runtime only
	mu          sync.RWMutex     // não exportado
	broadcast   chan Message     // não exportado
}

type Client struct {
	ID       string
	Username string
	RoomID   string
	Send     chan Message
}

type RoomInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	UserCount   int    `json:"user_count"`
}

func NewRoom(id, name, description string) *Room {
	return &Room{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now().Format(time.RFC3339),
		clients:     make(map[*Client]bool),
		broadcast:   make(chan Message, 256),
	}
}

func (r *Room) AddClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.clients[client] = true
}

func (r *Room) RemoveClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.clients, client)
	close(client.Send)
}

func (r *Room) GetClients() []*Client {
	r.mu.RLock()
	defer r.mu.RUnlock()
	clients := make([]*Client, 0, len(r.clients))
	for client := range r.clients {
		clients = append(clients, client)
	}
	return clients
}

func (r *Room) GetUserCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.clients)
}

func (r *Room) Broadcast() chan<- Message {
	return r.broadcast
}

func (r *Room) GetBroadcastChan() <-chan Message {
	return r.broadcast
}
