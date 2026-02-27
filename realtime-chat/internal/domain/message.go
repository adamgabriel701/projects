package domain

import "time"

type Message struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"room_id"`
	Username  string    `json:"username"`
	Content   string    `json:"content"`
	Type      string    `json:"type"` // "text", "join", "leave", "system"
	CreatedAt time.Time `json:"created_at"`
}

type MessageRequest struct {
	Username string `json:"username"`
	Content  string `json:"content"`
	RoomID   string `json:"room_id"`
}

type MessageResponse struct {
	Messages []Message `json:"messages"`
	RoomID   string    `json:"room_id"`
}
