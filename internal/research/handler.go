package research

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"febros16-backend/config"
	"febros16-backend/internal/middleware"
	"febros16-backend/internal/models"
)

type CreateResearchInput struct {
	Title       string  `json:"title"`
	Slug        string  `json:"slug"`
	Abstract    *string `json:"abstract"`
	ContentBody string  `json:"content_body"`
}

// GetResearchProjects - Public API to list projects
func GetResearchProjects(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"success": false, "error": {"code": "METHOD_NOT_ALLOWED", "message": "Njia hairuhusiwi"}}`))
		return
	}

	authorIDQuery := r.URL.Query().Get("author_id")
	
	query := `
		SELECT id, title, slug, abstract, content_body, author_id, status, start_date, end_date, created_at, updated_at
		FROM research_projects
	`
	var args []interface{}
	
	if authorIDQuery != "" {
		query += ` WHERE author_id = $1`
		args = append(args, authorIDQuery)
	}
	
	query += ` ORDER BY created_at DESC`

	rows, err := config.DB.Query(query, args...)
	if err != nil {
		log.Printf("[RESEARCH DB ERROR] Kosa kuvuta miradi: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "error": {"code": "SERVER_ERROR", "message": "Imeshindwa kuvuta miradi ya utafiti"}}`))
		return
	}
	defer rows.Close()

	var projects []models.ResearchProject
	for rows.Next() {
		var p models.ResearchProject
		var abstract, authorID *string
		var startDate, endDate, createdAt, updatedAt *time.Time

		err := rows.Scan(&p.ID, &p.Title, &p.Slug, &abstract, &p.ContentBody, &authorID, &p.Status, &startDate, &endDate, &createdAt, &updatedAt)
		if err != nil {
			log.Printf("[RESEARCH DB ERROR] Scan error: %v", err)
			continue
		}
		
		p.Abstract = abstract
		p.AuthorID = authorID
		p.StartDate = startDate
		p.EndDate = endDate
		p.CreatedAt = createdAt
		p.UpdatedAt = updatedAt

		projects = append(projects, p)
	}

	if projects == nil {
		projects = []models.ResearchProject{}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"projects": projects,
		},
	})
}

// GetResearchProject - Public API to get a single project by ID
func GetResearchProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"success": false, "error": {"code": "METHOD_NOT_ALLOWED", "message": "Njia hairuhusiwi"}}`))
		return
	}

	id := r.PathValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "error": {"code": "BAD_REQUEST", "message": "ID ya mradi inahitajika"}}`))
		return
	}

	query := `
		SELECT id, title, slug, abstract, content_body, author_id, status, start_date, end_date, created_at, updated_at
		FROM research_projects
		WHERE id = $1
	`
	var p models.ResearchProject
	var abstract, authorID *string
	var startDate, endDate, createdAt, updatedAt *time.Time

	err := config.DB.QueryRow(query, id).Scan(&p.ID, &p.Title, &p.Slug, &abstract, &p.ContentBody, &authorID, &p.Status, &startDate, &endDate, &createdAt, &updatedAt)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"success": false, "error": {"code": "NOT_FOUND", "message": "Mradi haujapatikana"}}`))
		return
	}

	p.Abstract = abstract
	p.AuthorID = authorID
	p.StartDate = startDate
	p.EndDate = endDate
	p.CreatedAt = createdAt
	p.UpdatedAt = updatedAt

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"project": p,
		},
	})
}

// CreateResearchProject - Protected API
func CreateResearchProject(w http.ResponseWriter, r *http.Request) {
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

	var input CreateResearchInput
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
		INSERT INTO research_projects (title, slug, abstract, content_body, author_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 'ongoing', $6, $7)
		RETURNING id
	`
	var insertedID string
	err := config.DB.QueryRow(query, input.Title, input.Slug, input.Abstract, input.ContentBody, userID, now, now).Scan(&insertedID)
	if err != nil {
		log.Printf("[RESEARCH DB ERROR] Kosa kutengeneza mradi: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "error": {"code": "SERVER_ERROR", "message": "Imeshindwa kutengeneza mradi"}}`))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Mradi umesajiliwa kikamilifu",
		"data": map[string]interface{}{
			"id": insertedID,
		},
	})
}

// UpdateResearchProject - Protected API
func UpdateResearchProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPut {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"success": false, "error": {"code": "METHOD_NOT_ALLOWED", "message": "Njia hairuhusiwi"}}`))
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	role, _ := r.Context().Value(middleware.UserRoleKey).(string)
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Imeshindwa kuthibitisha mtumiaji"}}`))
		return
	}

	id := r.PathValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "error": {"code": "BAD_REQUEST", "message": "ID ya mradi inahitajika"}}`))
		return
	}

	var input CreateResearchInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "error": {"code": "BAD_REQUEST", "message": "Taarifa hazisomeki"}}`))
		return
	}

	now := time.Now()
	var updateQuery string
	var err error
	var result sql.Result
	
	if role == "admin" {
		updateQuery = `
			UPDATE research_projects 
			SET title = $1, slug = $2, abstract = $3, content_body = $4, updated_at = $5
			WHERE id = $6
		`
		result, err = config.DB.Exec(updateQuery, input.Title, input.Slug, input.Abstract, input.ContentBody, now, id)
	} else {
		updateQuery = `
			UPDATE research_projects 
			SET title = $1, slug = $2, abstract = $3, content_body = $4, updated_at = $5
			WHERE id = $6 AND author_id = $7
		`
		result, err = config.DB.Exec(updateQuery, input.Title, input.Slug, input.Abstract, input.ContentBody, now, id, userID)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "error": {"code": "SERVER_ERROR", "message": "Imeshindwa kurekebisha mradi"}}`))
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"success": false, "error": {"code": "NOT_FOUND", "message": "Mradi haujapatikana au huna ruhusa ya kuurekebisha"}}`))
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Mradi umerekebishwa kikamilifu",
	})
}

// DeleteResearchProject - Protected API
func DeleteResearchProject(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodDelete {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"success": false, "error": {"code": "METHOD_NOT_ALLOWED", "message": "Njia hairuhusiwi"}}`))
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	role, _ := r.Context().Value(middleware.UserRoleKey).(string)
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Imeshindwa kuthibitisha mtumiaji"}}`))
		return
	}

	id := r.PathValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "error": {"code": "BAD_REQUEST", "message": "ID ya mradi inahitajika"}}`))
		return
	}

	var deleteQuery string
	var err error
	var result sql.Result
	
	if role == "admin" {
		deleteQuery = `DELETE FROM research_projects WHERE id = $1`
		result, err = config.DB.Exec(deleteQuery, id)
	} else {
		deleteQuery = `DELETE FROM research_projects WHERE id = $1 AND author_id = $2`
		result, err = config.DB.Exec(deleteQuery, id, userID)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "error": {"code": "SERVER_ERROR", "message": "Imeshindwa kufuta mradi"}}`))
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"success": false, "error": {"code": "NOT_FOUND", "message": "Mradi haujapatikana au huna ruhusa ya kuufuta"}}`))
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Mradi umefutwa kikamilifu",
	})
}
