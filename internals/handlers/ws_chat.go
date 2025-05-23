package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, restrict the domains
	},
}

type ChatClient struct {
	Conn     *websocket.Conn
	UserID   int
	Username string
}

type ChatHub struct {
	clients    map[int]*ChatClient
	broadcast  chan []byte
	register   chan *ChatClient
	unregister chan *ChatClient
	mu         sync.Mutex
}

var chatHub = ChatHub{
	clients:    make(map[int]*ChatClient),
	broadcast:  make(chan []byte, 100), // Buffered channel
	register:   make(chan *ChatClient),
	unregister: make(chan *ChatClient),
}

type WebSocketMessage struct {
	Type           string    `json:"type"`
	Content        string    `json:"content"`
	SenderID       int       `json:"senderId"`
	SenderUsername string    `json:"senderUsername"`
	RecipientID    int       `json:"recipientId"`
	Timestamp      time.Time `json:"timestamp,omitempty"`
}

func (h *Handler) HandleChatWebSocket(w http.ResponseWriter, r *http.Request) {
	tokenString := r.Header.Get("Sec-WebSocket-Protocol")
	claims, err := h.VerifyJWTToken(tokenString)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, http.Header{"Sec-WebSocket-Protocol": {tokenString}})
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	username, err := h.DB.GetUsername(int(claims.UserID))
	if err != nil {
		username = "User" // Default if username lookup fails
	}

	client := &ChatClient{
		Conn:     conn,
		UserID:   int(claims.UserID),
		Username: username,
	}

	// Register client
	chatHub.register <- client

	// Configure connection
	conn.SetReadLimit(1024) // 1KB max message size
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Start reader and writer goroutines
	go h.readPump(client)
	go h.writePump(client)
}

func (h *Handler) readPump(client *ChatClient) {
	defer func() {
		chatHub.unregister <- client
		client.Conn.Close()
	}()

	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		var msg WebSocketMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Invalid message format: %v", err)
			continue
		}

		// Validate message
		if msg.Type != "chat" || msg.Content == "" || msg.RecipientID <= 0 {
			log.Printf("Invalid message: %+v", msg)
			continue
		}

		// Set sender info from connection
		msg.SenderID = client.UserID
		msg.SenderUsername = client.Username
		msg.Timestamp = time.Now()

		// Save to database
		if _, err := h.DB.SendMessage(msg.SenderID, msg.RecipientID, msg.Content); err != nil {
			log.Printf("Error saving message: %v", err)
			continue
		}

		// Before broadcasting:
		messageBytes, err := json.Marshal(map[string]interface{}{
			"sender_id":       msg.SenderID,
			"sender_username": msg.SenderUsername,
			"recipient_id":    msg.RecipientID,
			"content":         msg.Content,
			"created_at":      msg.Timestamp.Format(time.RFC3339), // Format timestamp consistently
		})
		if err != nil {
			log.Printf("Error marshaling message: %v", err)
			continue
		}
		chatHub.broadcast <- messageBytes
	}
}

func (h *Handler) writePump(client *ChatClient) {
	ticker := time.NewTicker(30 * time.Second) // Ping interval
	defer func() {
		ticker.Stop()
		client.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-chatHub.broadcast:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub closed the channel
				client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			var msg WebSocketMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Error decoding broadcast message: %v", err)
				continue
			}

			// Only send if client is recipient or sender (for echo)
			if msg.RecipientID == client.UserID || msg.SenderID == client.UserID {
				if err := client.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
					log.Printf("Write error: %v", err)
					return
				}
			}

		case <-ticker.C:
			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *ChatHub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			log.Printf("User connected: %d (%s)", client.UserID, client.Username)
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				log.Printf("User disconnected: %d", client.UserID)
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			h.mu.Lock()
			var msg WebSocketMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Error decoding broadcast message: %v", err)
				h.mu.Unlock()
				continue
			}

			// Only send to recipient (no echo to sender)
			if recipient, ok := h.clients[msg.RecipientID]; ok {
				log.Printf("Routing message from %d to %d", msg.SenderID, msg.RecipientID)
				if err := recipient.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
					log.Printf("Error sending to recipient %d: %v", msg.RecipientID, err)
				}
			}
			h.mu.Unlock()
		}
	}
}

func init() {
	go chatHub.run()
}
