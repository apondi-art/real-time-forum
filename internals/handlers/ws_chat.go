package handlers

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, restrict this
	},
}

type ChatClient struct {
	Conn   *websocket.Conn
	UserID int
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
	broadcast:  make(chan []byte),
	register:   make(chan *ChatClient),
	unregister: make(chan *ChatClient),
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

	client := &ChatClient{Conn: conn, UserID: int(claims.UserID)}
	chatHub.register <- client

	go handleChatMessages(client)
}

func handleChatMessages(client *ChatClient) {
	defer func() {
		chatHub.unregister <- client
	}()
	for {
		_, message, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}
		chatHub.broadcast <- message
	}
}

func init() {
	go func() {
		for {
			select {
			case client := <-chatHub.register:
				chatHub.mu.Lock()
				chatHub.clients[client.UserID] = client
				chatHub.mu.Unlock()

			case client := <-chatHub.unregister:
				chatHub.mu.Lock()
				if _, ok := chatHub.clients[client.UserID]; ok {
					client.Conn.Close()
					delete(chatHub.clients, client.UserID)
				}
				chatHub.mu.Unlock()

			case message := <-chatHub.broadcast:
				chatHub.mu.Lock()
				for _, client := range chatHub.clients {
					client.Conn.WriteMessage(websocket.TextMessage, message)
				}
				chatHub.mu.Unlock()
			}
		}
	}()
}
