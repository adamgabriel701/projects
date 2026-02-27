package domain

import "time"

type Tag struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	PostCount int       `json:"post_count"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateTagRequest struct {
	Name string `json:"name" validate:"required"`
}
