package model

import "time"

// Genre is the GORM model for the genres table.
type Genre struct {
	ID          string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Slug        string `gorm:"type:varchar(50);not null;uniqueIndex"`
	Label       string `gorm:"type:varchar(100);not null"`
	Description string `gorm:"type:text;not null;default:''"`
	SortOrder   int    `gorm:"not null;default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Genre) TableName() string {
	return "genres"
}
