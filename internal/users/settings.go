package users

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"febros16-backend/config"
	"febros16-backend/internal/middleware"

	"golang.org/x/crypto/bcrypt"
)

type UpdateProfileReq struct {
	FullName  *string `json:"full_name"`
	AvatarURL *string `json:"avatar_url"`
	Bio       *string `json:"bio"`
}

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, `{"success": false, "error": {"message": "Njia hairuhusiwi"}}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success": false, "error": {"message": "Imeshindwa kuthibitisha mtumiaji"}}`))
		return
	}

	var req UpdateProfileReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "error": {"message": "Mfumo wa data sio sahihi"}}`))
		return
	}

	query := `
		UPDATE users 
		SET 
			full_name = COALESCE($1, full_name),
			profile_picture_url = COALESCE($2, profile_picture_url),
			bio = COALESCE($3, bio),
			updated_at = $4
		WHERE id = $5
	`
	_, err := config.DB.Exec(query, req.FullName, req.AvatarURL, req.Bio, time.Now(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "error": {"message": "Imeshindwa kusasisha taarifa"}}`))
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Taarifa zimesasishwa kikamilifu",
	})
}

type ChangePasswordReq struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, `{"success": false, "error": {"message": "Njia hairuhusiwi"}}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success": false, "error": {"message": "Imeshindwa kuthibitisha mtumiaji"}}`))
		return
	}

	var req ChangePasswordReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "error": {"message": "Mfumo wa data sio sahihi"}}`))
		return
	}

	var passwordHash *string
	err := config.DB.QueryRow(`SELECT password_hash FROM users WHERE id = $1`, userID).Scan(&passwordHash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "error": {"message": "Mtumiaji hapatikani"}}`))
		return
	}

	// Kama akaunti ilifunguliwa kwa Google (hana nenosiri)
	if passwordHash == nil || *passwordHash == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "error": {"message": "Akaunti hii inatumia Google Login, huwezi kubadili nenosiri hapa."}}`))
		return
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(*passwordHash), []byte(req.OldPassword)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success": false, "error": {"message": "Nenosiri la zamani si sahihi"}}`))
		return
	}

	// Hash new password
	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "error": {"message": "Imeshindwa kuchakata nenosiri jipya"}}`))
		return
	}

	_, err = config.DB.Exec(`UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`, string(newHash), time.Now(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "error": {"message": "Imeshindwa kuhifadhi nenosiri jipya"}}`))
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Nenosiri limebadilishwa kikamilifu",
	})
}

type DeleteAccountReq struct {
	Confirmation string `json:"confirmation"`
}

func DeleteAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, `{"success": false, "error": {"message": "Njia hairuhusiwi"}}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success": false, "error": {"message": "Imeshindwa kuthibitisha mtumiaji"}}`))
		return
	}

	var req DeleteAccountReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "error": {"message": "Mfumo wa data sio sahihi"}}`))
		return
	}

	if strings.ToUpper(req.Confirmation) != "DELETE" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "error": {"message": "Tafadhali andika 'DELETE' kuthibitisha kufuta akaunti yako"}}`))
		return
	}

	_, err := config.DB.Exec(`DELETE FROM users WHERE id = $1`, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "error": {"message": "Imeshindwa kufuta akaunti"}}`))
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Akaunti yako imefutwa kikamilifu pamoja na taarifa zake zote",
	})
}
