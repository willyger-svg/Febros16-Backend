package public

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Response Structs
type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Version string `json:"version"`
}

type StatsResponse struct {
	ArticlesCount int64 `json:"articles_count"`
	ProjectsCount int64 `json:"projects_count"`
	TotalCount    int64 `json:"total_count"`
}

type SearchResultItem struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Abstract  string    `json:"abstract"`
	Type      string    `json:"type"` // "article" | "research_project"
	CreatedAt time.Time `json:"created_at"`
}

type SearchResponse struct {
	Query string             `json:"query"`
	Count int                `json:"count"`
	Data  []SearchResultItem `json:"data"`
}

type CategoryItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type NewsletterRequest struct {
	Email string `json:"email"`
}

// Helpers
func RespondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func RespondError(w http.ResponseWriter, status int, message string) {
	RespondJSON(w, status, map[string]string{
		"status":  "error",
		"message": message,
	})
}

// 1. HealthCheckHandler (GET /api/v1/health)
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	response := HealthResponse{
		Status:  "success",
		Message: "FEBROS16 API is running",
		Version: "1.0.0",
	}
	RespondJSON(w, http.StatusOK, response)
}

// 2. StatsHandler (GET /api/v1/stats)
func StatsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var articlesCount int64 = 100
		var projectsCount int64 = 25

		// If DB is connected, fetch real counts
		if db != nil {
			ctx := r.Context()
			_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM articles").Scan(&articlesCount)
			_ = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM research_projects").Scan(&projectsCount)
		}

		response := StatsResponse{
			ArticlesCount: articlesCount,
			ProjectsCount: projectsCount,
			TotalCount:    articlesCount + projectsCount,
		}
		RespondJSON(w, http.StatusOK, response)
	}
}

// 3. SearchHandler (GET /api/v1/search?q=...)
func SearchHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		query := strings.TrimSpace(r.URL.Query().Get("q"))
		if query == "" {
			RespondError(w, http.StatusBadRequest, "Query parameter 'q' is required")
			return
		}

		results := make([]SearchResultItem, 0)

		if db != nil {
			ctx := r.Context()
			searchTerm := fmt.Sprintf("%%%s%%", query)

			// 1. Query articles
			articleQuery := `
				SELECT id, title, COALESCE(abstract, ''), created_at
				FROM articles
				WHERE title ILIKE $1 OR abstract ILIKE $1
				ORDER BY created_at DESC
				LIMIT 20
			`
			if rows, err := db.QueryContext(ctx, articleQuery, searchTerm); err == nil {
				defer rows.Close()
				for rows.Next() {
					var item SearchResultItem
					if err := rows.Scan(&item.ID, &item.Title, &item.Abstract, &item.CreatedAt); err == nil {
						item.Type = "article"
						results = append(results, item)
					}
				}
			}

			// 2. Query research_projects
			projQuery := `
				SELECT id, title, COALESCE(abstract, ''), created_at
				FROM research_projects
				WHERE title ILIKE $1 OR abstract ILIKE $1
				ORDER BY created_at DESC
				LIMIT 20
			`
			if projRows, err := db.QueryContext(ctx, projQuery, searchTerm); err == nil {
				defer projRows.Close()
				for projRows.Next() {
					var item SearchResultItem
					if err := projRows.Scan(&item.ID, &item.Title, &item.Abstract, &item.CreatedAt); err == nil {
						item.Type = "research_project"
						results = append(results, item)
					}
				}
			}
		} else {
			// Mock result if DB not configured
			results = append(results, SearchResultItem{
				ID:        "art-01",
				Title:     "The Architecture of Autonomous Knowledge Synthesis Systems",
				Abstract:  "How verifiable metadata and distributed graph databases ensure that synthetic research assistants remain grounded in primary citations.",
				Type:      "article",
				CreatedAt: time.Now(),
			})
		}

		response := SearchResponse{
			Query: query,
			Count: len(results),
			Data:  results,
		}
		RespondJSON(w, http.StatusOK, response)
	}
}

// 4. CategoriesHandler (GET /api/v1/categories)
func CategoriesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		categories := make([]CategoryItem, 0)

		if db != nil {
			ctx := r.Context()
			rows, err := db.QueryContext(ctx, "SELECT id, name, slug, COALESCE(description, ''), COALESCE(icon, '') FROM categories ORDER BY name ASC")
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var cat CategoryItem
					if err := rows.Scan(&cat.ID, &cat.Name, &cat.Slug, &cat.Description, &cat.Icon); err == nil {
						categories = append(categories, cat)
					}
				}
			}
		}

		if len(categories) == 0 {
			// Default categories
			categories = []CategoryItem{
				{ID: "education", Name: "Education", Slug: "education", Description: "Pedagogical frameworks", Icon: "GraduationCap"},
				{ID: "technology", Name: "Technology", Slug: "technology", Description: "Systems architecture", Icon: "Cpu"},
				{ID: "environment", Name: "Environment", Slug: "environment", Description: "Ecological resilience", Icon: "Trees"},
				{ID: "research", Name: "Research", Slug: "research", Description: "Scientific methods", Icon: "Microscope"},
			}
		}

		RespondJSON(w, http.StatusOK, map[string]interface{}{
			"categories": categories,
		})
	}
}

// 5. NewsletterHandler (POST /api/v1/newsletter)
func NewsletterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			RespondError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		var req NewsletterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || !strings.Contains(req.Email, "@") {
			RespondError(w, http.StatusBadRequest, "Invalid email address")
			return
		}

		if db != nil {
			_, _ = db.ExecContext(r.Context(), "INSERT INTO newsletter_subscribers (email) VALUES ($1) ON CONFLICT (email) DO NOTHING", strings.TrimSpace(req.Email))
		}

		RespondJSON(w, http.StatusOK, map[string]string{
			"status":  "success",
			"message": "Thank you for subscribing to the Febros16 Dispatch.",
		})
	}
}
