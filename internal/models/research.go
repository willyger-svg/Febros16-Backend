package models

import (
	"time"
)

type ResearchProject struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Abstract    *string    `json:"abstract"`
	ContentBody string     `json:"content_body"`
	AuthorID    *string    `json:"author_id"`
	Status      string     `json:"status"`
	StartDate   *time.Time `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	CreatedAt   *time.Time `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}
