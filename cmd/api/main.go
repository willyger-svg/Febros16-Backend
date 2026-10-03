package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"febros16-backend/config"
	"febros16-backend/internal/auth"
	"febros16-backend/internal/content"
	"febros16-backend/internal/middleware"
	"febros16-backend/internal/research"
	"febros16-backend/internal/users"
	"febros16-backend/internal/public"
)

func main() {
	// Initialize Database and Auto-Migrate
	config.ConnectDB()

	// Create a new ServeMux for routing
	mux := http.NewServeMux()

	// Health Check / Root Endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success": true, "message": "FEBROS16 API is running", "status": "healthy"}`)
	})

	// PUBLIC Endpoints (API v1)
	mux.HandleFunc("/api/v1/auth/register", auth.Register)
	mux.HandleFunc("/api/v1/auth/login", auth.Login)

	// PROTECTED Endpoints (API v1) - zinatumia middleware.RequireAuth
	mux.HandleFunc("/api/v1/users/me", middleware.RequireAuth(users.GetMyProfile))

	// CONTENT Endpoints (Phase 3)
	mux.HandleFunc("GET /api/v1/articles", content.GetArticles)
	mux.HandleFunc("POST /api/v1/articles", middleware.RequireAuth(content.CreateArticle))
	
	// Single Article Operations
	mux.HandleFunc("GET /api/v1/articles/{id}", content.GetArticle)
	mux.HandleFunc("PUT /api/v1/articles/{id}", middleware.RequireAuth(content.UpdateArticle))
	mux.HandleFunc("DELETE /api/v1/articles/{id}", middleware.RequireAuth(content.DeleteArticle))

	// RESEARCH Endpoints (Phase 4)
	mux.HandleFunc("GET /api/v1/research", research.GetResearchProjects)
	mux.HandleFunc("POST /api/v1/research", middleware.RequireAuth(research.CreateResearchProject))
	mux.HandleFunc("GET /api/v1/research/{id}", research.GetResearchProject)
	mux.HandleFunc("PUT /api/v1/research/{id}", middleware.RequireAuth(research.UpdateResearchProject))
	mux.HandleFunc("DELETE /api/v1/research/{id}", middleware.RequireAuth(research.DeleteResearchProject))

	
	// PUBLIC Endpoints (Phase 4 / Landing Page Integration)
	mux.HandleFunc("GET /api/v1/health", public.HealthCheckHandler)
	mux.HandleFunc("GET /api/v1/stats", public.StatsHandler(config.DB))
	mux.HandleFunc("GET /api/v1/search", public.SearchHandler(config.DB))
	mux.HandleFunc("GET /api/v1/categories", public.CategoriesHandler(config.DB))
	mux.HandleFunc("POST /api/v1/newsletter", public.NewsletterHandler(config.DB))

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
