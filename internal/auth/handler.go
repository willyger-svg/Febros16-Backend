package auth

import (
	"crypto/rand"
	"encoding/hex"
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

	// Insert using raw SQL
	tokenBytes := make([]byte, 32)
	rand.Read(tokenBytes)
	verificationToken := hex.EncodeToString(tokenBytes)

	query := `
		INSERT INTO users (full_name, email, password_hash, role, bio, created_at, updated_at, verification_token) 
		VALUES ($1, $2, $3, 'user', 'Mtumiaji mpya wa FEBROS16', $4, $5, $6)
		RETURNING id
	`
	var insertedID string
	now := time.Now()
	err = config.DB.QueryRow(query, input.FullName, input.Email, string(hashedPassword), now, now, verificationToken).Scan(&insertedID)
	
	if err == nil {
		go func() {
			if sendErr := SendVerificationEmail(input.Email, verificationToken); sendErr != nil {
				log.Printf("Error sending email to %s: %v", input.Email, sendErr)
			}
		}()
	}
	if err != nil {
		log.Printf("[AUTH DB ERROR] Kosa wakati wa kusajili mtumiaji mpya: %v", err)
		w.WriteHeader(http.StatusConflict) // au StatusInternalServerError kutegemea na kosa
		w.Write([]byte(`{"success": false, "error": {"code": "CONFLICT", "message": "Kuna tatizo la Database. Email inaweza kuwa imeshasajiliwa."}}`))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Usajili umefanikiwa kikamilifu!",
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

// VerifyEmail handles email verification links
func VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token ya uthibitisho inahitajika", http.StatusBadRequest)
		return
	}

	query := `UPDATE users SET is_email_verified = TRUE, verification_token = NULL WHERE verification_token = $1 RETURNING id`
	var userID string
	err := config.DB.QueryRow(query, token).Scan(&userID)
	if err != nil {
		http.Error(w, "Token sio sahihi au imeshaisha muda wake", http.StatusBadRequest)
		return
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		http.Error(w, "KOSA: FRONTEND_URL haijasanidiwa kwenye server (Missing Env)", http.StatusInternalServerError)
		return
	}
	// Redirect to login with success message
	http.Redirect(w, r, frontendURL+"/login?verified=true", http.StatusTemporaryRedirect)
}
