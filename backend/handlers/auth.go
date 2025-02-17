package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"match-me/database"
	"match-me/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func GetMe(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserIDFromToken(r)

	var user models.User
	err := database.DB.QueryRow(`
		SELECT id, email
		FROM users
		WHERE id = $1
	`, userID).Scan(&user.ID, &user.Email)

	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	var userID int
	err = database.DB.QueryRow(`
		INSERT INTO users (email, password)
		VALUES ($1, $2)
		RETURNING id
	`, req.Email, string(hashedPassword)).Scan(&userID)

	if err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// Create empty profile and bio
	_, err = database.DB.Exec(`
		INSERT INTO profiles (user_id)
		VALUES ($1)
	`, userID)
	if err != nil {
		http.Error(w, "Error creating profile", http.StatusInternalServerError)
		return
	}

	_, err = database.DB.Exec(`
		INSERT INTO user_bios (user_id)
		VALUES ($1)
	`, userID)
	if err != nil {
		http.Error(w, "Error creating bio", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var user models.User
	err := database.DB.QueryRow(`
		SELECT id, email, password
		FROM users
		WHERE email = $1
	`, req.Email).Scan(&user.ID, &user.Email, &user.Password)

	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Error creating token", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"token": tokenString,
	})
}