package database

import (
	"fmt"
	"goaway/backend/settings"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DatabaseManager struct {
	Conn  *gorm.DB
	Mutex *sync.RWMutex
}

type Source struct {
	ID          uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string `json:"name"`
	URL         string `gorm:"unique" json:"url"`
	Active      bool   `json:"active"`
	LastUpdated int64  `json:"lastUpdated"`
}

type Blacklist struct {
	Domain   string `gorm:"primaryKey" json:"domain"`
	SourceID uint   `gorm:"primaryKey" json:"source_id"`
	Source   Source `gorm:"foreignKey:SourceID;references:ID" json:"source"`
}

type Whitelist struct {
	Domain string `gorm:"primaryKey" json:"domain"`
}

type RequestLog struct {
	ID                uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Timestamp         time.Time      `gorm:"not null;index:idx_request_log_timestamp_covering,priority:1;index:idx_request_log_timestamp_desc;index:idx_request_log_domain_timestamp,priority:2" json:"timestamp"`
	Domain            string         `gorm:"type:varchar(255);not null;index:idx_request_log_domain_timestamp,priority:1;index:idx_client_ip_domain,priority:2" json:"domain"`
	Blocked           bool           `gorm:"not null;index:idx_request_log_timestamp_covering,priority:2" json:"blocked"`
	Cached            bool           `gorm:"not null;index:idx_request_log_timestamp_covering,priority:3" json:"cached"`
	ResponseTimeNs    int64          `gorm:"not null" json:"response_time_ns"`
	ClientIP          string         `gorm:"type:varchar(45);index:idx_client_ip;index:idx_client_ip_domain,priority:1" json:"client_ip"`
	ClientName        string         `gorm:"type:varchar(255)" json:"client_name"`
	Status            string         `gorm:"type:varchar(50)" json:"status"`
	QueryType         string         `gorm:"type:varchar(10)" json:"query_type"`
	ResponseSizeBytes int            `json:"response_size_bytes"`
	Protocol          string         `gorm:"type:varchar(10)" json:"protocol"`
	IPs               []RequestLogIP `gorm:"foreignKey:RequestLogID;constraint:OnDelete:CASCADE" json:"ips"`
}

type RequestLogIP struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	RequestLogID uint       `gorm:"not null;index" json:"request_log_id"`
	IP           string     `gorm:"type:varchar(45);not null" json:"ip"`
	RType        string     `gorm:"type:varchar(10);not null" json:"rtype"`
	RequestLog   RequestLog `gorm:"foreignKey:RequestLogID;references:ID" json:"request_log"`
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
	Name      string    `gorm:"primaryKey" json:"name"`
	Key       string    `json:"key"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

type Notification struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Severity  string    `json:"severity"`
	Category  string    `json:"category"`
	Text      string    `json:"text"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

type Prefetch struct {
	Domain  string `gorm:"primaryKey" json:"domain"`
	Refresh int    `json:"refresh"`
	QType   int    `json:"qtype"`
}

type Audit struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Topic     string    `json:"topic"`
	Message   string    `json:"message"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

type Alert struct {
	Type    string `gorm:"primaryKey" json:"type"`
	Enabled bool   `json:"enabled"`
	Name    string `json:"name"`
	Webhook string `json:"webhook"`
}

func Initialize(config *settings.Config) *DatabaseManager {
	var db *gorm.DB
	var err error
	if config.DB.DbType == "postgres" {
		sslMode := "disable"
		if *config.DB.SSL {
			sslMode = "enable"
		}
		dbString := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s", *config.DB.Host, *config.DB.User, *config.DB.Pass, *config.DB.Database, *config.DB.Port, sslMode, *config.DB.TimeZone)
		db, err = gorm.Open(postgres.Open(dbString), &gorm.Config{})
	} else if config.DB.DbType == "sqlite" {
		if err := os.MkdirAll("data", 0755); err != nil {
			log.Fatal("failed to create data directory: %v", err)
		}

		databasePath := filepath.Join("data", "database.db")
		db, err = gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	} else {
		log.Fatal("invalid DB_TYPE")
	}
	if err != nil {
		log.Fatal("failed while initializing database: %v", err)
	}
	if db == nil {
		log.Fatal("failed to initialize database")
		os.Exit(1)
	}

	if err := AutoMigrate(db); err != nil {
		log.Fatal("auto migrate failed: %v", err)
	}

	return &DatabaseManager{
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
