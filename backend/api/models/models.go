package models

type QueryParams struct {
	Page      int
	PageSize  int
	Offset    int
	Search    string
	Column    string
	Direction string
}
