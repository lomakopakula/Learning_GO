package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/crypto/bcrypt"
)

const configFile = "config/config.json"

func main() {
	var config Config

	config.loadConfig(configFile)

	dataSourceName := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=%s",
		config.DB.Driver, config.DB.User, config.DB.Pass, config.DB.Addr, config.DB.Port, config.DB.Name, config.DB.SSLmode)

	db, err := sql.Open(config.DB.Driver, dataSourceName)

	if err != nil {
		log.Fatalf("Could not open database : %s\n", err)
	}

	router := http.NewServeMux()

	router.HandleFunc("/register", handleRegister(db))

	serverPort := ":" + config.Server.Port

	server := http.Server{
		Addr:    serverPort,
		Handler: router,
	}

	server.ListenAndServe()

}

type Config struct {
	DB     DatabaseConfig `json:"database"`
	Server ServerConfig   `json:"server"`
}

type ServerConfig struct {
	Port string `json:"port"`
}

type DatabaseConfig struct {
	Driver  string `json:"driver"`
	User    string `json:"user"`
	Pass    string `json:"password"`
	Addr    string `json:"address"`
	Port    string `json:"port"`
	Name    string `json:"name"`
	SSLmode string `json:"ssl_mode"`
}

func (config *Config) loadConfig(configFile string) {
	file, err := os.Open(configFile)

	if err != nil {
		log.Fatalf("Unable to open file %s : %s\n", configFile, err)
	}

	defer file.Close()

	reader := bufio.NewReader(file)

	json.NewDecoder(reader).Decode(config)
}

type User struct {
	Username       string `json:"username"`
	Password       string `json:"password"`
	HashedPassword string `json:"hashed_password,omitempty"`
	Email          string `json:"email"`
}

func handleRegister(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var user User

		err := json.NewDecoder(r.Body).Decode(&user)

		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
			return
		}

		err = user.hashPassword(user.Password)

		if err != nil {
			log.Printf("Unable to hash password: %s\n", err)
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"Internal server error - unable to hash password"}`, http.StatusInternalServerError)
			return
		}

		dbInsertStr := `INSERT INTO userData (username, hashedpassword, email) VALUES ($1, $2, $3) RETURNING id`

		var id int
		err = db.QueryRow(dbInsertStr, user.Username, user.HashedPassword, user.Email).Scan(&id)

		if err != nil {
			log.Printf("Unable to insert user to the database: %s\n", err)
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `{"error":"Internal server error - Unable to insert user data into the database"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message":"User registeres successfully"}`))

	}
}

func (user *User) hashPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	user.HashedPassword = string(hash)

	return nil
}
