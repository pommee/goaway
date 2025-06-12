package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

type DatabaseManager struct {
	Conn  *sql.DB
	Mutex sync.Mutex
}

func Initialize() (*DatabaseManager, error) {
	if err := os.MkdirAll("data", 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory %s: %w", "data", err)
	}

	databasePath := filepath.Join("data", "database.db")
	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		log.Warning("failed to set journal_mode to WAL")
	}

	err = NewBlacklistTable(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create blacklist table: %w", err)
	}

	err = NewWhitelistTable(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create whitelist table: %w", err)
	}

	err = NewSourcesTable(db)
	if err != nil {
		return nil, fmt.Errorf("failed to create sources table: %w", err)
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

	return &DatabaseManager{Conn: db, Mutex: sync.Mutex{}}, nil
}

func NewBlacklistTable(db *sql.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS blacklist (
            domain TEXT,
            source_id INTEGER,
            PRIMARY KEY (domain, source_id),
            FOREIGN KEY (source_id) REFERENCES sources(id)
        )
    `)
	if err != nil {
		return err
	}

	return nil
}

func NewWhitelistTable(db *sql.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS whitelist (
            domain TEXT PRIMARY KEY
        )
    `)
	if err != nil {
		return err
	}

	return nil
}

func NewSourcesTable(db *sql.DB) error {
	_, err := db.Exec(`
        CREATE TABLE IF NOT EXISTS sources (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            name TEXT UNIQUE,
            url TEXT,
			active INTEGER,
            lastUpdated INTEGER
        )
	`)
	if err != nil {
		return err
	}

	return nil
}

func NewRequestLogTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS request_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME NOT NULL,
			domain TEXT NOT NULL,
			blocked BOOLEAN NOT NULL,
			cached BOOLEAN NOT NULL,
			response_time_ns INTEGER NOT NULL,
			client_ip TEXT,
			client_name TEXT,
			status TEXT,
			query_type TEXT,
			response_size_bytes INTEGER
		);

		CREATE TABLE IF NOT EXISTS request_log_ips (
			id SERIAL PRIMARY KEY,
			request_log_id INTEGER REFERENCES request_log(id) ON DELETE CASCADE,
			ip TEXT NOT NULL,
			rtype TEXT NOT NULL
		);
    `)
	if err != nil {
		return err
	}

	return nil
}

func NewResolutionTable(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS resolution (
		domain TEXT NOT NULL PRIMARY KEY,
		ip TEXT
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
