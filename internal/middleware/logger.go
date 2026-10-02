package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggerMiddleware logs basic information about incoming HTTP requests
func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Call the next handler
		next.ServeHTTP(w, r)
		
		// Log the request
		duration := time.Since(start)
		log.Printf("[%s] %s %s", r.Method, r.URL.Path, duration)
	})
}
