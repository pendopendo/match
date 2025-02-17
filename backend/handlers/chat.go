package handlers

import (
	"log"
	"net/http"
	"sync"

	"match-me/database"
	"match-me/models"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

type ClientManager struct {
	clients    map[*websocket.Conn]int
	broadcast  chan models.Message
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.Mutex
}

// Manager is the global WebSocket client manager
var Manager = &ClientManager{
	clients:    make(map[*websocket.Conn]int),
	broadcast:  make(chan models.Message),
	register:   make(chan *websocket.Conn),
	unregister: make(chan *websocket.Conn),
}

func (manager *ClientManager) Run() {
	for {
		select {
		case conn := <-manager.register:
			manager.mutex.Lock()
			manager.clients[conn] = 0
			manager.mutex.Unlock()

		case conn := <-manager.unregister:
			manager.mutex.Lock()
			if _, ok := manager.clients[conn]; ok {
				delete(manager.clients, conn)
				conn.Close()
			}
			manager.mutex.Unlock()

		case message := <-manager.broadcast:
			manager.mutex.Lock()
			// Get the users involved in this connection
			var userID1, userID2 int
			err := database.DB.QueryRow(`
				SELECT user_id_1, user_id_2
				FROM connections
				WHERE id = $1
			`, message.ConnectionID).Scan(&userID1, &userID2)

			if err != nil {
				log.Printf("error getting connection users: %v", err)
				manager.mutex.Unlock()
				continue
			}

			// Only send the message to users who are part of this connection
			for conn, userID := range manager.clients {
				if userID == userID1 || userID == userID2 {
					if err := conn.WriteJSON(message); err != nil {
						conn.Close()
						delete(manager.clients, conn)
					}
				}
			}
			manager.mutex.Unlock()
		}
	}
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	userID, _ := getUserIDFromToken(r)
	connectionID := mux.Vars(r)["connectionId"]

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	Manager.register <- conn
	Manager.mutex.Lock()
	Manager.clients[conn] = userID
	Manager.mutex.Unlock()

	go func() {
		defer func() {
			Manager.unregister <- conn
			conn.Close()
		}()

		for {
			var msg models.WebSocketMessage
			err := conn.ReadJSON(&msg)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("error: %v", err)
				}
				break
			}

			var message models.Message
			err = database.DB.QueryRow(`
				WITH new_message AS (
					INSERT INTO messages (connection_id, sender_id, content)
					VALUES ($1, $2, $3)
					RETURNING id, connection_id, sender_id, content, read, created_at
				)
				UPDATE connections
				SET last_message = $3,
					last_message_at = NOW()
				WHERE id = $1
				RETURNING (SELECT * FROM new_message)
			`, connectionID, userID, msg.Content).Scan(
				&message.ID,
				&message.ConnectionID,
				&message.SenderID,
				&message.Content,
				&message.Read,
				&message.CreatedAt,
			)

			if err != nil {
				log.Printf("error saving message: %v", err)
				continue
			}

			Manager.broadcast <- message
		}
	}()
}
