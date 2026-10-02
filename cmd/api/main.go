package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"febros16-backend/config"
	"febros16-backend/internal/auth"
	"febros16-backend/internal/middleware"
)

func main() {
	// Initialize Database and Auto-Migrate
	config.ConnectDB()

	// Create a new ServeMux for routing
	mux := http.NewServeMux()

	// Health Check / Root Endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "FEBROS16 API is running", "status": "healthy"}`)
	})

	// Auth Endpoints (API v1)
	mux.HandleFunc("/api/v1/auth/register", auth.Register)
	mux.HandleFunc("/api/v1/auth/login", auth.Login)

	// Apply Middlewares: Logger -> CORS -> Mux
	handler := middleware.LoggerMiddleware(middleware.CORSMiddleware(mux))

	// Determine Port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start Server
	log.Printf("Server is starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
