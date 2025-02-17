package main

import (
	"log"
	"net/http"

	"match-me/database"
	"match-me/handlers"

	"github.com/gorilla/mux"
)

func main() {
	// Initialize database
	database.Init()
	defer database.DB.Close()

	// Start WebSocket manager
	go handlers.Manager.Run()

	// Router setup
	r := mux.NewRouter()

	// Auth routes
	r.HandleFunc("/api/register", handlers.Register).Methods("POST")
	r.HandleFunc("/api/login", handlers.Login).Methods("POST")

	// Protected routes
	r.HandleFunc("/api/me", handlers.AuthMiddleware(handlers.GetMe)).Methods("GET")
	r.HandleFunc("/api/me/profile", handlers.AuthMiddleware(handlers.GetMyProfile)).Methods("GET")
	r.HandleFunc("/api/me/bio", handlers.AuthMiddleware(handlers.GetMyBio)).Methods("GET")
	r.HandleFunc("/api/me/profile", handlers.AuthMiddleware(handlers.UpdateProfile)).Methods("PUT")
	r.HandleFunc("/api/me/bio", handlers.AuthMiddleware(handlers.UpdateBio)).Methods("PUT")
	r.HandleFunc("/api/users/{id}", handlers.AuthMiddleware(handlers.GetUser)).Methods("GET")
	r.HandleFunc("/api/users/{id}/profile", handlers.AuthMiddleware(handlers.GetUserProfile)).Methods("GET")
	r.HandleFunc("/api/users/{id}/bio", handlers.AuthMiddleware(handlers.GetUserBio)).Methods("GET")
	r.HandleFunc("/api/recommendations", handlers.AuthMiddleware(handlers.GetRecommendations)).Methods("GET")
	r.HandleFunc("/api/connections", handlers.AuthMiddleware(handlers.GetConnections)).Methods("GET")
	r.HandleFunc("/api/connections", handlers.AuthMiddleware(handlers.CreateConnection)).Methods("POST")
	r.HandleFunc("/api/connections/{id}/messages", handlers.AuthMiddleware(handlers.GetMessages)).Methods("GET")
	r.HandleFunc("/ws/chat/{connectionId}", handlers.AuthMiddleware(handlers.HandleWebSocket))

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