package model

import "time"

type RequestLogEntry struct {
	Domain            string        `json:"domain"`
	Status            string        `json:"status"`
	QueryType         string        `json:"queryType"`
	IP                []string      `json:"ip"`
	ResponseSizeBytes int           `json:"responseSizeBytes"`
	Timestamp         time.Time     `json:"timestamp"`
	ResponseTime      time.Duration `json:"responseTimeNS"`
	Blocked           bool          `json:"blocked"`
	Cached            bool          `json:"cached"`
	ClientInfo        *Client       `json:"client"`
}

type RequestLogEntryTimestamps struct {
	Timestamp time.Time `json:"timestamp"`
	Blocked   bool      `json:"blocked"`
}
