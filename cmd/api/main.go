package main

import (
	"fmt"
	"net/http"

	"febros16-backend/config"
	"febros16-backend/internal/auth"
)

func main() {
	// Washa database na utengeneze majedwali (Auto-Migrate)
	config.ConnectDB()

	// Njia ya mwanzo (Testing)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprintf(w, "Karibu kwenye Backend ya FEBROS16! API ipo hewani.")
	})

	// Njia za Usajili na Kuingia (Authentication Routes)
	http.HandleFunc("/api/v1/auth/register", auth.Register)
	http.HandleFunc("/api/v1/auth/login", auth.Login)

	// Washa Server
	fmt.Println("Server inawaka port 8080...")
	http.ListenAndServe(":8080", nil)
}
