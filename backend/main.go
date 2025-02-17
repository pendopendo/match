package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password string `json:"-"` // Never send password in JSON
}

type Profile struct {
	UserID         int    `json:"user_id"`
	Name           string `json:"name"`
	Bio            string `json:"bio"`
	ProfilePicture string `json:"profile_picture"`
	Location       string `json:"location"`
}

type UserBio struct {
	UserID           int      `json:"user_id"`
	Interests        []string `json:"interests"`
	Hobbies          []string `json:"hobbies"`
	MusicPreferences []string `json:"music_preferences"`
	FoodPreferences  []string `json:"food_preferences"`
	LookingFor       []string `json:"looking_for"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

var db *sql.DB
var jwtKey = []byte(os.Getenv("JWT_SECRET"))

func main() {
	// Database connection
	connStr := "postgres://postgres:postgres@localhost/matchme?sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create tables if they don't exist
	createTables()

	// Router setup
	r := mux.NewRouter()

	// Auth routes
	r.HandleFunc("/api/register", register).Methods("POST")
	r.HandleFunc("/api/login", login).Methods("POST")

	// Protected routes
	r.HandleFunc("/api/me", authMiddleware(getMe)).Methods("GET")
	r.HandleFunc("/api/me/profile", authMiddleware(getMyProfile)).Methods("GET")
	r.HandleFunc("/api/me/bio", authMiddleware(getMyBio)).Methods("GET")
	r.HandleFunc("/api/me/profile", authMiddleware(updateProfile)).Methods("PUT")
	r.HandleFunc("/api/me/bio", authMiddleware(updateBio)).Methods("PUT")
	r.HandleFunc("/api/users/{id}", authMiddleware(getUser)).Methods("GET")
	r.HandleFunc("/api/users/{id}/profile", authMiddleware(getUserProfile)).Methods("GET")
	r.HandleFunc("/api/users/{id}/bio", authMiddleware(getUserBio)).Methods("GET")
	r.HandleFunc("/api/recommendations", authMiddleware(getRecommendations)).Methods("GET")

	// CORS middleware
	corsMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	// Apply CORS middleware
	handler := corsMiddleware(r)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}

func createTables() {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS profiles (
			user_id INTEGER PRIMARY KEY REFERENCES users(id),
			name VARCHAR(255),
			bio TEXT,
			profile_picture TEXT,
			location VARCHAR(100),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS user_bio (
			user_id INTEGER PRIMARY KEY REFERENCES users(id),
			interests TEXT[],
			hobbies TEXT[],
			music_preferences TEXT[],
			food_preferences TEXT[],
			looking_for TEXT[],
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS connections (
			id SERIAL PRIMARY KEY,
			user_id_1 INTEGER REFERENCES users(id),
			user_id_2 INTEGER REFERENCES users(id),
			status VARCHAR(20) CHECK (status IN ('pending', 'accepted', 'rejected')),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id_1, user_id_2)
		)`,
		`CREATE TABLE IF NOT EXISTS messages (
			id SERIAL PRIMARY KEY,
			connection_id INTEGER REFERENCES connections(id),
			sender_id INTEGER REFERENCES users(id),
			content TEXT NOT NULL,
			read BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, query := range queries {
		_, err := db.Exec(query)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error hashing password", http.StatusInternalServerError)
		return
	}

	// Insert user
	_, err = db.Exec("INSERT INTO users (email, password) VALUES ($1, $2)", req.Email, hashedPassword)
	if err != nil {
		http.Error(w, "Email already exists", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user User
	err := db.QueryRow("SELECT id, email, password FROM users WHERE email = $1", req.Email).
		Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Create token
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

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Remove "Bearer " prefix
		tokenStr = tokenStr[7:]

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Add user ID to request context
		r = r.WithContext(r.Context())
		next(w, r)
	}
}

func getUserIDFromToken(r *http.Request) (int, error) {
	tokenStr := r.Header.Get("Authorization")[7:] // Remove "Bearer " prefix
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil || !token.Valid {
		return 0, err
	}

	return claims.UserID, nil
}

func getMe(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var user User
	err = db.QueryRow("SELECT id, email FROM users WHERE id = $1", userID).
		Scan(&user.ID, &user.Email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func getMyProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	profile, err := getProfileByUserID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Profile not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(profile)
}

func getMyBio(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	bio, err := getBioByUserID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Bio not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(bio)
}

func updateProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var profile Profile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	profile.UserID = userID

	_, err = db.Exec(`
		INSERT INTO profiles (user_id, name, bio, profile_picture, location)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			name = $2,
			bio = $3,
			profile_picture = $4,
			location = $5,
			updated_at = CURRENT_TIMESTAMP
	`, profile.UserID, profile.Name, profile.Bio, profile.ProfilePicture, profile.Location)

	if err != nil {
		http.Error(w, "Error updating profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func updateBio(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var bio UserBio
	if err := json.NewDecoder(r.Body).Decode(&bio); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	bio.UserID = userID

	_, err = db.Exec(`
		INSERT INTO user_bio (user_id, interests, hobbies, music_preferences, food_preferences, looking_for)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) 
		DO UPDATE SET 
			interests = $2,
			hobbies = $3,
			music_preferences = $4,
			food_preferences = $5,
			looking_for = $6,
			updated_at = CURRENT_TIMESTAMP
	`, bio.UserID, bio.Interests, bio.Hobbies, bio.MusicPreferences, bio.FoodPreferences, bio.LookingFor)

	if err != nil {
		http.Error(w, "Error updating bio", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var user User
	err = db.QueryRow("SELECT id, email FROM users WHERE id = $1", userID).
		Scan(&user.ID, &user.Email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"id": user.ID,
		// Email is intentionally omitted for privacy
	})
}

func getUserProfile(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	profile, err := getProfileByUserID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Profile not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(profile)
}

func getUserBio(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	bio, err := getBioByUserID(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Bio not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(bio)
}

func getProfileByUserID(userID int) (*Profile, error) {
	var profile Profile
	err := db.QueryRow(`
		SELECT user_id, name, bio, profile_picture, location 
		FROM profiles 
		WHERE user_id = $1
	`, userID).Scan(&profile.UserID, &profile.Name, &profile.Bio, &profile.ProfilePicture, &profile.Location)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func getBioByUserID(userID int) (*UserBio, error) {
	var bio UserBio
	err := db.QueryRow(`
		SELECT user_id, interests, hobbies, music_preferences, food_preferences, looking_for 
		FROM user_bio 
		WHERE user_id = $1
	`, userID).Scan(&bio.UserID, &bio.Interests, &bio.Hobbies, &bio.MusicPreferences, &bio.FoodPreferences, &bio.LookingFor)
	if err != nil {
		return nil, err
	}
	return &bio, nil
}

func getRecommendations(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromToken(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Check if user has completed their profile
	profile, err := getProfileByUserID(userID)
	if err != nil {
		http.Error(w, "Please complete your profile first", http.StatusBadRequest)
		return
	}

////////////////////////	bio, err := getBioByUserID(userID)
	if err != nil {
		http.Error(w, "Please complete your bio first", http.StatusBadRequest)
		return
	}

	// Get users with matching location and calculate compatibility score
	rows, err := db.Query(`
		WITH potential_matches AS (
			SELECT 
				u.id,
				p.location,
				b.interests,
				b.hobbies,
				b.music_preferences,
				b.food_preferences,
				b.looking_for
			FROM users u
			JOIN profiles p ON u.id = p.user_id
			JOIN user_bio b ON u.id = b.user_id
			WHERE u.id != $1
			AND p.location = $2
			AND NOT EXISTS (
				SELECT 1 FROM connections c
				WHERE (c.user_id_1 = $1 AND c.user_id_2 = u.id)
				OR (c.user_id_1 = u.id AND c.user_id_2 = $1)
			)
		)
		SELECT id FROM potential_matches
		LIMIT 10
	`, userID, profile.Location)

	if err != nil {
		http.Error(w, "Error fetching recommendations", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var recommendations []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			http.Error(w, "Error scanning recommendations", http.StatusInternalServerError)
			return
		}
		recommendations = append(recommendations, id)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"recommendations": recommendations,
	})
}
