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