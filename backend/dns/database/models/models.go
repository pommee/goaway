package models

import "time"

type Client struct {
	LastSeen time.Time
	Name     string
	Mac      string
	Vendor   string
}

type ClientDetails struct {
	IP, Name, MAC string
}

type ClientRequestDetails struct {
	LastSeen          string
	MostQueriedDomain string
	TotalRequests     int
	UniqueDomains     int
	BlockedRequests   int
	CachedRequests    int
	AvgResponseTimeMs float64
}

type Resolution struct {
	IP     string `json:"ip"`
	Domain string `json:"domain"`
}
