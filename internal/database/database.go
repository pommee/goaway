package database

import (
	"database/sql"
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

	return &Session{Con: db}, nil
}
