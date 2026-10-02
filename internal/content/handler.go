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

	authorIDQuery := r.URL.Query().Get("author_id")
	
	query := `
		SELECT id, title, slug, content_body, author_id, category_id, status, created_at, updated_at
		FROM articles
	`
	var args []interface{}
	
	if authorIDQuery != "" {
		query += ` WHERE author_id = $1`
		args = append(args, authorIDQuery)
	}
	
	query += ` ORDER BY created_at DESC`

	rows, err := config.DB.Query(query, args...)
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

// GetArticle - Public API to get a single article by ID
func GetArticle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"success": false, "error": {"code": "METHOD_NOT_ALLOWED", "message": "Njia hairuhusiwi"}}`))
		return
	}

	id := r.PathValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "error": {"code": "BAD_REQUEST", "message": "ID ya makala inahitajika"}}`))
		return
	}

	query := `
		SELECT id, title, slug, content_body, author_id, category_id, status, created_at, updated_at
		FROM articles
		WHERE id = $1
	`
	var a models.Article
	var authorID, categoryID *string
	var createdAt, updatedAt *time.Time

	err := config.DB.QueryRow(query, id).Scan(&a.ID, &a.Title, &a.Slug, &a.ContentBody, &authorID, &categoryID, &a.Status, &createdAt, &updatedAt)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"success": false, "error": {"code": "NOT_FOUND", "message": "Makala haijapatikana"}}`))
		return
	}

	a.AuthorID = authorID
	a.CategoryID = categoryID
	a.CreatedAt = createdAt
	a.UpdatedAt = updatedAt

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"article": a,
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

// UpdateArticle - Protected API
func UpdateArticle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"success": false, "error": {"code": "METHOD_NOT_ALLOWED", "message": "Njia hairuhusiwi"}}`))
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	role, _ := r.Context().Value(middleware.RoleKey).(string)
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Imeshindwa kuthibitisha mtumiaji"}}`))
		return
	}

	id := r.PathValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "error": {"code": "BAD_REQUEST", "message": "ID ya makala inahitajika"}}`))
		return
	}

	var input CreateArticleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "error": {"code": "BAD_REQUEST", "message": "Taarifa hazisomeki"}}`))
		return
	}

	now := time.Now()
	
	// Ownership check and Update
	var updateQuery string
	var err error
	var result string
	
	if role == "admin" {
		updateQuery = `
			UPDATE articles 
			SET title = $1, slug = $2, content_body = $3, updated_at = $4
			WHERE id = $5 RETURNING id
		`
		err = config.DB.QueryRow(updateQuery, input.Title, input.Slug, input.ContentBody, now, id).Scan(&result)
	} else {
		updateQuery = `
			UPDATE articles 
			SET title = $1, slug = $2, content_body = $3, updated_at = $4
			WHERE id = $5 AND author_id = $6 RETURNING id
		`
		err = config.DB.QueryRow(updateQuery, input.Title, input.Slug, input.ContentBody, now, id, userID).Scan(&result)
	}

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"success": false, "error": {"code": "FORBIDDEN", "message": "Huna ruhusa ya kurekebisha makala hii au haipo"}}`))
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Makala imerekebishwa kikamilifu",
	})
}

// DeleteArticle - Protected API
func DeleteArticle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"success": false, "error": {"code": "METHOD_NOT_ALLOWED", "message": "Njia hairuhusiwi"}}`))
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	role, _ := r.Context().Value(middleware.RoleKey).(string)
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Imeshindwa kuthibitisha mtumiaji"}}`))
		return
	}

	id := r.PathValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "error": {"code": "BAD_REQUEST", "message": "ID ya makala inahitajika"}}`))
		return
	}

	var deleteQuery string
	var err error
	var result string
	
	if role == "admin" {
		deleteQuery = `DELETE FROM articles WHERE id = $1 RETURNING id`
		err = config.DB.QueryRow(deleteQuery, id).Scan(&result)
	} else {
		deleteQuery = `DELETE FROM articles WHERE id = $1 AND author_id = $2 RETURNING id`
		err = config.DB.QueryRow(deleteQuery, id, userID).Scan(&result)
	}

	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"success": false, "error": {"code": "FORBIDDEN", "message": "Huna ruhusa ya kufuta makala hii au haipo"}}`))
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Makala imefutwa kikamilifu",
	})
}
