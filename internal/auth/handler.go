package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
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

	otp := fmt.Sprintf("%06d", rand.Intn(1000000))
	otpExpiry := time.Now().Add(15 * time.Minute)
	now := time.Now()

	// Check if user exists and if they are verified
	var existingID string
	var isVerified bool
	checkQuery := `SELECT id, COALESCE(is_email_verified, false) FROM users WHERE email = $1`
	err = config.DB.QueryRow(checkQuery, input.Email).Scan(&existingID, &isVerified)

	if err == nil {
		// User exists
		if isVerified {
			w.WriteHeader(http.StatusConflict)
			w.Write([]byte(`{"success": false, "error": {"code": "CONFLICT", "message": "Email hii tayari inatumika na imethibitishwa."}}`))
			return
		}

		// Update existing unverified user
		updateQuery := `
			UPDATE users 
			SET full_name = $1, password_hash = $2, otp_code = $3, otp_expiry = $4, updated_at = $5
			WHERE id = $6
		`
		_, updateErr := config.DB.Exec(updateQuery, input.FullName, string(hashedPassword), otp, otpExpiry, now, existingID)
		if updateErr != nil {
			log.Printf("[AUTH DB ERROR] Kosa wakati wa kusasisha mtumiaji: %v", updateErr)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"success": false, "error": {"code": "SERVER_ERROR", "message": "Kuna tatizo la Database."}}`))
			return
		}
	} else if err == sql.ErrNoRows {
		// Insert new user
		insertQuery := `
			INSERT INTO users (full_name, email, password_hash, role, bio, created_at, updated_at, otp_code, otp_expiry) 
			VALUES ($1, $2, $3, 'user', 'Mtumiaji mpya wa FEBROS16', $4, $5, $6, $7)
		`
		_, insertErr := config.DB.Exec(insertQuery, input.FullName, input.Email, string(hashedPassword), now, now, otp, otpExpiry)
		if insertErr != nil {
			log.Printf("[AUTH DB ERROR] Kosa wakati wa kusajili mtumiaji mpya: %v", insertErr)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"success": false, "error": {"code": "SERVER_ERROR", "message": "Kuna tatizo la Database."}}`))
			return
		}
	} else {
		log.Printf("[AUTH DB ERROR] Kosa wakati wa kuangalia email: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "error": {"code": "SERVER_ERROR", "message": "Kuna tatizo la Database."}}`))
		return
	}

	// Tuma Email Asynchronously kutumia Goroutine
	go func(email, otpCode string) {
		if sendErr := SendOTPEmail(email, otpCode); sendErr != nil {
			log.Printf("SMTP ERROR (Background): Imeshindwa kutuma barua pepe kwa %s: %v", email, sendErr)
		}
	}(input.Email, otp)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Akaunti imetengenezwa/imesasishwa, angalia email yako kwa OTP",
	})
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

	if !isEmailVerified {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": "Tafadhali thibitisha barua pepe yako kwanza. OTP imetumwa."}`))
		return
	}

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