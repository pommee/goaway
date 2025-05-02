package models

type QueryParams struct {
	Page, PageSize    int
	Search            string
	Column, Direction string
	Offset            int
}
