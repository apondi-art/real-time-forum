package database

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

//go:embed schema.sql
var schemaFS embed.FS

type Database struct {
	DB *sql.DB
}
type User struct {
	ID       int
	Nickname string
	Email    string
}

type Post struct {
	ID        int
	UserID    int
	CategoryID int    // Added CategoryID field
	Category   string // Added Category name field for convenience
	Title     string
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
	Author    string // Added field to store the author's nickname
}

// Category represents a forum category
type Category struct {
	ID          int
	Name        string
	Description string
	PostCount   int       // Count of posts in this category
	CreatedAt   time.Time
}



// Initialize the database connection and schema
func New(dbPath string) (*Database, error) {
	if dbPath == "" {
		return nil, errors.New("database path cannot be empty")
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxIdleConns(25)
	db.SetMaxOpenConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("error enaling foreign keys: %w", err)
	}

	// Initialize schema
	schema, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to read schema: %w", err)
	}

	if _, err := db.Exec(string(schema)); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	log.Println("Database initialized successfully!")
	return &Database{DB: db}, nil
}

func (d *Database) Close() error {
	return d.DB.Close()
}

// RegisterUser adds a new user to the database
func (d *Database) RegisterUser(nickname, email, password, lname, fname, gender string, age int) error {
	// Check if user already exists
	var count int
	err := d.DB.QueryRow("SELECT COUNT(*) FROM users WHERE nickname = ? OR email = ?",
		nickname, email).Scan(&count)
	if err != nil {
		return fmt.Errorf("error checking existing user: %w", err)
	}

	if count > 0 {
		return errors.New("username or email already exists")
	}

	// Insert the new user
	_, err = d.DB.Exec(
		"INSERT INTO users (nickname, email,age,gender,first_name,last_name,password, created_at) VALUES (?, ?, ?, ?,?,?,?,?)",
		nickname, email, age, gender, fname, lname, password, time.Now())
	if err != nil {
		return fmt.Errorf("error creating user: %w", err)
	}

	return nil
}

func PasswordHashing(pasword string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pasword), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("Error occured during password hashing: %v\n", err)
		return "", err
	}
	return string(bytes), nil
}

// Fixed AuthenticateUser function with additional safeguards and debugging

func (d *Database) AuthenticateUser(nickname, email, password string) (*User, error) {
	// Debugging info
	fmt.Printf("AuthenticateUser called with: nickname='%s', email='%s', password_length=%d\n",
		nickname, email, len(password))

	// Check if we have at least some credentials
	if nickname == "" && email == "" {
		return nil, fmt.Errorf("no credentials provided")
	}

	// Check if password is provided
	if password == "" {
		return nil, fmt.Errorf("password required")
	}

	// Define a struct to hold user data from database
	var user struct {
		ID       int
		Nickname string
		Email    string
		Password string
	}

	// Build the query based on provided credentials
	query := `SELECT id, nickname, email, password FROM users WHERE `
	var conditions []string
	var queryParams []interface{}

	if nickname != "" {
		conditions = append(conditions, "nickname = ?")
		queryParams = append(queryParams, nickname)
	}

	if email != "" {
		conditions = append(conditions, "email = ?")
		queryParams = append(queryParams, email)
	}

	query += strings.Join(conditions, " OR ")

	fmt.Printf("Executing SQL query: %s with params: %v\n", query, queryParams)

	// Fetch user data from database
	err := d.DB.QueryRow(query, queryParams...).Scan(&user.ID, &user.Nickname, &user.Email, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("DB query result: No user found")
			return nil, fmt.Errorf("user not found")
		}
		fmt.Printf("DB query error: %v\n", err)
		return nil, fmt.Errorf("database error: %w", err)
	}

	fmt.Printf("User found in DB: ID=%d, Nickname=%s, Email=%s\n", user.ID, user.Nickname, user.Email)

	// Compare stored hashed password with provided password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		fmt.Println("Password comparison failed")
		return nil, fmt.Errorf("invalid password")
	}

	fmt.Println("Password validation successful")

	// Return user data without the password for security
	return &User{
		ID:       user.ID,
		Nickname: user.Nickname,
		Email:    user.Email,
	}, nil
}

// CreatePost creates a new post with a category
func (d *Database) CreatePost(userID int, categoryID int, title, content string) error {
	_, err := d.DB.Exec(
		"INSERT INTO posts (user_id, category_id, title, content, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		userID, categoryID, title, content, time.Now(), time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to create post: %w", err)
	}
	return nil
}
// GetAllPosts retrieves all posts with category information
func (db *Database) GetAllPosts() ([]Post, error) {
	rows, err := db.DB.Query(`
		SELECT p.id, p.title, p.content, p.user_id, p.category_id, c.name, u.nickname, p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		JOIN categories c ON p.category_id = c.id
		ORDER BY p.created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query all posts: %w", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var createdAtStr string
		var updatedAtStr string
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.UserID,
			&post.CategoryID,
			&post.Category, // Scan the category name
			&post.Author,   // Scan the nickname into the Author field
			&createdAtStr,
			&updatedAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post row: %w", err)
		}

		// Parse the timestamp strings into time.Time
		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			log.Printf("failed to parse created_at timestamp: %v", err)
			createdAt = time.Now() // Fallback to current time on parse error
		}
		post.CreatedAt = createdAt

		updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			log.Printf("failed to parse updated_at timestamp: %v", err)
			updatedAt = time.Now() // Fallback to current time on parse error
		}
		post.UpdatedAt = updatedAt

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return posts, nil
}

// GetAllCategories retrieves all categories with post counts
func (db *Database) GetAllCategories() ([]Category, error) {
	rows, err := db.DB.Query(`
		SELECT c.id, c.name, c.description, COUNT(p.id) as post_count, c.created_at
		FROM categories c
		LEFT JOIN posts p ON c.id = p.category_id
		GROUP BY c.id
		ORDER BY c.name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []Category
	for rows.Next() {
		var category Category
		var createdAtStr string
		var description sql.NullString
		
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&description,
			&category.PostCount,
			&createdAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category row: %w", err)
		}

		if description.Valid {
			category.Description = description.String
		}

		// Parse the timestamp string into time.Time
		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			log.Printf("failed to parse category created_at timestamp: %v", err)
			createdAt = time.Now() // Fallback to current time on parse error
		}
		category.CreatedAt = createdAt

		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during categories iteration: %w", err)
	}

	return categories, nil
}

// GetCategoryByName retrieves a category by its name
func (db *Database) GetCategoryByName(name string) (*Category, error) {
	var category Category
	var createdAtStr string
	var description sql.NullString

	err := db.DB.QueryRow(`
		SELECT id, name, description, created_at
		FROM categories
		WHERE name = ?
	`, name).Scan(
		&category.ID,
		&category.Name,
		&description,
		&createdAtStr,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found: %s", name)
		}
		return nil, fmt.Errorf("failed to query category: %w", err)
	}

	if description.Valid {
		category.Description = description.String
	}

	// Parse the timestamp string into time.Time
	createdAt, err := time.Parse(time.RFC3339, createdAtStr)
	if err != nil {
		log.Printf("failed to parse category created_at timestamp: %v", err)
		createdAt = time.Now() // Fallback to current time on parse error
	}
	category.CreatedAt = createdAt

	return &category, nil
}

// GetPostsByCategory retrieves all posts for a specific category
func (db *Database) GetPostsByCategory(categoryID int) ([]Post, error) {
	rows, err := db.DB.Query(`
		SELECT p.id, p.title, p.content, p.user_id, p.category_id, c.name, u.nickname, p.created_at, p.updated_at
		FROM posts p
		JOIN users u ON p.user_id = u.id
		JOIN categories c ON p.category_id = c.id
		WHERE p.category_id = ?
		ORDER BY p.created_at DESC
	`, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query posts by category: %w", err)
	}
	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var createdAtStr string
		var updatedAtStr string
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.UserID,
			&post.CategoryID,
			&post.Category,
			&post.Author,
			&createdAtStr,
			&updatedAtStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post row: %w", err)
		}

		// Parse the timestamp strings into time.Time
		createdAt, err := time.Parse(time.RFC3339, createdAtStr)
		if err != nil {
			log.Printf("failed to parse created_at timestamp: %v", err)
			createdAt = time.Now() // Fallback to current time on parse error
		}
		post.CreatedAt = createdAt

		updatedAt, err := time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			log.Printf("failed to parse updated_at timestamp: %v", err)
			updatedAt = time.Now() // Fallback to current time on parse error
		}
		post.UpdatedAt = updatedAt

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return posts, nil
}
