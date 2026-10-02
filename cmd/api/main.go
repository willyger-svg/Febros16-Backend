package main

import (
	"fmt"
	"net/http"

	"febros16-backend/config"
	"febros16-backend/internal/auth"
)

// CorsMiddleware inawezesha Frontend ya Vercel (na popote pengine) kuwasiliana na Backend ya Render
func CorsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		// Kama ni preflight request (OPTIONS), irudishe mapema
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func main() {
	// Washa database na utengeneze majedwali (Auto-Migrate)
	config.ConnectDB()

	// Njia ya mwanzo (Testing)
	http.HandleFunc("/", CorsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Karibu kwenye Backend ya FEBROS16! API ipo hewani.")
	}))

	// Njia za Usajili na Kuingia (Authentication Routes) zimefungwa ndani ya CorsMiddleware
	http.HandleFunc("/api/v1/auth/register", CorsMiddleware(auth.Register))
	http.HandleFunc("/api/v1/auth/login", CorsMiddleware(auth.Login))

	// Washa Server
	fmt.Println("Server inawaka port 8080...")
	http.ListenAndServe(":8080", nil)
}
