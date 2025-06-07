package database

import (
	"goaway/backend/dns/database"
	model "goaway/backend/dns/server/models"
	"os"
	"testing"
	"time"
)

func createTestDB() (*database.DatabaseManager, error) {
	_ = os.Remove("database.db")

	db, err := database.Initialize()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func removeTestDB(db *database.DatabaseManager) {
	_ = db.Conn.Close()
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
	dbManager, err := createTestDB()
	if err != nil {
		b.Fatalf("Failed to create test database: %v", err)
	}
	defer removeTestDB(dbManager)

	batch := make([]model.RequestLogEntry, 1000)
	for i := range batch {
		batch[i] = MockRequestLogEntry()
	}

	for iteration := 0; b.Loop(); iteration++ {
		database.SaveRequestLog(dbManager.Conn, batch)
	}
}

func BenchmarkQueryRequestLog(b *testing.B) {
	dbManager, err := createTestDB()
	if err != nil {
		b.Fatalf("Failed to create test database: %v", err)
	}
	defer removeTestDB(dbManager)

	batch := make([]model.RequestLogEntry, 100000)
	for i := range batch {
		batch[i] = MockRequestLogEntry()
	}
	database.SaveRequestLog(dbManager.Conn, batch)

	for b.Loop() {
		rows, err := dbManager.Conn.Query("SELECT * FROM request_log WHERE domain = ?", "example.com")
		if err != nil {
			b.Fatalf("Failed to query request log: %v", err)
		}
		_ = rows.Close()
	}
}

func TestFetchResolution(t *testing.T) {
	dbManager, err := createTestDB()
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer removeTestDB(dbManager)

	_, err = dbManager.Conn.Exec("INSERT INTO resolution (ip, domain) VALUES (?, ?)", "192.168.1.1", "example.com")
	if err != nil {
		t.Fatalf("Failed to insert exact domain: %v", err)
	}

	_, err = dbManager.Conn.Exec("INSERT INTO resolution (ip, domain) VALUES (?, ?)", "127.0.0.1", "*.google.com")
	if err != nil {
		t.Fatalf("Failed to insert wildcard domain: %v", err)
	}

	tests := []struct {
		name     string
		domain   string
		expected string
	}{
		{
			name:     "Exact match",
			domain:   "example.com",
			expected: "192.168.1.1",
		},
		{
			name:     "Wildcard match",
			domain:   "somethingrandom.google.com",
			expected: "127.0.0.1",
		},
		{
			name:     "No match",
			domain:   "nonexistent.com",
			expected: "",
		},
		{
			name:     "Trailing dot removed",
			domain:   "example.com.",
			expected: "192.168.1.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := database.FetchResolution(dbManager.Conn, tt.domain)
			if err != nil {
				t.Fatalf("FetchResolution failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
