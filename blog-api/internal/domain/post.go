package domain

import "time"

type PostStatus string

const (
	StatusDraft     PostStatus = "draft"
	StatusPublished PostStatus = "published"
	StatusArchived  PostStatus = "archived"
)

type Post struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Excerpt     string     `json:"excerpt"`
	Content     string     `json:"content"`
	HTMLContent string     `json:"html_content"`
	AuthorID    string     `json:"author_id"`
	Author      *User      `json:"author,omitempty"`
	CategoryID  string     `json:"category_id"`
	Category    *Category  `json:"category,omitempty"`
	Tags        []Tag      `json:"tags"`
	Status      PostStatus `json:"status"`
	Featured    bool       `json:"featured"`
	Views       int        `json:"views"`
	Likes       int        `json:"likes"`
	LikedBy     []string   `json:"-"` // user IDs que deram like
	CoverImage  string     `json:"cover_image"`
	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreatePostRequest struct {
	Title      string   `json:"title" validate:"required"`
	Content    string   `json:"content" validate:"required"`
	CategoryID string   `json:"category_id" validate:"required"`
	Tags       []string `json:"tags"`
	Featured   bool     `json:"featured"`
	CoverImage string   `json:"cover_image"`
	Status     PostStatus `json:"status"`
}

type UpdatePostRequest struct {
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	CategoryID string     `json:"category_id"`
	Tags       []string   `json:"tags"`
	Featured   *bool      `json:"featured"`
	CoverImage string     `json:"cover_image"`
	Status     PostStatus `json:"status"`
}

type PostFilter struct {
	Status     PostStatus
	CategoryID string
	Tag        string
	AuthorID   string
	Featured   bool
	Search     string
	Page       int
	Limit      int
}

type PostListResponse struct {
	Posts      []Post `json:"posts"`
	Total      int    `json:"total"`
	Page       int    `json:"page"`
	TotalPages int    `json:"total_pages"`
}
