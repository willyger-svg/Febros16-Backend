package auth

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"febros16-backend/config"
	"febros16-backend/internal/models"
)

type RegisterInput struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"success": false, "error": {"code": "METHOD_NOT_ALLOWED", "message": "Njia hairuhusiwi"}}`))
		return
	}

	var input RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "error": {"code": "BAD_REQUEST", "message": "Taarifa hazisomeki"}}`))
		return
	}

	if input.FullName == "" || input.Email == "" || input.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "error": {"code": "VALIDATION_FAILED", "message": "Jaza taarifa zote muhimu"}}`))
		return
	}

	if len(input.Password) < 8 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "error": {"code": "VALIDATION_FAILED", "message": "Nenosiri lazima liwe na angalau herufi 8"}}`))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		log.Printf("[AUTH ERROR] Kushindwa ku-hash password: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "error": {"code": "SERVER_ERROR", "message": "Kosa la kiusalama"}}`))
		return
	}

	now := time.Now()
	var existingID string
	checkQuery := `SELECT id FROM users WHERE email = $1`
	err = config.DB.QueryRow(checkQuery, input.Email).Scan(&existingID)

	if err == nil {
		// User exists
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(`{"success": false, "error": {"code": "CONFLICT", "message": "Email hii tayari inatumika."}}`))
		return
	} else if err == sql.ErrNoRows {
		// Insert new user
		insertQuery := `
			INSERT INTO users (full_name, email, password_hash, role, bio, created_at, updated_at, is_email_verified) 
			VALUES ($1, $2, $3, 'user', 'Mtumiaji mpya wa FEBROS16', $4, $5, TRUE)
			RETURNING id, role
		`
		var userID, role string
		err := config.DB.QueryRow(insertQuery, input.FullName, input.Email, string(hashedPassword), now, now).Scan(&userID, &role)
		if err != nil {
			log.Printf("[AUTH DB ERROR] Kosa wakati wa kusajili mtumiaji mpya: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"success": false, "error": {"code": "SERVER_ERROR", "message": "Kuna tatizo la Database."}}`))
			return
		}

		// Generate Token
		tokenString, _ := GenerateToken(userID, role)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Akaunti imetengenezwa kikamilifu",
			"data": map[string]interface{}{
				"token": tokenString,
			},
		})
		return
	} else {
		log.Printf("[AUTH DB ERROR] Kosa wakati wa kuangalia email: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "error": {"code": "SERVER_ERROR", "message": "Kuna tatizo la Database."}}`))
		return
	}
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"success": false, "error": {"code": "METHOD_NOT_ALLOWED", "message": "Njia hairuhusiwi"}}`))
		return
	}

	var input LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "error": {"code": "BAD_REQUEST", "message": "Taarifa hazisomeki"}}`))
		return
	}

	var user models.User
	var hasCompletedAssessment *bool
	var isEmailVerified bool
	query := `SELECT id, password_hash, role, has_completed_assessment, COALESCE(is_email_verified, false) FROM users WHERE email = $1`
	err := config.DB.QueryRow(query, input.Email).Scan(&user.ID, &user.PasswordHash, &user.Role, &hasCompletedAssessment, &isEmailVerified)
	if err != nil {
		log.Printf("[AUTH DB ERROR] Mtumiaji hajapatikana au kosa la SQL wakati wa Login: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Email au Nenosiri sio sahihi"}}`))
		return
	}

	
	if user.PasswordHash == nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Akaunti hii inatumia Google Login. Tafadhali ingia na Google."}}`))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(input.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Email au Nenosiri sio sahihi"}}`))
		return
	}
	// OTP Requirement Removed completely. Everyone can login if password matches.

	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		secretKey = "siri_ya_akiba"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		log.Printf("[AUTH ERROR] Imeshindwa kusign JWT Token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "error": {"code": "SERVER_ERROR", "message": "Imeshindwa kutengeneza ufunguo"}}`))
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Umeingia kikamilifu!",
		"data": map[string]interface{}{
			"token": tokenString,
			"has_completed_assessment": hasCompletedAssessment != nil && *hasCompletedAssessment,
		},
	})
}

type VerifyOTPInput struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

func VerifyEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var input VerifyOTPInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "error": {"message": "Taarifa hazisomeki"}}`))
		return
	}

	var userID, role string
	var otpCode *string
	var otpExpiry *time.Time
	
	err := config.DB.QueryRow(`SELECT id, role, otp_code, otp_expiry FROM users WHERE email = $1`, input.Email).Scan(&userID, &role, &otpCode, &otpExpiry)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success": false, "error": {"message": "Mtumiaji hajapatikana"}}`))
		return
	}

	if otpCode == nil || *otpCode != input.OTP {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success": false, "error": {"message": "Namba ya OTP sio sahihi"}}`))
		return
	}

	if otpExpiry != nil && time.Now().After(*otpExpiry) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success": false, "error": {"message": "Namba ya OTP imeshaisha muda wake. Tafadhali omba nyingine."}}`))
		return
	}

	// OTP is valid
	_, err = config.DB.Exec(`UPDATE users SET is_email_verified = TRUE, otp_code = NULL, otp_expiry = NULL WHERE id = $1`, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Generate Token
	tokenString, err := GenerateToken(userID, role)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Akaunti imethibitishwa kikamilifu!",
		"data": map[string]interface{}{
			"token": tokenString,
		},
	})
}


func GenerateToken(userID, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "super_secret_default_key_change_me"
	}
	return token.SignedString([]byte(secret))
}