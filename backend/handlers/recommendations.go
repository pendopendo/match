package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"match-me/database"
)

func GetRecommendations(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserIDFromToken(r)

	rows, err := database.DB.Query(`
		WITH user_interests AS (
			SELECT unnest(interests) as interest
			FROM user_bios
			WHERE user_id = $1
		),
		matching_users AS (
			SELECT DISTINCT ub.user_id
			FROM user_bios ub
			JOIN user_interests ui ON ui.interest = ANY(ub.interests)
			WHERE ub.user_id != $1
			AND ub.user_id NOT IN (
				SELECT user_id_2 FROM connections WHERE user_id_1 = $1
				UNION
				SELECT user_id_1 FROM connections WHERE user_id_2 = $1
			)
			LIMIT 10
		)
		SELECT 
			p.user_id,
			p.name,
			p.bio,
			p.profile_picture,
			p.location,
			ub.interests
		FROM matching_users mu
		JOIN profiles p ON p.user_id = mu.user_id
		JOIN user_bios ub ON ub.user_id = mu.user_id
	`, userID)

	if err != nil {
		http.Error(w, "Error fetching recommendations", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var recommendations []map[string]interface{}
	for rows.Next() {
		var profile struct {
			UserID         int
			Name           string
			Bio            string
			ProfilePicture sql.NullString
			Location       sql.NullString
			Interests      []string
		}

		err := rows.Scan(
			&profile.UserID,
			&profile.Name,
			&profile.Bio,
			&profile.ProfilePicture,
			&profile.Location,
			&profile.Interests,
		)

		if err != nil {
			http.Error(w, "Error scanning recommendations", http.StatusInternalServerError)
			return
		}

		recommendation := map[string]interface{}{
			"user_id":   profile.UserID,
			"name":      profile.Name,
			"bio":       profile.Bio,
			"interests": profile.Interests,
		}

		if profile.ProfilePicture.Valid {
			recommendation["profile_picture"] = profile.ProfilePicture.String
		}
		if profile.Location.Valid {
			recommendation["location"] = profile.Location.String
		}

		recommendations = append(recommendations, recommendation)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"recommendations": recommendations,
	})
}
