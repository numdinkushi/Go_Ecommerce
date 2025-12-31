package dto

import "time"

type ProductQuery struct {
	PaginationParams
	Search    string     `json:"search" query:"search"`
	Beginning *time.Time `json:"beginning" query:"beginning"` // ISO 8601 date format: 2024-01-01T00:00:00Z
	Ending    *time.Time `json:"ending" query:"ending"`       // ISO 8601 date format: 2024-02-01T00:00:00Z
}
