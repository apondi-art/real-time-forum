package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"real-time-forum/internals/database"
	"real-time-forum/internals/handlers"

	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
)

func main() {
	// Initialize the database
	db, err := database.New("./internals/database/real_time.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	// Initialize handlers with the database connection
	handler := handlers.NewHandler(db)

	// Serve static files with correct MIME type
	fileServer := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasPrefix(path, "/static/") {
			name := filepath.Join("./static", strings.TrimPrefix(path, "/static/"))
			http.ServeFile(w, r, name)
			return
		}
		// Serve index.html for other routes
		http.ServeFile(w, r, "./template/index.html")
	})
	http.Handle("/", fileServer)

	// Define your API routes
	http.HandleFunc("/api/register", handler.Register)
	http.HandleFunc("/api/login", handler.Login)
	http.HandleFunc("/api/categories", handler.GetCategories)
	http.HandleFunc("api/home", handler.Home)
	http.HandleFunc("/api/posts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetPosts(w, r)
		case http.MethodPost:
			handler.CreatePost(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	http.HandleFunc("/api/logout", handler.Logout)
	

	// Start the server
	port := "8080" // Choose a port
	log.Printf("Server listening on port %s", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}