package main

import (
	"fmt"
	"net/http"
)

func main() {
	// Hii ndio API yetu ya kwanza itakayopokea maombi
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Karibu kwenye Backend ya FEBROS16 imeandaliwa na Go!")
	})

	// Hapa tunawasha server yetu
	fmt.Println("Server ya FEBROS inawaka kwenye port 8080...")
	http.ListenAndServe(":8080", nil)
}
