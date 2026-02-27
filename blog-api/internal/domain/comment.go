package domain

import "time"

type CommentStatus string

const (
	CommentStatusPending  CommentStatus = "pending"
	CommentStatusApproved CommentStatus = "approved"
	CommentStatusRejected CommentStatus = "rejected"
)

type Comment struct {
	ID        string        `json:"id"`
	PostID    string        `json:"post_id"`
	AuthorID  string        `json:"author_id,omitempty"`
	Author    *User         `json:"author,omitempty"`
	ParentID  string        `json:"parent_id,omitempty"` // para respostas
	AuthorName string       `json:"author_name"`         // para anônimos
	Email     string        `json:"email,omitempty"`
	Content   string        `json:"content"`
	Status    CommentStatus `json:"status"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
	Replies   []Comment     `json:"replies,omitempty"`
}

type CreateCommentRequest struct {
	PostID     string `json:"post_id" validate:"required"`
	ParentID   string `json:"parent_id,omitempty"`
	AuthorName string `json:"author_name"`
	Email      string `json:"email"`
	Content    string `json:"content" validate:"required,min=5"`
}

type UpdateCommentStatusRequest struct {
	Status CommentStatus `json:"status" validate:"required"`
}
