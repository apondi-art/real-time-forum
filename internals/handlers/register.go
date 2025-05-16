package handlers

import (
	"net/http"

	"real-time-forum/internals/database"
)

// Handler will hold the database connection
type Handler struct {
    DB *database.Database
}

// NewHandler creates a new handler with the database connection
func NewHandler(db *database.Database) *Handler {
    return &Handler{
        DB: db,
    }
}

// Register is a placeholder for the registration handler
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement registration logic
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Registration endpoint"))
}

// Login is a placeholder for the login handler
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement login logic
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Login endpoint"))
}

func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
    // TODO: Implement login logic
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Login endpoint"))
}