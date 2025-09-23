package model

import "time"

type RequestLogEntry struct {
	ID                int64         `json:"id"`
	Domain            string        `json:"domain"`
	Status            string        `json:"status"`
	QueryType         string        `json:"queryType"`
	IP                []ResolvedIP  `json:"ip"`
	ResponseSizeBytes int           `json:"responseSizeBytes"`
	Timestamp         time.Time     `json:"timestamp"`
	ResponseTime      time.Duration `json:"responseTimeNS"`
	Blocked           bool          `json:"blocked"`
	Cached            bool          `json:"cached"`
	ClientInfo        *Client       `json:"client"`
	Protocol          Protocol      `json:"protocol"`
}

type Protocol string

const (
	UDP Protocol = "UDP"
	TCP Protocol = "TCP"
	DoT Protocol = "DoT"
	DoH Protocol = "DoH"
)

type ResolvedIP struct {
	IP    string `json:"ip"`
	RType string `json:"rtype"`
}

type RequestLogIntervalSummary struct {
	IntervalStart string `json:"start"`
	BlockedCount  int    `json:"blocked"`
	CachedCount   int    `json:"cached"`
	AllowedCount  int    `json:"allowed"`
}

type ResponseSizeSummary struct {
	StartUnix            int64     `json:"-"`
	Start                time.Time `json:"start"`
	TotalSizeBytes       int       `json:"total_size_bytes"`
	AvgResponseSizeBytes int       `json:"avg_response_size_bytes"`
	MinResponseSizeBytes int       `json:"min_response_size_bytes"`
	MaxResponseSizeBytes int       `json:"max_response_size_bytes"`
}
