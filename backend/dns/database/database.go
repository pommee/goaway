package database

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type Session struct {
	Con *sql.DB
}

func Initialize() (*Session, error) {
	db, err := sql.Open("sqlite", "database.db")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		log.Warning("failed to set journal_mode to WAL")
	}

	err = NewRequestLogTable(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create request_log table: %w", err)
	}

	err = NewResolutionTable(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create resolution table: %w", err)
	}

	err = NewMacTable(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create mac_addresses table: %w", err)
	}

	err = NewUserTable(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create user table: %w", err)
	}

	err = NewAPIKeyTable(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create api key table: %w", err)
	}

	err = NewNotificationsTable(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create notifications table: %w", err)
	}

	err = NewPrefetchTable(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create prefetch table: %w", err)
	}

	return &Session{Con: db}, nil
}

func NewRequestLogTable(db *sql.DB) error {
	_, err := db.Exec(`
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
			query_type TEXT,
			response_size_bytes TEXT
        );
    `)
	if err != nil {
		return err
	}

	return nil
}

func NewResolutionTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS resolution (
		ip TEXT PRIMARY KEY,
		domain TEXT NOT NULL
	)`)
	if err != nil {
		return err
	}

	return nil
}

func NewMacTable(db *sql.DB) error {
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

func NewUserTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS user (
		username TEXT PRIMARY KEY,
		password TEXT
	)`)
	if err != nil {
		return err
	}

	return nil
}

func NewAPIKeyTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS apikey (
		name TEXT PRIMARY KEY,
		key TEXT,
		created_at DATETIME NOT NULL
	)`)
	if err != nil {
		return err
	}

	return nil
}

func NewNotificationsTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS notifications (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		severity TEXT,
		category TEXT,
		text TEXT,
		read BOOLEAN,
		created_at DATETIME NOT NULL
	)`)
	if err != nil {
		return err
	}

	return nil
}

func NewPrefetchTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS prefetch (
		domain TEXT PRIMARY KEY,
		refresh INTEGER,
		qtype INTEGER
	)`)
	if err != nil {
		return err
	}

	return nil
}
