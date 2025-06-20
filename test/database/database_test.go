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
	_ = os.RemoveAll("data")
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
		Timestamp: timestamp,
		Domain:    "example.com",
		IP: []model.ResolvedIP{
			{
				IP:    "192.168.0.1",
				RType: "A",
			},
		},
		Blocked:      false,
		Cached:       false,
		ResponseTime: 1337,
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

	testData := []struct {
		ip     string
		domain string
	}{
		{"192.168.1.1", "example.com"},
		{"127.0.0.1", "*.google.com"},
		{"10.0.0.1", "*.sub1.example.com"},
		{"172.16.0.1", "*.example.com"},
	}

	for _, data := range testData {
		err = database.CreateNewResolution(dbManager.Conn, data.ip, data.domain)
		if err != nil {
			t.Fatalf("Failed to insert domain %s: %v", data.domain, err)
		}
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
			name:     "Single level wildcard",
			domain:   "somethingrandom.google.com",
			expected: "127.0.0.1",
		},
		{
			name:     "Multi-level subdomain",
			domain:   "sub2.sub1.example.com",
			expected: "10.0.0.1",
		},
		{
			name:     "Multi-level subdomain",
			domain:   "sub3.sub2.example.com",
			expected: "172.16.0.1",
		},
		{
			name:     "Deep nesting",
			domain:   "a.b.c.sub1.example.com",
			expected: "10.0.0.1",
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
