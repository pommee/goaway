package database

import (
	"fmt"
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
	dsn := fmt.Sprintf("file:%s?cache=shared&mode=rwc&_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL&_query_only=false", databasePath)

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		log.Fatal("failed while initializing database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("failed to get database connection: %w", err)
	}
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(0)

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
		&MacAddress{},
		&User{},
		&APIKey{},
		&Notification{},
		&Prefetch{},
		&Audit{},
		&Alert{},
	)
}
