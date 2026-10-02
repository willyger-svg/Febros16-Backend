package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "user_id"
const UserRoleKey contextKey = "user_role"

// RequireAuth ni middleware inayokagua JWT Token kabla ya kuruhusu API iitwe
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Token inahitajika kwenye header"}}`)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Muundo wa token sio sahihi, tumia 'Bearer <token>'"}}`)
			return
		}

		tokenString := parts[1]
		secretKey := os.Getenv("JWT_SECRET")
		if secretKey == "" {
			secretKey = "siri_ya_akiba"
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return []byte(secretKey), nil
		})

		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Token imeisha muda au si halali"}}`)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"success": false, "error": {"code": "UNAUTHORIZED", "message": "Token imeshindwa kusomeka"}}`)
			return
		}

		userID := claims["user_id"].(string)
		userRole := claims["role"].(string)

		// Hifadhi taarifa za mtumiaji kwenye Request Context ili zitumike kwenye endpoint
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, UserRoleKey, userRole)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// RequireRole inazuia watu wasio na haki (Role) fulani (Mfano: "admin")
func RequireRole(role string, next http.HandlerFunc) http.HandlerFunc {
	return RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		userRole := r.Context().Value(UserRoleKey).(string)
		if userRole != role {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, `{"success": false, "error": {"code": "FORBIDDEN", "message": "Huna haki (Permission) ya kufanya hivi"}}`)
			return
		}
		next.ServeHTTP(w, r)
	})
}
