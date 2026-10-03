package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleUserInfo holds user data returned by Google OAuth2 UserInfo endpoint
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	Picture       string `json:"picture"`
}

// OAuthService manages authentication logic with database/sql and JWT
type OAuthService struct {
	DB          *sql.DB
	Config      *oauth2.Config
	JWTSecret   []byte
	FrontendURL string
}

// NewOAuthService initializes an OAuthService instance with environment variables
func NewOAuthService(db *sql.DB) *OAuthService {
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		log.Fatal("KOSA: FRONTEND_URL environment variable haijapatikana!")
	}

	secretKey := os.Getenv("JWT_SECRET")
	if secretKey == "" {
		secretKey = "siri_ya_akiba"
	}

	return &OAuthService{
		DB: db,
		Config: &oauth2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
		JWTSecret:   []byte(secretKey),
		FrontendURL: frontendURL,
	}
}

func generateRandomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (s *OAuthService) HandleGoogleLogin(w http.ResponseWriter, r *http.Request) {
	state, err := generateRandomState()
	if err != nil {
		http.Error(w, "Hitilafu katika kutengeneza token ya usalama", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Expires:  time.Now().Add(10 * time.Minute),
		HttpOnly: true,
		Secure:   r.TLS != nil || os.Getenv("ENV") == "production",
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})

	url := s.Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (s *OAuthService) HandleGoogleCallback(w http.ResponseWriter, r *http.Request) {
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
		http.Error(w, "Usalama: State token haioani", http.StatusBadRequest)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Path:     "/",
	})

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Kodi ya uthibitisho haijapatikana kutoka Google", http.StatusBadRequest)
		return
	}

	token, err := s.Config.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Imeshindikana kubadilisha code: %v", err), http.StatusInternalServerError)
		return
	}

	client := s.Config.Client(r.Context(), token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, "Imeshindikana kupata taarifa za mtumiaji kutoka Google", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Google API imerudisha hitilafu wakati wa kupata taarifa", http.StatusBadRequest)
		return
	}

	var googleUser GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&googleUser); err != nil {
		http.Error(w, "Hitilafu katika kusoma taarifa za mtumiaji", http.StatusInternalServerError)
		return
	}

	userID, role, err := s.upsertGoogleUser(r.Context(), googleUser)
	if err != nil {
		http.Error(w, fmt.Sprintf("Hitilafu kwenye Database: %v", err), http.StatusInternalServerError)
		return
	}

	jwtToken, err := s.generateSystemJWT(userID, role)
	if err != nil {
		http.Error(w, "Imeshindikana kutengeneza token ya mfumo", http.StatusInternalServerError)
		return
	}

	// Important: Check if they have completed assessment
	var hasCompletedAssessment *bool
	_ = s.DB.QueryRowContext(r.Context(), `SELECT has_completed_assessment FROM users WHERE id = $1`, userID).Scan(&hasCompletedAssessment)
	var assessedStr string
	if hasCompletedAssessment != nil && *hasCompletedAssessment {
		assessedStr = "true"
	} else {
		assessedStr = "false"
	}

	redirectTarget := fmt.Sprintf("%s/login?token=%s&assessed=%s", s.FrontendURL, jwtToken, assessedStr)
	http.Redirect(w, r, redirectTarget, http.StatusTemporaryRedirect)
}

func (s *OAuthService) upsertGoogleUser(ctx context.Context, gu GoogleUserInfo) (string, string, error) {
	var userID, role string

	queryByGoogleID := `SELECT id, role FROM users WHERE google_id = $1 LIMIT 1`
	err := s.DB.QueryRowContext(ctx, queryByGoogleID, gu.ID).Scan(&userID, &role)
	if err == nil {
		updateAvatarQuery := `UPDATE users SET profile_picture_url = COALESCE(profile_picture_url, $1), updated_at = $2 WHERE id = $3`
		_, _ = s.DB.ExecContext(ctx, updateAvatarQuery, gu.Picture, time.Now(), userID)
		return userID, role, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return "", "", err
	}

	queryByEmail := `SELECT id, role FROM users WHERE email = $1 LIMIT 1`
	err = s.DB.QueryRowContext(ctx, queryByEmail, gu.Email).Scan(&userID, &role)
	if err == nil {
		linkQuery := `UPDATE users SET google_id = $1, profile_picture_url = COALESCE(profile_picture_url, $2), updated_at = $3 WHERE id = $4`
		_, err = s.DB.ExecContext(ctx, linkQuery, gu.ID, gu.Picture, time.Now(), userID)
		if err != nil {
			return "", "", err
		}
		return userID, role, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return "", "", err
	}

	insertQuery := `
		INSERT INTO users (full_name, email, google_id, profile_picture_url, role, bio, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 'user', 'Mtumiaji wa FEBROS16 (Google Auth)', $5, $6)
		RETURNING id, role
	`
	now := time.Now()
	err = s.DB.QueryRowContext(ctx, insertQuery, gu.Name, gu.Email, gu.ID, gu.Picture, now, now).Scan(&userID, &role)
	if err != nil {
		return "", "", err
	}

	return userID, role, nil
}

func (s *OAuthService) generateSystemJWT(userID string, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.JWTSecret)
}
