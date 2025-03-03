package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"

	"fmt"
)

type Session struct {
	Con *sql.DB
}

func Initialize() (*Session, error) {
	db, err := sql.Open("sqlite3", "database.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS request_log (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            timestamp DATETIME NOT NULL,
            domain TEXT NOT NULL,
			ip TEXT NOT NULL,
            blocked BOOLEAN NOT NULL,
            cached BOOLEAN NOT NULL,
            response_time_ns INTEGER NOT NULL,
            client_ip TEXT,
            client_name TEXT,
			status TEXT,
			query_type TEXT
        );
    `)
	if err != nil {
		return nil, fmt.Errorf("failed to create request_log table: %w", err)
	}

	err = NewMacDatabase(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create mac_addresses table: %w", err)
	}

	return &Session{Con: db}, nil
}

func NewMacDatabase(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS mac_addresses (
		mac TEXT PRIMARY KEY,
		ip TEXT,
		vendor TEXT
	)`)
	if err != nil {
		return err
	}

	return nil
}
