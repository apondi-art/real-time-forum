package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"real-time-forum/internals/database"
	"real-time-forum/internals/handlers"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Initialize the database
	db, err := database.New("./internals/database/real_time.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	handler := handlers.NewHandler(db)

	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// WebSocket endpoint for online users
	http.HandleFunc("/ws/online", handler.OnlineWebSocket)

	http.HandleFunc("/ws/chat", handler.HandleChatWebSocket)

	// API routes
	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		apiRouter(w, r, handler)
	})

	// Serve SPA fallback (for React/Vue apps etc.)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./template/index.html")
	})

	port := "8080"
	log.Printf("Server listening on port %s", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}

func apiRouter(w http.ResponseWriter, r *http.Request, h *handlers.Handler) {
	path := strings.TrimPrefix(r.URL.Path, "/api")
	method := r.Method

	switch {
	case path == "/register" && method == http.MethodPost:
		h.Register(w, r)
	case path == "/login" && method == http.MethodPost:
		h.Login(w, r)
	case path == "/categories" && method == http.MethodGet:
		h.GetCategories(w, r)
	case path == "/home" && method == http.MethodGet:
		h.Home(w, r)
	case path == "/posts" && method == http.MethodGet:
		h.GetPosts(w, r)
	case path == "/posts" && method == http.MethodPost:
		h.CreatePost(w, r)
	case strings.HasPrefix(path, "/posts/") && strings.HasSuffix(path, "/comments"):
		parts := strings.Split(path, "/")
		if len(parts) >= 3 {
			postID := parts[2]
			if method == http.MethodGet {
				h.GetComments(w, r, postID)
			} else if method == http.MethodPost {
				h.AddComment(w, r, postID)
			}
		}
	case strings.HasPrefix(path, "/posts/") && method == http.MethodGet:
		postID := strings.TrimPrefix(path, "/posts/")
		h.GetPostByID(w, r, postID)
	case path == "/logout" && method == http.MethodPost:
		h.Logout(w, r)
	case path == "/online-users" && method == http.MethodGet:
		h.GetOnlineUsers(w, r)
	case path == "/online-status" && method == http.MethodPost:
		h.UpdateOnlineStatus(w, r)
	case path == "/chat/ws" && method == http.MethodGet:
		h.HandleChatWebSocket(w, r)
	case path == "/validate-token" && method == http.MethodGet:
		h.ValidateToken(w, r)

	default:
		http.NotFound(w, r)
	}
}
