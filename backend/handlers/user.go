package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"match-me/database"
	"match-me/models"

	"github.com/gorilla/mux"
)

func GetUser(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var user models.User
	err = database.DB.QueryRow(`
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

func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var profile models.Profile
	err = database.DB.QueryRow(`
		SELECT user_id, name, bio, profile_picture, location
		FROM profiles
		WHERE user_id = $1
	`, userID).Scan(
		&profile.UserID,
		&profile.Name,
		&profile.Bio,
		&profile.ProfilePicture,
		&profile.Location,
	)

	if err != nil {
		http.Error(w, "Profile not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(profile)
}

func GetUserBio(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	userID, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var bio models.UserBio
	err = database.DB.QueryRow(`
		SELECT user_id, interests, hobbies, music_preferences, food_preferences, looking_for
		FROM user_bios
		WHERE user_id = $1
	`, userID).Scan(
		&bio.UserID,
		&bio.Interests,
		&bio.Hobbies,
		&bio.MusicPreferences,
		&bio.FoodPreferences,
		&bio.LookingFor,
	)

	if err != nil {
		http.Error(w, "Bio not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(bio)
}