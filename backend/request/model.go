package request

import "time"

type Client struct {
	LastSeen time.Time
	Name     string
	Mac      string
	Vendor   string
}

type ClientNameAndIP struct {
	Name string
	IP   string
}
