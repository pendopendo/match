package models

import (
	"time"
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

type Connection struct {
	ID            int       `json:"id"`
	UserID1       int       `json:"user_id_1"`
	UserID2       int       `json:"user_id_2"`
	LastMessage   string    `json:"last_message"`
	LastMessageAt time.Time `json:"last_message_at"`
	UnreadCount   int       `json:"unread_count"`
}

type Message struct {
	ID           int       `json:"id"`
	ConnectionID int       `json:"connection_id"`
	SenderID     int       `json:"sender_id"`
	Content      string    `json:"content"`
	Read         bool      `json:"read"`
	CreatedAt    time.Time `json:"created_at"`
}

type WebSocketMessage struct {
	ConnectionID int    `json:"connection_id"`
	Content      string `json:"content"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}