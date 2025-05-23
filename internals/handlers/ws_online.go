package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type onlineUser struct {
	UserID   int
	Nickname string
	Conn     *websocket.Conn
}

var (
	onlineUsers   = make(map[int]*onlineUser) // key: userID
	onlineUsersMu sync.RWMutex
)

var onlineUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// WebSocket message for status broadcast
type StatusMessage struct {
	Type     string `json:"type"`
	UserID   int    `json:"userId"`
	Nickname string `json:"nickname"`
	Online   bool   `json:"online"`
	Time     string `json:"time"`
}

func (h *Handler) HandleOnlineWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Token is required", http.StatusUnauthorized)
		return
	}

	claims, err := h.VerifyJWTToken(token)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := onlineUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	user, err := h.DB.GetUserByID(int(claims.UserID))
	if err != nil {
		log.Println("Failed to get user:", err)
		conn.Close()
		return
	}

	userID := int(claims.UserID)

	onlineUsersMu.Lock()
	onlineUsers[userID] = &onlineUser{UserID: userID, Nickname: user.Nickname, Conn: conn}
	onlineUsersMu.Unlock()

	broadcastUserStatus(userID, user.Nickname, true)

	defer func() {
		conn.Close()
		onlineUsersMu.Lock()
		delete(onlineUsers, userID)
		onlineUsersMu.Unlock()
		broadcastUserStatus(userID, user.Nickname, false)
	}()

	for {
		if _, _, err := conn.NextReader(); err != nil {
			break
		}
	}
}

func (h *Handler) OnlineWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "WebSocket upgrade failed", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Basic keep-alive logic or online user tracking
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break // Disconnect
		}
	}
}

func broadcastUserStatus(userID int, nickname string, online bool) {
	status := StatusMessage{
		Type:     "status",
		UserID:   userID,
		Nickname: nickname,
		Online:   online,
		Time:     time.Now().Format(time.RFC3339),
	}
	data, _ := json.Marshal(status)

	onlineUsersMu.RLock()
	defer onlineUsersMu.RUnlock()
	for _, user := range onlineUsers {
		user.Conn.WriteMessage(websocket.TextMessage, data)
	}
}
