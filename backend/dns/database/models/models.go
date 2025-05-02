package models

import "time"

type Client struct {
	Name, Mac, Vendor string
	LastSeen          time.Time
}

type ClientDetails struct {
	IP, Name, MAC string
}

type ClientRequestDetails struct {
	TotalRequests, UniqueDomains, BlockedRequests, CachedRequests int
	AvgResponseTimeMs                                             float64
	LastSeen, MostQueriedDomain                                   string
}

type Resolution struct {
	IP     string `json:"ip"`
	Domain string `json:"domain"`
}
