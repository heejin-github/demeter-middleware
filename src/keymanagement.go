package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

var (
	dbPath string
	db     *sql.DB
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err!= nil {
		log.Fatalf("Error getting home directory: %v", err)
	}
	dbPath = filepath.Join(homeDir, ".demeter", "apikeys.db")

	err = os.MkdirAll(filepath.Dir(dbPath), 0755)
	if err != nil {
		log.Fatalf("error creating directory: %v", err)
	}

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
		key TEXT NOT NULL UNIQUE,
		username TEXT NOT NULL,
		protocol TEXT NOT NULL
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

func storeAPIKey(apiKey, username, protocol string) error {
	_, err := db.Exec("INSERT INTO api_keys (key, username, protocol) VALUES (?, ?, ?)", apiKey, username, protocol)
	if err != nil {
		log.Printf("Error storing API key: %v", err)
	} else {
		log.Printf("Stored new API key: %s for user: %s with protocol: %s", apiKey, username, protocol)
	}
	return err
}

func listAPIKeys() ([]struct {
	Key      string
	Username string
	Protocol string
}, error) {
	rows, err := db.Query("SELECT key, username, protocol FROM api_keys")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []struct {
		Key      string
		Username string
		Protocol string
	}
	for rows.Next() {
		var key, username, protocol string
		if err := rows.Scan(&key, &username, &protocol); err != nil {
			return nil, err
		}
		keys = append(keys, struct {
			Key      string
			Username string
			Protocol string
		}{key, username, protocol})
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

func getAPIKeyDetails(apiKey string) (string, string, error) {
	var username, protocol string
	err := db.QueryRow("SELECT username, protocol FROM api_keys WHERE key = ?", apiKey).Scan(&username, &protocol)
	if err != nil {
		log.Printf("Error getting API key details: %v", err)
		return "", "", err
	}
	return username, protocol, nil
}
