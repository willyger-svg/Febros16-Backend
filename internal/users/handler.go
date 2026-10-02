package users

import (
	"encoding/json"
	"log"
	"net/http"

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
		&user.Bio,
		&user.ProfilePictureURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	
	if err != nil {
		log.Printf("[USER DB ERROR] Kosa kuvuta profile ya mtumiaji %s: %v", userID, err)
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"success": false, "error": {"code": "NOT_FOUND", "message": "Mtumiaji hajapatikana"}}`))
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"user": user,
		},
	})
}
