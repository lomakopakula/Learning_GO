package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
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
		if r.Method != http.MethodPost {
			w.Header().Set("Content-Type", "application/json")
			http.Error(w, `"error":"method not allowed"`, http.StatusMethodNotAllowed)
			return
		}

		var user User

		err := json.NewDecoder(r.Body).Decode(&user)

		if err != nil {
			handleHTTPError(err, "invalid request body", w, http.StatusBadRequest)
			return
		}

		exists, err := user.checkUserExist(db)
		if err != nil {
			handleHTTPError(err, "internal server error - unable to verify users", w, http.StatusInternalServerError)
			return
		}

		if exists {
			handleHTTPError(err, "username already exists", w, http.StatusConflict)
			return
		}

		exists, err = user.checkEMailExist(db)
		if err != nil {
			handleHTTPError(err, "internal server error - unable to verify email", w, http.StatusInternalServerError)
			return
		}

		if exists {
			handleHTTPError(err, "email already exists", w, http.StatusConflict)
			return
		}

		err = user.hashPassword(user.Password)
		if err != nil {
			handleHTTPError(err, "internal server error - unable to hash password", w, http.StatusInternalServerError)
			return
		}

		dbInsertStr := `INSERT INTO userData (username, hashedpassword, email) VALUES ($1, $2, $3) RETURNING id`

		var id int
		err = db.QueryRow(dbInsertStr, user.Username, user.HashedPassword, user.Email).Scan(&id)
		if err != nil {
			handleHTTPError(err, "internal server error - cannot insert to database", w, http.StatusInternalServerError)
			return
		}

		if id != 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"message":"User registeres successfully"}`))
		}
	}
}

func handleHTTPError(err error, message string, w http.ResponseWriter, statusCode int) {
	jsonError := fmt.Sprintf(`{"error":"%s"}`, message)

	log.Printf("%s: %s\n", message, err)

	w.Header().Set("Content-Type", "application/json")

	http.Error(w, jsonError, statusCode)
}

func (user *User) checkUserExist(db *sql.DB) (bool, error) {
	var exists bool

	dbSelectStr := `SELECT EXISTS (SELECT 1 FROM userData WHERE username=$1)`

	err := db.QueryRow(dbSelectStr, user.Username).Scan(&exists)

	if exists {
		exists = true
	}

	return exists, err
}

func (user *User) checkEMailExist(db *sql.DB) (bool, error) {
	var exists bool

	dbSelectStr := `SELECT EXISTS (SELECT 1 FROM userData WHERE email=$1)`

	err := db.QueryRow(dbSelectStr, user.Email).Scan(&exists)

	if exists {
		exists = true
	}

	return exists, err
}

func (user *User) hashPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return err
	}

	user.HashedPassword = string(hash)

	return nil
}
