package main

import (
	"log"
	"net/http"
	"os"

	"real-time-forum/internals/database"
	"real-time-forum/internals/handlers"

	"github.com/gorilla/mux"
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

	// Create a new router
	r := mux.NewRouter()

	// Serve static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))

	// API Routes
	r.HandleFunc("/api/register", handler.Register).Methods("POST")
	r.HandleFunc("/api/login", handler.Login).Methods("POST")
	r.HandleFunc("/api/categories", handler.GetCategories).Methods("GET")
	r.HandleFunc("/api/home", handler.Home).Methods("GET")

	// Post routes
	r.HandleFunc("/api/posts", handler.GetPosts).Methods("GET")
	r.HandleFunc("/api/posts", handler.CreatePost).Methods("POST")
	r.HandleFunc("/api/posts/{id}", handler.GetPosts).Methods("GET") // Route for specific post

	// Comment routes under a specific post
	r.HandleFunc("/api/posts/{postId}/comments", handler.GetComments).Methods("GET")
	r.HandleFunc("/api/posts/{postId}/comments", handler.AddComment).Methods("POST")

	r.HandleFunc("/api/logout", handler.Logout).Methods("POST")

	// User status routes
	r.HandleFunc("/api/online-users", handler.GetOnlineUsers).Methods("GET")
	r.HandleFunc("/api/online-status", handler.UpdateOnlineStatus).Methods("POST")

	// Serve the index.html for other routes (SPA handling)
	r.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./template/index.html")
	})

	// Start the server
	port := "8080"
	log.Printf("Server listening on port %s", port)
	err = http.ListenAndServe(":"+port, r) // Use the router
	if err != nil {
		log.Fatalf("ListenAndServe error: %v", err)
	}
}