package model

import "time"

type RequestLogEntry struct {
	Domain       string        `json:"domain,omitempty"`
	Status       string        `json:"status,omitempty"`
	QueryType    string        `json:"queryType,omitempty"`
	IP           []string      `json:"ip,omitempty"`
	Timestamp    time.Time     `json:"timestamp,omitempty"`
	ResponseTime time.Duration `json:"responseTimeNS,omitempty"`
	Blocked      bool          `json:"blocked"`
	Cached       bool          `json:"cached,omitempty"`
	ClientInfo   *Client       `json:"client,omitempty"`
}

type RequestLogEntryTimestamps struct {
	Timestamp time.Time `json:"timestamp"`
	Blocked   bool      `json:"blocked"`
}
