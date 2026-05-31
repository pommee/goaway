package model

import (
	"fmt"
	"time"
)

type RequestLogEntry struct {
	Timestamp         time.Time     `json:"timestamp"`
	ClientInfo        *Client       `json:"client"`
	Domain            string        `json:"domain"`
	Status            string        `json:"status"`
	QueryType         string        `json:"queryType"`
	Protocol          Protocol      `json:"protocol"`
	IP                []ResolvedIP  `json:"ip"`
	ID                uint          `json:"id"`
	ResponseSizeBytes int           `json:"responseSizeBytes"`
	ResponseTime      time.Duration `json:"responseTimeNS"`
	Blocked           bool          `json:"blocked"`
	Cached            bool          `json:"cached"`
}

func (r *RequestLogEntry) String() string {
	return fmt.Sprintf(
		"Time: %d, Client: %v, Domain: %s, Status: %s, Type: %s, Protocol: %s, IPs: %+v, ID: %d, ResponseSize: %d, ResponseTime: %dns, Blocked: %t, Cached: %t",
		r.Timestamp.Unix(),
		r.ClientInfo,
		r.Domain,
		r.Status,
		r.QueryType,
		r.Protocol,
		r.IP,
		r.ID,
		r.ResponseSizeBytes,
		r.ResponseTime,
		r.Blocked,
		r.Cached,
	)
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
	Start                time.Time `json:"start"`
	StartUnix            int64     `json:"-"`
	TotalSizeBytes       int       `json:"total_size_bytes"`
	AvgResponseSizeBytes int       `json:"avg_response_size_bytes"`
	MinResponseSizeBytes int       `json:"min_response_size_bytes"`
	MaxResponseSizeBytes int       `json:"max_response_size_bytes"`
}
