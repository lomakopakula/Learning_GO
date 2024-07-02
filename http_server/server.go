package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome on the main page!")
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	var user map[string]map[string]string
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		log.Fatalf("Unable to decode JSON from request body: %s\n", err)
	}

	displayMap(user)

	w.Write([]byte(`"message":"OK"`))
}

func main() {
	router := http.NewServeMux()

	router.HandleFunc("/", defaultHandler)
	router.HandleFunc("/users", getUsers)

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	server.ListenAndServe()
}

func displayMap(m map[string]map[string]string) {
	for key, val := range m {
		fmt.Printf("User: %s\n", key)
		for key, val := range val {
			fmt.Printf(" %s: %s\n", key, val)
		}
	}
}
