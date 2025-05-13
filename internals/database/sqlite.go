package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB


// Initialize the database connection and schema
func InitDB() error {
	var err error
	DB, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// Test the connection
	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Enable foreign keys
	if _, err := DB.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("error enabling foreign keys: %v", err)
	}

	// Create tables if they don't exist
	if err := createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %v", err)
	}

	log.Println("Database initialized successfully!")
	return nil
}

// Create tables using schema.sql
func createTables() error {
	schema, err := os.ReadFile("internals/database/schema.sql") // Read the schema.sql file
	if err != nil {
		return fmt.Errorf("failed to read schema file: %v", err)
	}

	_, err = DB.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("failed to execute schema: %v", err)
	}
	return nil
}
