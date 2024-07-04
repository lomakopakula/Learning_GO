package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"
)

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome on the main page!")
}

func getUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var user map[string]map[string]string
		err := json.NewDecoder(r.Body).Decode(&user)

		if err != nil {
			log.Fatalf("Unable to decode JSON from request body: %s\n", err)
		}

		InsertUsers(user, db)

		w.Write([]byte(`"message":"OK"`))
	}
}

func InsertUsers(m map[string]map[string]string, db *sql.DB) {
	for key := range m {
		userName := m[key]["userName"]
		firstName := m[key]["firstName"]
		secondName := m[key]["secondName"]
		email := m[key]["email"]
		createdAt := time.Now()

		insertStmt := `INSERT INTO userData (username, firstname, secondname, email, createdat) VALUES ($1, $2, $3, $4, $5) RETURNING id`

		var id int
		err := db.QueryRow(insertStmt, userName, firstName, secondName, email, createdAt).Scan(&id)

		if err != nil {
			log.Fatalf("Could not insert data into table userData: %s\n", err)
		}
	}
}

func main() {
	connStr := "postgres://postgres:secretdbpassword@localhost:5432/users?sslmode=disable"

	db, err := sql.Open("postgres", connStr)

	if err != nil {
		log.Fatalf("Could not open database: %s\n", err)
	}

	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Could not ping database: %s\n", err)
	}

	router := http.NewServeMux()

	router.HandleFunc("/", defaultHandler)
	router.HandleFunc("/users", getUsers(db))

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	server.ListenAndServe()
}
