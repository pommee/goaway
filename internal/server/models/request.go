package model

import "time"

type RequestLogEntry struct {
	Domain         string        `json:"domain"`
	Status         string        `json:"status"`
	QueryType      string        `json:"queryType"`
	IP             []string      `json:"ip"`
	Timestamp      time.Time     `json:"timestamp"`
	ResponseTimeNS time.Duration `json:"responseTimeNS"`
	Blocked        bool          `json:"blocked"`
	Cached         bool          `json:"cached"`
	ClientInfo     *Client       `json:"client"`
}
