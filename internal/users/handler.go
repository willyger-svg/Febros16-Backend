package users

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"febros16-backend/config"
	"febros16-backend/internal/middleware"
	"febros16-backend/internal/models"
)

func GetMyProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"success": false, "error": {"code": "METHOD_NOT_ALLOWED", "message": "Njia hairuhusiwi"}}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Imeshindwa kuthibitisha mtumiaji"}}`))
		return
	}

	var user models.User
	var bio, profilePic *string
	var createdAt, updatedAt *time.Time

	query := `
		SELECT id, full_name, email, role, bio, profile_picture_url, created_at, updated_at 
		FROM users 
		WHERE id = $1
	`
	err := config.DB.QueryRow(query, userID).Scan(
		&user.ID,
		&user.FullName,
		&user.Email,
		&user.Role,
		&bio,
		&profilePic,
		&createdAt,
		&updatedAt,
	)
	
	if bio != nil {
		user.Bio = *bio
	}
	if profilePic != nil {
		user.ProfilePictureURL = *profilePic
	}
	if createdAt != nil {
		user.CreatedAt = *createdAt
	}
	if updatedAt != nil {
		user.UpdatedAt = *updatedAt
	}
	
	if err != nil {
		log.Printf("[USER DB ERROR] Kosa kuvuta profile ya mtumiaji %s: %v", userID, err)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"success": false, "error": {"code": "NOT_FOUND", "message": "Mtumiaji hajapatikana"}}`))
		return
	}

	// Kuvuta hesabu ya makala alizoandika mtumiaji
	var totalArticles int
	countQuery := `SELECT COUNT(*) FROM articles WHERE author_id = $1`
	err = config.DB.QueryRow(countQuery, userID).Scan(&totalArticles)
	if err != nil {
		log.Printf("[USER DB ERROR] Kosa kuvuta hesabu ya makala za %s: %v", userID, err)
		totalArticles = 0 // Default to 0 on error
	}

	// Kuvuta hesabu ya miradi ya utafiti
	var totalResearchProjects int
	researchCountQuery := `SELECT COUNT(*) FROM research_projects WHERE author_id = $1`
	err = config.DB.QueryRow(researchCountQuery, userID).Scan(&totalResearchProjects)
	if err != nil {
		log.Printf("[USER DB ERROR] Kosa kuvuta hesabu ya miradi ya utafiti ya %s: %v", userID, err)
		totalResearchProjects = 0
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"user": user,
			"stats": map[string]int{
				"total_articles": totalArticles,
				"total_research_projects": totalResearchProjects,
			},
		},
	})
}
