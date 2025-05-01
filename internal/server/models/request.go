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

type RequestLogIntervalSummary struct {
	IntervalStart time.Time `json:"start"`
	BlockedCount  int       `json:"blocked"`
	CachedCount   int       `json:"cached"`
	AllowedCount  int       `json:"allowed"`
}
