package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"log"
	"net/http"
	"os"
	"time"

	_ "authentication-service/docs" // Import the generated docs

	"github.com/lib/pq"
	httpSwagger "github.com/swaggo/http-swagger" // http-swagger middleware
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte("your_secret_key")

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var db *sql.DB

func initDB() {
	var err error
	user := os.Getenv("DB_USER")
	dbname := os.Getenv("DB_NAME")
	password := os.Getenv("DB_PASSWORD")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	sslmode := os.Getenv("DB_SSLMODE")
	connStr := fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%s sslmode=%s", user, dbname, password, host, port, sslmode)
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}

}

// signupHandler handles user sign-up requests
// @Summary Sign up a new user
// @Description Create a new user with a username and password
// @Accept  json
// @Produce  json
// @Param   user  body  User  true  "User credentials"
// @Success 200 {string} string "User with username {username} was created"
// @Failure 400 {string} string "Error parsing JSON"
// @Failure 409 {string} string "An user with the provided username already exists"
// @Failure 500 {string} string "Error hashing password" or "Failed to get a response from database"
// @Router /signup [post]
func signupHandler(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO users (username, password) VALUES ($1, $2)", user.Username, string(hashedPassword))
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			http.Error(w, "An user with the provided username already exists", http.StatusConflict)
		} else {
			http.Error(w, "Failed to get a response from database", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "User with username %s was created", user.Username)
}

// signinHandler handles user sign-in requests
// @Summary Sign in an existing user
// @Description Authenticate a user with a username and password
// @Accept  json
// @Produce  json
// @Param   user  body  User  true  "User credentials"
// @Success 200 {object} map[string]string "JWT token"
// @Failure 400 {string} string "Error parsing JSON"
// @Failure 401 {string} string "Password is incorrect"
// @Failure 404 {string} string "No user with that username"
// @Failure 405 {string} string "Invalid request method"
// @Failure 500 {string} string "Failed to get a response from database" or "Error generating token"
// @Router /signin [get]
func signinHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Error parsing JSON", http.StatusBadRequest)
		return
	}

	var storedHashedPassword string
	err = db.QueryRow("SELECT password FROM users WHERE username = $1", user.Username).Scan(&storedHashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "No user with that username", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get a response from database", http.StatusInternalServerError)
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHashedPassword), []byte(user.Password))
	if err != nil {
		http.Error(w, "Password is incorrect", http.StatusUnauthorized)
		return
	}

	// Create the JWT token
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.StandardClaims{
		ExpiresAt: expirationTime.Unix(),
		Subject:   user.Username,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	// Return the JWT token
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

func main() {
	initDB()
	http.HandleFunc("/signup", signupHandler)
	http.HandleFunc("/signin", signinHandler)
	// Serve the Swagger UI
	http.Handle("/swagger/", httpSwagger.WrapHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
