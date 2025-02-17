package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"match-me/database"
	"match-me/models"
)

func GetMyProfile(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserIDFromToken(r)

	var profile models.Profile
	err := database.DB.QueryRow(`
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
		log.Printf("Profile not found for user %d, creating new profile", userID)

		// Create empty profile
		_, err = database.DB.Exec(`
			INSERT INTO profiles (user_id, name, bio, profile_picture, location)
			VALUES ($1, '', '', '', '')
			ON CONFLICT (user_id) DO NOTHING
		`, userID)

		if err != nil {
			log.Printf("Error creating profile: %v", err)
			http.Error(w, "Failed to create profile", http.StatusInternalServerError)
			return
		}

		// Initialize empty profile
		profile = models.Profile{
			UserID:         userID,
			Name:           "",
			Bio:            "",
			ProfilePicture: "",
			Location:       "",
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}

func GetMyBio(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserIDFromToken(r)

	var bio models.UserBio
	err := database.DB.QueryRow(`
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
		log.Printf("Bio not found for user %d, creating new bio", userID)

		// Create empty bio
		_, err = database.DB.Exec(`
			INSERT INTO user_bios (user_id, interests, hobbies, music_preferences, food_preferences, looking_for)
			VALUES ($1, ARRAY[]::TEXT[], ARRAY[]::TEXT[], ARRAY[]::TEXT[], ARRAY[]::TEXT[], ARRAY[]::TEXT[])
			ON CONFLICT (user_id) DO NOTHING
		`, userID)

		if err != nil {
			log.Printf("Error creating bio: %v", err)
			http.Error(w, "Failed to create bio", http.StatusInternalServerError)
			return
		}

		// Initialize empty bio
		bio = models.UserBio{
			UserID:           userID,
			Interests:        []string{},
			Hobbies:          []string{},
			MusicPreferences: []string{},
			FoodPreferences:  []string{},
			LookingFor:       []string{},
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bio)
}

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserIDFromToken(r)

	var profile models.Profile
	if err := json.NewDecoder(r.Body).Decode(&profile); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec(`
		UPDATE profiles
		SET name = $1, bio = $2, profile_picture = $3, location = $4, updated_at = NOW()
		WHERE user_id = $5
	`,
		profile.Name,
		profile.Bio,
		profile.ProfilePicture,
		profile.Location,
		userID,
	)

	if err != nil {
		http.Error(w, "Error updating profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func UpdateBio(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserIDFromToken(r)

	var bio models.UserBio
	if err := json.NewDecoder(r.Body).Decode(&bio); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	_, err := database.DB.Exec(`
		UPDATE user_bios
		SET interests = $1, hobbies = $2, music_preferences = $3, food_preferences = $4, looking_for = $5, updated_at = NOW()
		WHERE user_id = $6
	`,
		bio.Interests,
		bio.Hobbies,
		bio.MusicPreferences,
		bio.FoodPreferences,
		bio.LookingFor,
		userID,
	)

	if err != nil {
		http.Error(w, "Error updating bio", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
