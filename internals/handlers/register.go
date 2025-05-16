package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"real-time-forum/internals/database"

	"github.com/golang-jwt/jwt/v5"
)

type Handler struct {
	DB *database.Database
}

func NewHandler(db *database.Database) *Handler {
	return &Handler{DB: db}
}

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
    Identifier string `json:"identifier"`
    Password   string `json:"password"`
}


type AuthResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Token   string      `json:"token,omitempty"`
	User    *database.User `json:"user,omitempty"`
}

// JWT secret key (in production, use environment variable)
var jwtSecret = []byte("your-strong-secret-key")

type Claims struct {
	UserID int `json:"userId"`
	jwt.RegisteredClaims
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user UserRegistration
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		sendAuthResponse(w, false, "Invalid request body", http.StatusBadRequest, nil, "")
		return
	}

	// Validate input
	if user.Nickname == "" || user.Email == "" || user.Password == "" {
		sendAuthResponse(w, false, "All fields are required", http.StatusBadRequest, nil, "")
		return
	}

	// Hash password
	hashedPassword, err := database.PasswordHashing(user.Password)
	if err != nil {
		sendAuthResponse(w, false, "Failed to secure password", http.StatusInternalServerError, nil, "")
		return
	}

	// Register user
	err = h.DB.RegisterUser(user.Nickname, user.Email, hashedPassword, user.LastName, user.FirstName, user.Gender, user.Age)
	if err != nil {
		log.Printf("Registration error: %v", err)
		message := "Registration failed"
		if err.Error() == "username or email already exists" {
			message = err.Error()
		}
		sendAuthResponse(w, false, message, http.StatusBadRequest, nil, "")
		return
	}

	// Authenticate the new user to get user details
	authUser, err := h.DB.AuthenticateUser(user.Nickname, user.Email, user.Password)
	if err != nil {
		sendAuthResponse(w, false, "Registration complete but login failed", http.StatusOK, nil, "")
		return
	}

	// Generate JWT token
	token, err := generateJWTToken(authUser.ID)
	if err != nil {
		sendAuthResponse(w, false, "Token generation failed", http.StatusInternalServerError, nil, "")
		return
	}

	sendAuthResponse(w, true, "Registration successful", http.StatusCreated, authUser, token)
}

// Debugged version of Login handler with detailed logging

// Complete updated handler using this struct
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    
    var userLogin LoginUser
    if err := json.NewDecoder(r.Body).Decode(&userLogin); err != nil {
        sendAuthResponse(w, false, "Invalid request body", http.StatusBadRequest, nil, "")
        return
    }
    
    // Validate required fields
    if userLogin.Identifier == "" {
        sendAuthResponse(w, false, "Username or email is required", http.StatusBadRequest, nil, "")
        return
    }
    
    if userLogin.Password == "" {
        sendAuthResponse(w, false, "Password is required", http.StatusBadRequest, nil, "")
        return
    }
    
    // Determine if identifier is email or nickname
    var nickname, email string
    if strings.Contains(userLogin.Identifier, "@") {
        email = userLogin.Identifier
    } else {
        nickname = userLogin.Identifier
    }
    
    // Authenticate user
    user, err := h.DB.AuthenticateUser(nickname, email, userLogin.Password)
    if err != nil {
        status := http.StatusUnauthorized
        if err.Error() == "user not found" {
            status = http.StatusNotFound
        } else if err.Error() == "no credentials provided" {
            status = http.StatusBadRequest
        }
        sendAuthResponse(w, false, err.Error(), status, nil, "")
        return
    }
    
    // Generate JWT token
    token, err := generateJWTToken(user.ID)
    if err != nil {
        sendAuthResponse(w, false, "Token generation failed", http.StatusInternalServerError, nil, "")
        return
    }
    
    sendAuthResponse(w, true, "Login successful", http.StatusOK, user, token)
}

func generateJWTToken(userID int) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "real-time-forum",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}



// Home handles home page data
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	// Implementation will go here
	w.Write([]byte("Home endpoint"))
}

// Helper function to send JSON responses
func sendAuthResponse(w http.ResponseWriter, success bool, message string, statusCode int, user *database.User, token string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := AuthResponse{
		Success: success,
		Message: message,
		Token:   token,
		User:    user,
	}

	json.NewEncoder(w).Encode(response)
}
