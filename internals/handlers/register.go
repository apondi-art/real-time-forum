package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"real-time-forum/internals/database"
)

// Handler holds dependencies like the database connection
type Handler struct {
	DB *database.Database
}

// NewHandler creates a new Handler with the given database
func NewHandler(db *database.Database) *Handler {
	return &Handler{DB: db}
}

// UserRegistration represents registration form data
type UserRegistration struct {
	Nickname  string `json:"nickname"`
	Age       int    `json:"age"`
	Gender    string `json:"gender"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}
type LoginUser struct {
	Nickname string `json:"nickname"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Response represents API response structure
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Register handles user registration
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body
	var user UserRegistration
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		sendJSONResponse(w, false, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input (basic validation)
	if user.Nickname == "" || user.Email == "" || user.Password == "" {
		sendJSONResponse(w, false, "Username, email, and password are required", http.StatusBadRequest)
		return
	}
	user_password, err := database.PasswordHashing(user.Password)
	if err != nil {
		sendJSONResponse(w, false, err.Error(), http.StatusInternalServerError)
		return
	}

	// Register the user in the database
	err = h.DB.RegisterUser(user.Nickname, user.Email, user_password, user.LastName, user.FirstName, user.Gender, user.Age)
	if err != nil {
		log.Printf("Registration error: %v", err)
		sendJSONResponse(w, false, "Registration failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Successful registration
	sendJSONResponse(w, true, "User registered successfully", http.StatusCreated)
}

// Login handles user authentication
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the request body
	var userLogin LoginUser
	err := json.NewDecoder(r.Body).Decode(&userLogin)
	if err != nil {
		sendJSONResponse(w, false, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err = h.DB.AuthenticateUser(userLogin.Nickname, userLogin.Email, userLogin.Password); err != nil {
		sendJSONResponse(w, false, "User Not Found", http.StatusNotFound)
		return
	}
	
}

// Home handles home page data
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	// Implementation will go here
	w.Write([]byte("Home endpoint"))
}

// Helper function to send JSON responses
func sendJSONResponse(w http.ResponseWriter, success bool, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := Response{
		Success: success,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}
