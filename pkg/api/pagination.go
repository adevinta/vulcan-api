/*
Copyright 2021 Adevinta
*/

package api

// Pagination represents the pagination data requested.
type Pagination struct {
	Page int `json:"page"`
	Size int `json:"size"`
}

// PaginationInfo represents the pagination data provided for each vulnerability DB response.
type PaginationInfo struct {
	Limit  int  `json:"limit"`
	Offset int  `json:"offset"`
	Total  int  `json:"total"`
	More   bool `json:"more"`
}
