package models

import "time"

type QueryParams struct {
	Search       string
	Column       string
	Direction    string
	FilterClient string
	Page         int
	PageSize     int
	Offset       int
}

type DomainHistory struct {
	Domain    string    `json:"domain"`
	Timestamp time.Time `json:"timestamp"`
}

type QueryTypeCount struct {
	QueryType string `json:"queryType"`
	Count     int    `json:"count"`
}
