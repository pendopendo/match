package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"match-me/database"
	"match-me/models"

	"github.com/gorilla/mux"
)

func GetConnections(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserIDFromToken(r)

	rows, err := database.DB.Query(`
		WITH connection_info AS (
			SELECT 
				c.id,
				c.user_id_1,
				c.user_id_2,
				c.last_message,
				c.last_message_at,
				CASE 
					WHEN c.user_id_1 = $1 THEN c.user_id_2
					ELSE c.user_id_1
				END as other_user_id,
				(
					SELECT COUNT(*)
					FROM messages m
					WHERE m.connection_id = c.id
					AND m.sender_id != $1
					AND NOT m.read
				) as unread_count
			FROM connections c
			WHERE c.user_id_1 = $1 OR c.user_id_2 = $1
		)
		SELECT 
			ci.*,
			p.name,
			p.profile_picture
		FROM connection_info ci
		LEFT JOIN profiles p ON p.user_id = ci.other_user_id
		ORDER BY ci.last_message_at DESC NULLS LAST
	`, userID)

	if err != nil {
		http.Error(w, "Error fetching connections", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var connections []map[string]interface{}
	for rows.Next() {
		var conn struct {
			ID            int
			UserID1       int
			UserID2       int
			LastMessage   string
			LastMessageAt string
			OtherUserID   int
			UnreadCount   int
			Name          string
			ProfilePicture string
		}

		err := rows.Scan(
			&conn.ID,
			&conn.UserID1,
			&conn.UserID2,
			&conn.LastMessage,
			&conn.LastMessageAt,
			&conn.OtherUserID,
			&conn.UnreadCount,
			&conn.Name,
			&conn.ProfilePicture,
		)

		if err != nil {
			http.Error(w, "Error scanning connections", http.StatusInternalServerError)
			return
		}

		connection := map[string]interface{}{
			"id":              conn.ID,
			"other_user_id":   conn.OtherUserID,
			"unread_count":    conn.UnreadCount,
			"name":            conn.Name,
			"profile_picture": conn.ProfilePicture,
			"last_message":    conn.LastMessage,
			"last_message_at": conn.LastMessageAt,
		}

		connections = append(connections, connection)
	}

	json.NewEncoder(w).Encode(connections)
}

func CreateConnection(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserIDFromToken(r)

	var req struct {
		UserID int `json:"user_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == userID {
		http.Error(w, "Cannot create connection with yourself", http.StatusBadRequest)
		return
	}

	var connectionID int
	err := database.DB.QueryRow(`
		INSERT INTO connections (user_id_1, user_id_2)
		VALUES ($1, $2)
		ON CONFLICT (user_id_1, user_id_2) DO NOTHING
		RETURNING id
	`, userID, req.UserID).Scan(&connectionID)

	if err != nil {
		// If no rows were affected, check if connection exists in reverse order
		err = database.DB.QueryRow(`
			SELECT id FROM connections
			WHERE (user_id_1 = $2 AND user_id_2 = $1)
		`, userID, req.UserID).Scan(&connectionID)

		if err != nil {
			http.Error(w, "Failed to create connection", http.StatusInternalServerError)
			return
		}
	}

	json.NewEncoder(w).Encode(map[string]int{
		"connection_id": connectionID,
	})
}

func GetMessages(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserIDFromToken(r)
	params := mux.Vars(r)
	connectionID, err := strconv.Atoi(params["id"])
	if err != nil {
		http.Error(w, "Invalid connection ID", http.StatusBadRequest)
		return
	}

	// Verify user is part of the connection
	var count int
	err = database.DB.QueryRow(`
		SELECT COUNT(*) FROM connections 
		WHERE id = $1 AND (user_id_1 = $2 OR user_id_2 = $2)
	`, connectionID, userID).Scan(&count)

	if err != nil || count == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Mark messages as read
	_, err = database.DB.Exec(`
		UPDATE messages
		SET read = true
		WHERE connection_id = $1
		AND sender_id != $2
		AND NOT read
	`, connectionID, userID)

	if err != nil {
		http.Error(w, "Error marking messages as read", http.StatusInternalServerError)
		return
	}

	// Fetch messages
	rows, err := database.DB.Query(`
		SELECT id, connection_id, sender_id, content, read, created_at
		FROM messages
		WHERE connection_id = $1
		ORDER BY created_at ASC
	`, connectionID)

	if err != nil {
		http.Error(w, "Error fetching messages", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []models.Message
	for rows.Next() {
		var msg models.Message
		err := rows.Scan(
			&msg.ID,
			&msg.ConnectionID,
			&msg.SenderID,
			&msg.Content,
			&msg.Read,
			&msg.CreatedAt,
		)
		if err != nil {
			http.Error(w, "Error scanning messages", http.StatusInternalServerError)
			return
		}
		messages = append(messages, msg)
	}

	json.NewEncoder(w).Encode(messages)
}