package database

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Manager struct {
	Conn  *gorm.DB
	Mutex *sync.RWMutex
}

type Source struct {
	Name        string `json:"name"`
	URL         string `gorm:"unique" json:"url"`
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	LastUpdated int64  `json:"lastUpdated"`
	Active      bool   `json:"active"`
}

type Blacklist struct {
	Domain   string `gorm:"primaryKey" json:"domain"`
	Source   Source `gorm:"foreignKey:SourceID;references:ID" json:"source"`
	SourceID uint   `gorm:"primaryKey" json:"source_id"`
}

type Whitelist struct {
	Domain string `gorm:"primaryKey" json:"domain"`
}

type RequestLog struct {
	Timestamp         time.Time      `gorm:"not null;index:idx_request_log_timestamp_covering,priority:1;index:idx_request_log_timestamp_desc;index:idx_request_log_domain_timestamp,priority:2" json:"timestamp"`
	QueryType         string         `gorm:"type:varchar(10)" json:"query_type"`
	Domain            string         `gorm:"type:varchar(255);not null;index:idx_request_log_domain_timestamp,priority:1;index:idx_client_ip_domain,priority:2" json:"domain"`
	ClientIP          string         `gorm:"type:varchar(45);index:idx_client_ip;index:idx_client_ip_domain,priority:1" json:"client_ip"`
	ClientName        string         `gorm:"type:varchar(255)" json:"client_name"`
	Status            string         `gorm:"type:varchar(50)" json:"status"`
	Protocol          string         `gorm:"type:varchar(10)" json:"protocol"`
	IPs               []RequestLogIP `gorm:"foreignKey:RequestLogID;constraint:OnDelete:CASCADE" json:"ips"`
	ResponseTimeNs    int64          `gorm:"not null" json:"response_time_ns"`
	ID                uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	ResponseSizeBytes int            `json:"response_size_bytes"`
	Blocked           bool           `gorm:"not null;index:idx_request_log_timestamp_covering,priority:2" json:"blocked"`
	Cached            bool           `gorm:"not null;index:idx_request_log_timestamp_covering,priority:3" json:"cached"`
}

type RequestLogIP struct {
	IP           string     `gorm:"type:varchar(45);not null" json:"ip"`
	RType        string     `gorm:"type:varchar(10);not null" json:"rtype"`
	RequestLog   RequestLog `gorm:"foreignKey:RequestLogID;references:ID" json:"request_log"`
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	RequestLogID uint       `gorm:"not null;index" json:"request_log_id"`
}

type Resolution struct {
	Domain string `gorm:"primaryKey" json:"domain"`
	IP     string `json:"ip"`
}

type MacAddress struct {
	MAC    string `gorm:"primaryKey" json:"mac"`
	IP     string `json:"ip"`
	Vendor string `json:"vendor"`
}

type User struct {
	Username string `gorm:"primaryKey" json:"username"`
	Password string `json:"password"`
}

type APIKey struct {
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	Name      string    `gorm:"primaryKey" json:"name"`
	Key       string    `json:"key"`
}

type Notification struct {
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	Severity  string    `json:"severity"`
	Category  string    `json:"category"`
	Text      string    `json:"text"`
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Read      bool      `json:"read"`
}

type Prefetch struct {
	Domain  string `gorm:"primaryKey" json:"domain"`
	Refresh int    `json:"refresh"`
	QType   int    `json:"qtype"`
}

type Audit struct {
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
	Topic     string    `json:"topic"`
	Message   string    `json:"message"`
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
}

type Alert struct {
	Type    string `gorm:"primaryKey" json:"type"`
	Name    string `json:"name"`
	Webhook string `json:"webhook"`
	Enabled bool   `json:"enabled"`
}

func Initialize() *Manager {
	if err := os.MkdirAll("data", 0755); err != nil {
		log.Fatal("failed to create data directory: %v", err)
	}

	databasePath := filepath.Join("data", "database.db")
	db, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	if err != nil {
		log.Fatal("failed while initializing database: %v", err)
	}

	if err := AutoMigrate(db); err != nil {
		log.Fatal("auto migrate failed: %v", err)
	}

	return &Manager{
		Conn:  db,
		Mutex: &sync.RWMutex{},
	}
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
