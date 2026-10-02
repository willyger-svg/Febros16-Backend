package models

import "time"

// Article represents the articles table
type Article struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	ContentBody string     `json:"content_body"`
	AuthorID    *string    `json:"author_id"` // can be null if author deleted
	CategoryID  *string    `json:"category_id"`
	Status      string     `json:"status"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}
