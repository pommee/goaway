package database

import (
	"goaway/internal/database"
	model "goaway/internal/server/models"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func createTestDB() (*database.Session, error) {
	_ = os.Remove("database.db")

	db, err := database.Initialize()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func removeTestDB(db *database.Session) {
	_ = db.Con.Close()
	_ = os.Remove("database.db")
}

func BenchmarkInsertRequestLog(b *testing.B) {
	db, err := createTestDB()
	if err != nil {
		b.Fatalf("Failed to create test database: %v", err)
	}
	defer removeTestDB(db)

	batch := make([]model.RequestLogEntry, 1000)
	for i := range batch {
		batch[i] = model.RequestLogEntry{
			Timestamp:    time.Now(),
			Domain:       "example.com",
			IP:           []string{"192.168.0.1"},
			Blocked:      false,
			Cached:       true,
			ResponseTime: 1000000,
			ClientInfo: &model.Client{
				IP:   "192.168.1.2",
				Name: "client1",
				MAC:  "00:1A:2B:3C:4D:5E",
			},
			Status:    "NOERROR",
			QueryType: "A",
		}
	}

	for iteration := 0; b.Loop(); iteration++ {
		database.SaveRequestLog(db.Con, batch)
	}
}

func BenchmarkQueryRequestLog(b *testing.B) {
	db, err := createTestDB()
	if err != nil {
		b.Fatalf("Failed to create test database: %v", err)
	}
	defer removeTestDB(db)

	batch := make([]model.RequestLogEntry, 100000)
	for i := range batch {
		batch[i] = model.RequestLogEntry{
			Timestamp:    time.Now(),
			Domain:       "example.com",
			IP:           []string{"192.168.0.1"},
			Blocked:      false,
			Cached:       true,
			ResponseTime: 1000000,
			ClientInfo: &model.Client{
				IP:   "192.168.1.2",
				Name: "client1",
				MAC:  "00:1A:2B:3C:4D:5E",
			},
			Status:    "OK",
			QueryType: "A",
		}
	}
	database.SaveRequestLog(db.Con, batch)

	for b.Loop() {
		rows, err := db.Con.Query("SELECT * FROM request_log WHERE domain = ?", "example.com")
		if err != nil {
			b.Fatalf("Failed to query request log: %v", err)
		}
		_ = rows.Close()
	}
}
