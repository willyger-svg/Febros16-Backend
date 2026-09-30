package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Hizi mistari mbili ndizo zinazoruhusu CORS (Mawasiliano kutoka Netlify)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		
		fmt.Fprintf(w, "Karibu kwenye Backend ya FEBROS16 imeandaliwa na Go!")
	})

	fmt.Println("Server ya FEBROS inawaka kwenye port 8080...")
	http.ListenAndServe(":8080", nil)
}
