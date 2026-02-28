package database

import (
	"log"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func Initialize() *gorm.DB {
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatal("failed to create data directory: %w", err)
	}

	databasePath := filepath.Join("data", "database.db")
	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		log.Fatal("failed while initializing database: %w", err)
	}

	if err := AutoMigrate(db); err != nil {
		log.Fatal("auto migrate failed: %w", err)
	}

	return db
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Source{},
		&Blacklist{},
		&Whitelist{},
		&RequestLog{},
		&RequestLogIP{},
		&Resolution{},
		&MacAddress{},
		&User{},
		&APIKey{},
		&Notification{},
		&Prefetch{},
		&Audit{},
		&Alert{},
	)
}
