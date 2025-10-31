package models

type QueryParams struct {
	Search       string
	Column       string
	Direction    string
	FilterClient string
	Page         int
	PageSize     int
	Offset       int
}
