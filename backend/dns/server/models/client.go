package model

import "time"

// Client represents a DNS client with associated metadata.
// It includes the client's IP address, hostname, MAC address, and an ignored flag.'
// The 'bypass' field indicates whether the client should be allowed to bypass blacklist rules.
type Client struct {
	IP       string    `json:"ip"`
	LastSeen time.Time `json:"lastSeen"`
	Name     string    `json:"name"`
	Mac      string    `json:"mac"`
	Vendor   string    `json:"vendor"`
	Bypass   bool      `json:"bypass"`
}
