package database

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

//go:embed schema.sql
var schemaFS embed.FS

type Database struct {
	DB *sql.DB
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

func (d *Database) AuthenticateUser(nickname, email, password string) error {
	var hashedPassword string

	// Fetch password hash from database
	err := d.DB.QueryRow("SELECT password FROM users WHERE nickname = ? or email = ? ", nickname, email).Scan(&hashedPassword)
	if err != nil {
		return err // User not found
	}

	// Compare stored hashed password with the provided password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return err // Incorrect password
	}

	return nil
}
