package users

import (
	"encoding/json"
	"net/http"

	"febros16-backend/config"
	"febros16-backend/internal/middleware"
	"febros16-backend/internal/models"
)

// GetMyProfile inarudisha taarifa za mtumiaji aliye-login pamoja na wasifu wake
func GetMyProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": {"code": "METHOD_NOT_ALLOWED", "message": "Njia hairuhusiwi"}}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Pata ID ya mtumiaji kutoka kwenye Token context (Iliwekwa na RequireAuth Middleware)
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Imeshindwa kuthibitisha mtumiaji"}}`))
		return
	}

	var user models.User
	// Preload "Profile" ili ilete taarifa zote kwa pamoja
	if err := config.DB.Preload("Profile").Where("id = ?", userID).First(&user).Error; err != nil {
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
