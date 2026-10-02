package auth

import (
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
	query := `
		INSERT INTO users (full_name, email, password_hash, role, bio, created_at, updated_at) 
		VALUES ($1, $2, $3, 'user', 'Mtumiaji mpya wa FEBROS16', $4, $5)
		RETURNING id
	`
	var insertedID string
	now := time.Now()
	err = config.DB.QueryRow(query, input.FullName, input.Email, string(hashedPassword), now, now).Scan(&insertedID)
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
	query := `SELECT id, password_hash, role FROM users WHERE email = $1`
	err := config.DB.QueryRow(query, input.Email).Scan(&user.ID, &user.PasswordHash, &user.Role)
	if err != nil {
		log.Printf("[AUTH DB ERROR] Mtumiaji hajapatikana au kosa la SQL wakati wa Login: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Email au Nenosiri sio sahihi"}}`))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
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
		"data": map[string]string{
			"token": tokenString,
		},
	})
}
