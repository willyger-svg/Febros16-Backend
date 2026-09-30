package auth

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"febros16-backend/config"
	"febros16-backend/internal/models"
)

// -------- SEHEMU YA USAJILI (REGISTER) --------

type RegisterInput struct {
	FullName string `json:"full_name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Njia hairuhusiwi", http.StatusMethodNotAllowed)
		return
	}

	var input RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Taarifa hazisomeki", http.StatusBadRequest)
		return
	}

	if input.FullName == "" || input.Email == "" || input.Password == "" {
		http.Error(w, "Jaza taarifa zote", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	if err != nil {
		http.Error(w, "Kosa la kiusalama", http.StatusInternalServerError)
		return
	}

	user := models.User{
		FullName:     input.FullName,
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
	}

	if err := config.DB.Create(&user).Error; err != nil {
		http.Error(w, "Email imeshasajiliwa", http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Usajili umefanikiwa kikamilifu!",
	})
}

// -------- SEHEMU YA KUINGIA (LOGIN) --------

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Njia hairuhusiwi", http.StatusMethodNotAllowed)
		return
	}

	var input LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Taarifa hazisomeki", http.StatusBadRequest)
		return
	}

	var user models.User
	// 1. Tafuta kama Email ipo kwenye Database
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		http.Error(w, "Email au Nenosiri sio sahihi", http.StatusUnauthorized)
		return
	}

	// 2. Pima kama nenosiri linafanana na lile lililofichwa (Hash)
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		http.Error(w, "Email au Nenosiri sio sahihi", http.StatusUnauthorized)
		return
	}

	// 3. Tengeneza JWT Token (Ufunguo)
	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		secretKey = "siri_ya_akiba"
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 72).Unix(), // Token itadumu masaa 72
	})

	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		http.Error(w, "Imeshindwa kutengeneza ufunguo", http.StatusInternalServerError)
		return
	}

	// 4. Rudisha majibu na huo ufunguo kwa Frontend
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Umeingia kikamilifu!",
		"token":   tokenString,
	})
}

