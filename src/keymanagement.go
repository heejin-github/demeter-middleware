package main

import (
	"database/sql"
	"log"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

const dbPath = "./apikeys.db"

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}

	createTable()
}

func createTable() {
	query := `
	CREATE TABLE IF NOT EXISTS api_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT NOT NULL UNIQUE
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

func generateAPIKey() string {
	return uuid.New().String()
}

func storeAPIKey(apiKey string) error {
	_, err := db.Exec("INSERT INTO api_keys (key) VALUES (?)", apiKey)
	if err != nil {
		log.Printf("Error storing API key: %v", err)
	} else {
		log.Printf("Stored new API key: %s", apiKey)
	}
	return err
}

func listAPIKeys() ([]string, error) {
	rows, err := db.Query("SELECT key FROM api_keys")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}

	log.Printf("Retrieved %d API keys", len(keys))
	return keys, nil
}

func deleteAPIKey(apiKey string) error {
	_, err := db.Exec("DELETE FROM api_keys WHERE key = ?", apiKey)
	if err != nil {
		log.Printf("Error deleting API key: %v", err)
	} else {
		log.Printf("Deleted API key: %s", apiKey)
	}
	return err
}

func isValidAPIKey(apiKey string) bool {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM api_keys WHERE key = ?", apiKey).Scan(&count)

	if err != nil {
		log.Printf("Error checking API key: %v", err)
		return false
	}
	return count > 0
}
