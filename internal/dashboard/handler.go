package dashboard

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"febros16-backend/config"
	"febros16-backend/internal/middleware"
)

type DashboardStats struct {
	CurrentStreak       int        `json:"current_streak"`
	ActivitiesCompleted int        `json:"activities_completed"`
	TriggersIdentified  int        `json:"triggers_identified"`
	LastCheckinDate     *time.Time `json:"last_checkin_date"`
}

func GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"success": false, "error": {"message": "Njia hairuhusiwi"}}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	// Pata userID kutoka context
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success": false, "error": {"message": "Imeshindwa kuthibitisha mtumiaji"}}`))
		return
	}

	var stats DashboardStats

	query := `
		SELECT current_streak, activities_completed, triggers_identified, last_checkin_date 
		FROM user_progress 
		WHERE user_id = $1
	`
	err := config.DB.QueryRow(query, userID).Scan(
		&stats.CurrentStreak,
		&stats.ActivitiesCompleted,
		&stats.TriggersIdentified,
		&stats.LastCheckinDate,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Kama mtumiaji hana rekodi, rudisha 0 (Default Values)
			stats = DashboardStats{
				CurrentStreak:       0,
				ActivitiesCompleted: 0,
				TriggersIdentified:  0,
				LastCheckinDate:     nil,
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"success": false, "error": {"message": "Tatizo la Database"}}`))
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}
