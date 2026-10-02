package middleware

import (
	"net/http"
	"os"
	"strings"
)

// CORSMiddleware handles Cross-Origin Resource Sharing based on environment configuration
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		if allowedOrigins == "" {
			allowedOrigins = "*" // Fallback for dev, but should be restricted in production
		}

		origin := r.Header.Get("Origin")
		
		// If the origin is in our allowed list (or if we allow all)
		allow := false
		if allowedOrigins == "*" {
			allow = true
		} else {
			origins := strings.Split(allowedOrigins, ",")
			for _, o := range origins {
				if strings.TrimSpace(o) == origin {
					allow = true
					break
				}
			}
		}

		if allow && origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if allowedOrigins == "*" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, PATCH")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
