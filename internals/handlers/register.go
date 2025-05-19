package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"real-time-forum/internals/database"
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
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Token   string         `json:"token,omitempty"`
	User    *database.User `json:"user,omitempty"`
}

type NewPostRequest struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	Category  string `json:"category"` // Added category field
}

type PostResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	PostID  int    `json:"postId,omitempty"`
}

// Category response types
type CategoryResponse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	PostCount   int    `json:"postCount"`
}

// JWT secret key (in production, use environment variable)
var jwtSecret = []byte("your-strong-secret-key")

type Claims struct {
	UserID    int64  `json:"userId"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	Issuer    string `json:"iss"`
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
	token, err := h.generateJWTToken(int64(authUser.ID))
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
	token, err := h.generateJWTToken(int64(user.ID))
	if err != nil {
		sendAuthResponse(w, false, "Token generation failed", http.StatusInternalServerError, nil, "")
		return
	}

	sendAuthResponse(w, true, "Login successful", http.StatusOK, user, token)
}

func (h *Handler) generateJWTToken(userID int64) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:    userID,
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: expirationTime.Unix(),
		Issuer:    "real-time-forum",
	}

	// Encode the header
	header := map[string]string{"alg": "HS256", "typ": "JWT"}
	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	headerEncoded := base64Encode(headerJSON)

	// Encode the payload (claims)
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payloadEncoded := base64Encode(payloadJSON)

	// Create the signature
	unsignedToken := headerEncoded + "." + payloadEncoded
	hashing := hmac.New(sha256.New, jwtSecret)
	hashing.Write([]byte(unsignedToken))
	signature := base64Encode(hashing.Sum(nil))

	// Combine the parts
	token := unsignedToken + "." + signature
	return token, nil
}

func (h *Handler) VerifyJWTToken(tokenString string) (*Claims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	headerEncoded := parts[0]
	payloadEncoded := parts[1]
	signature := parts[2]

	// Verify the signature
	unsignedToken := headerEncoded + "." + payloadEncoded
	hashing := hmac.New(sha256.New, jwtSecret)
	hashing.Write([]byte(unsignedToken))
	expectedSignature := base64Encode(hashing.Sum(nil))

	if signature != expectedSignature {
		return nil, fmt.Errorf("invalid token signature")
	}

	// Decode the payload
	payloadBytes, err := base64Decode(payloadEncoded)
	if err != nil {
		return nil, fmt.Errorf("invalid payload encoding: %v", err)
	}

	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %v", err)
	}

	// Basic expiry check
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, fmt.Errorf("token has expired")
	}

	return &claims, nil
}

func base64Encode(src []byte) string {
	return base64.RawURLEncoding.EncodeToString(src)
}

func base64Decode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
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

// Updated CreatePost handler with category support
func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify JWT token to get user ID
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Remove "Bearer " prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	claims, err := h.VerifyJWTToken(tokenString)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userID := int(claims.UserID)

	var newPost NewPostRequest
	if err := json.NewDecoder(r.Body).Decode(&newPost); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if newPost.Title == "" || newPost.Content == "" {
		http.Error(w, "Title and content are required", http.StatusBadRequest)
		return
	}

	if newPost.Category == "" {
		http.Error(w, "Category is required", http.StatusBadRequest)
		return
	}

	// Get category ID from category name
	category, err := h.DB.GetCategoryByName(newPost.Category)
	if err != nil {
		log.Printf("Error finding category: %v", err)
		http.Error(w, "Invalid category", http.StatusBadRequest)
		return
	}

	// Create the post in the database with category ID
	err = h.DB.CreatePost(userID, category.ID, newPost.Title, newPost.Content)
	if err != nil {
		log.Printf("Error creating post: %v", err)
		http.Error(w, "Failed to create post", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(PostResponse{Success: true, Message: "Post created successfully"})
}



type LogoutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(LogoutResponse{Success: true, Message: "Logged out successfully"})
}

// Go code for GetPosts handler - add this to your handlers.go file

// Post represents a forum post
// Updated GetPosts handler with category information
type Post struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Category  string    `json:"category"`
	AuthorID  int       `json:"authorId"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"createdAt"`
}

// GetPosts handles retrieving all posts with category information
func (h *Handler) GetPosts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check for category filter in query params
	categoryIDStr := r.URL.Query().Get("category")
	var posts []database.Post
	var err error

	if categoryIDStr != "" {
		// If category filter is provided, get posts for that category
		categoryID, err := strconv.Atoi(categoryIDStr)
		if err != nil {
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		}
		posts, err = h.DB.GetPostsByCategory(categoryID)
	} else {
		// Otherwise get all posts
		posts, err = h.DB.GetAllPosts()
	}

	if err != nil {
		log.Printf("Error retrieving posts: %v", err)
		http.Error(w, "Failed to retrieve posts", http.StatusInternalServerError)
		return
	}

	// Format response
	type PostResponse struct {
		ID        int    `json:"id"`
		Title     string `json:"title"`
		Content   string `json:"content"`
		Category  string `json:"category"`
		Author    string `json:"author"`
		CreatedAt string `json:"createdAt"`
	}

	var responsePosts []PostResponse
	for _, post := range posts {
		responsePosts = append(responsePosts, PostResponse{
			ID:        post.ID,
			Title:     post.Title,
			Content:   post.Content,
			Category:  post.Category,
			Author:    post.Author,
			CreatedAt: post.CreatedAt.Format(time.RFC3339),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(responsePosts)
}

// GetCategories returns all available categories
func (h *Handler) GetCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	categories, err := h.DB.GetAllCategories()
	if err != nil {
		log.Printf("Error retrieving categories: %v", err)
		http.Error(w, "Failed to retrieve categories", http.StatusInternalServerError)
		return
	}

	var response []CategoryResponse
	for _, cat := range categories {
		response = append(response, CategoryResponse{
			ID:          cat.ID,
			Name:        cat.Name,
			Description: cat.Description,
			PostCount:   cat.PostCount,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}


type UserStatus struct {
    ID        int       `json:"id"`
    Nickname  string    `json:"nickname"`
    Online    bool      `json:"online"`
    LastSeen  time.Time `json:"lastSeen"`
}

// GetOnlineUsers returns all users with their online status
func (h *Handler) GetOnlineUsers(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    users, err := h.DB.GetAllUsersWithStatus()
    if err != nil {
        log.Printf("Error retrieving users: %v", err)
        http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
        return
    }

    var response []UserStatus
    for _, user := range users {
        response = append(response, UserStatus{
            ID:       user.User.ID,
            Nickname: user.User.Nickname,
            Online:   user.Online,
            LastSeen: user.LastSeen,
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

// UpdateOnlineStatus updates the current user's online status
func (h *Handler) UpdateOnlineStatus(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Verify JWT token to get user ID
    tokenString := r.Header.Get("Authorization")
    if tokenString == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }
    tokenString = strings.TrimPrefix(tokenString, "Bearer ")

    claims, err := h.VerifyJWTToken(tokenString)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var status struct {
        Online bool `json:"online"`
    }
    if err := json.NewDecoder(r.Body).Decode(&status); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    err = h.DB.UpdateUserStatus(int(claims.UserID), status.Online)
    if err != nil {
        log.Printf("Error updating user status: %v", err)
        http.Error(w, "Failed to update status", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]bool{"success": true})
}