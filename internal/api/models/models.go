package models

type QueryParams struct {
	Page, PageSize    int
	Search            string
	Column, Direction string
	Offset            int
}

type NewResolution struct {
	IP     string
	Domain string
}
