package entity

import "time"

// Genre represents a genre master record.
type Genre struct {
	ID          string
	Slug        string
	Label       string
	Description string
	SortOrder   int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
