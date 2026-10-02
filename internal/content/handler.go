package content

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"febros16-backend/config"
	"febros16-backend/internal/middleware"
	"febros16-backend/internal/models"
)

type CreateArticleInput struct {
	Title       string  `json:"title"`
	Slug        string  `json:"slug"`
	ContentBody string  `json:"content_body"`
	CategoryID  *string `json:"category_id"`
}

// GetArticles - Public API to list articles
func GetArticles(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"success": false, "error": {"code": "METHOD_NOT_ALLOWED", "message": "Njia hairuhusiwi"}}`))
		return
	}

	query := `
		SELECT id, title, slug, content_body, author_id, category_id, status, created_at, updated_at
		FROM articles
		ORDER BY created_at DESC
	`
	rows, err := config.DB.Query(query)
	if err != nil {
		log.Printf("[CONTENT DB ERROR] Kosa kuvuta makala: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "error": {"code": "SERVER_ERROR", "message": "Imeshindwa kuvuta makala"}}`))
		return
	}
	defer rows.Close()

	var articles []models.Article
	for rows.Next() {
		var a models.Article
		var authorID, categoryID *string
		var createdAt, updatedAt *time.Time

		err := rows.Scan(&a.ID, &a.Title, &a.Slug, &a.ContentBody, &authorID, &categoryID, &a.Status, &createdAt, &updatedAt)
		if err != nil {
			log.Printf("[CONTENT DB ERROR] Scan error: %v", err)
			continue // ruka makala iliyofeli badala ya ku-crash
		}
		
		a.AuthorID = authorID
		a.CategoryID = categoryID
		a.CreatedAt = createdAt
		a.UpdatedAt = updatedAt

		articles = append(articles, a)
	}

	// Kama hakuna makala, rudisha list tupu badala ya null
	if articles == nil {
		articles = []models.Article{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"articles": articles,
		},
	})
}

// CreateArticle - Protected API
func CreateArticle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"success": false, "error": {"code": "METHOD_NOT_ALLOWED", "message": "Njia hairuhusiwi"}}`))
		return
	}

	// Pata author ID kutoka kwa JWT
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Imeshindwa kuthibitisha mtumiaji"}}`))
		return
	}

	var input CreateArticleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "error": {"code": "BAD_REQUEST", "message": "Taarifa hazisomeki"}}`))
		return
	}

	if input.Title == "" || input.Slug == "" || input.ContentBody == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "error": {"code": "VALIDATION_FAILED", "message": "Title, Slug na Content ni lazima"}}`))
		return
	}

	now := time.Now()
	query := `
		INSERT INTO articles (title, slug, content_body, author_id, category_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'published', $6, $7)
		RETURNING id
	`
	var insertedID string
	err := config.DB.QueryRow(query, input.Title, input.Slug, input.ContentBody, userID, input.CategoryID, now, now).Scan(&insertedID)
	if err != nil {
		log.Printf("[CONTENT DB ERROR] Kosa kutengeneza makala: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "error": {"code": "SERVER_ERROR", "message": "Imeshindwa kutengeneza makala"}}`))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Makala imechapishwa kikamilifu",
		"data": map[string]interface{}{
			"id": insertedID,
		},
	})
}
