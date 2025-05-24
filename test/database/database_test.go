package database

import (
	"goaway/backend/dns/database"
	model "goaway/backend/dns/server/models"
	"os"
	"testing"
	"time"
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

func MockClient() *model.Client {
	return &model.Client{
		IP:   "192.168.13.37",
		Name: "mock-client",
		MAC:  "00:1A:2B:3C:4D:5E",
	}
}

func MockRequestLogEntry() model.RequestLogEntry {
	timestamp := time.Now()
	return model.RequestLogEntry{
		Timestamp:    timestamp,
		Domain:       "example.com",
		IP:           []string{"192.168.0.1"},
		Blocked:      false,
		Cached:       false,
		ResponseTime: 13371337,
		ClientInfo:   MockClient(),
		Status:       "NOERROR",
		QueryType:    "A",
	}
}

func BenchmarkInsertRequestLog(b *testing.B) {
	db, err := createTestDB()
	if err != nil {
		b.Fatalf("Failed to create test database: %v", err)
	}
	defer removeTestDB(db)

	batch := make([]model.RequestLogEntry, 1000)
	for i := range batch {
		batch[i] = MockRequestLogEntry()
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
		batch[i] = MockRequestLogEntry()
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
