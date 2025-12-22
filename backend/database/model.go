package database

import (
	"time"
)

type Source struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string    `json:"name" validate:"required"`
	URL         string    `gorm:"unique;not null" json:"url" validate:"required,url"`
	Active      bool      `gorm:"default:true" json:"active"`
	LastUpdated time.Time `gorm:"not null" json:"lastUpdated"`
	CreatedAt   time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt   time.Time `gorm:"not null" json:"updatedAt"`
}

type Blacklist struct {
	Domain    string    `gorm:"primaryKey" json:"domain" validate:"required,fqdn"`
	SourceID  uint      `gorm:"primaryKey;not null" json:"sourceID" validate:"required"`
	Source    Source    `gorm:"foreignKey:SourceID;constraint:OnDelete:CASCADE" json:"source"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null" json:"updatedAt"`
}

type Whitelist struct {
	Domain    string    `gorm:"primaryKey" json:"domain" validate:"required,fqdn"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null" json:"updatedAt"`
}

type RequestLog struct {
	ID                uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Timestamp         time.Time      `gorm:"not null;index:idx_timestamp_response_size,priority:1" json:"timestamp"`
	Domain            string         `gorm:"type:varchar(255);not null;index:idx_domain_timestamp,priority:1;index:idx_client_ip_domain,priority:2" json:"domain" validate:"required"`
	ClientIP          string         `gorm:"type:varchar(45);not null;index:idx_client_ip;index:idx_client_ip_domain,priority:1" json:"clientIP" validate:"required,ip"`
	ClientName        string         `gorm:"type:varchar(255)" json:"clientName"`
	QueryType         string         `gorm:"type:varchar(10);index:idx_query_type" json:"queryType"`
	Status            string         `gorm:"type:varchar(20)" json:"status"`
	Protocol          string         `gorm:"type:varchar(10)" json:"protocol"`
	ResponseTimeNs    int64          `gorm:"not null" json:"repsonseTimeNS"`
	ResponseSizeBytes int            `gorm:"index:idx_timestamp_response_size,priority:2" json:"responseSizeBytes"`
	Blocked           bool           `gorm:"not null;index:idx_timestamp_covering,priority:2;default:false" json:"blocked"`
	Cached            bool           `gorm:"not null;index:idx_timestamp_covering,priority:3;default:false" json:"cached"`
	IPs               []RequestLogIP `gorm:"foreignKey:RequestLogID;constraint:OnDelete:CASCADE" json:"ips"`
	CreatedAt         time.Time      `gorm:"not null" json:"createdAt"`
}

type RequestLogIP struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	RequestLogID uint       `gorm:"not null;index" json:"requestLogID" validate:"required"`
	IP           string     `gorm:"type:varchar(45);not null" json:"ip" validate:"required,ip"`
	RecordType   string     `gorm:"type:varchar(10);not null" json:"recordType" validate:"required"`
	RequestLog   RequestLog `gorm:"foreignKey:RequestLogID;constraint:OnDelete:CASCADE" json:"requestLog"`
	CreatedAt    time.Time  `gorm:"not null" json:"createdAt"`
}

type Resolution struct {
	Domain    string    `gorm:"primaryKey" json:"domain" validate:"required,fqdn"`
	IP        string    `gorm:"index" json:"ip" validate:"required,ip"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null" json:"updatedAt"`
}

type MacAddress struct {
	MAC       string    `gorm:"primaryKey" json:"mac" validate:"required,mac"`
	IP        string    `gorm:"index" json:"ip" validate:"required,ip"`
	Vendor    string    `json:"vendor"`
	Bypass    bool      `gorm:"default:false" json:"bypass"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null" json:"updatedAt"`
}

type User struct {
	Username  string    `gorm:"primaryKey" json:"username" validate:"required,min=3,max=50"`
	Password  string    `json:"password" validate:"required,min=8"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null" json:"updatedAt"`
}

type APIKey struct {
	Name      string    `gorm:"primaryKey" json:"name" validate:"required"`
	Key       string    `gorm:"unique;not null" json:"key" validate:"required"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null" json:"updatedAt"`
}

type Notification struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Severity  string    `gorm:"type:varchar(20);not null" json:"severity" validate:"required,oneof=info warning error critical"`
	Category  string    `gorm:"type:varchar(50);not null" json:"category" validate:"required"`
	Text      string    `gorm:"type:text;not null" json:"text" validate:"required"`
	Read      bool      `gorm:"default:false;index" json:"read"`
	CreatedAt time.Time `gorm:"not null;index" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null" json:"updatedAt"`
}

type Prefetch struct {
	Domain    string    `gorm:"primaryKey" json:"domain" validate:"required,fqdn"`
	QueryType int       `gorm:"not null" json:"queryType" validate:"required,min=1"`
	Refresh   int       `gorm:"not null" json:"refresh" validate:"required,min=1"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null" json:"updatedAt"`
}

type Audit struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Topic     string    `gorm:"type:varchar(100);not null;index" json:"topic" validate:"required"`
	Message   string    `gorm:"type:text;not null" json:"message" validate:"required"`
	CreatedAt time.Time `gorm:"not null;index" json:"createdAt"`
}

type Alert struct {
	Type      string    `gorm:"primaryKey" json:"type" validate:"required"`
	Name      string    `gorm:"not null" json:"name" validate:"required"`
	Webhook   string    `json:"webhook" validate:"omitempty,url"`
	Enabled   bool      `gorm:"default:false" json:"enabled"`
	CreatedAt time.Time `gorm:"not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"not null" json:"updatedAt"`
}
