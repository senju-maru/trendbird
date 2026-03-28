package model

import "time"

// Topic is the GORM model for the topics table.
type Topic struct {
	ID             string   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name           string   `gorm:"type:varchar(100);not null;uniqueIndex:uq_topics_name_genre_id,priority:1"`
	Keywords       string   `gorm:"type:jsonb;not null;default:'[]'"`
	GenreID        string   `gorm:"type:uuid;not null;uniqueIndex:uq_topics_name_genre_id,priority:2"`
	Status         int32    `gorm:"not null;default:3"`
	ChangePercent  float64  `gorm:"not null;default:0"`
	ZScore         *float64 `gorm:""`
	CurrentVolume  int32    `gorm:"not null;default:0"`
	BaselineVolume int32    `gorm:"not null;default:0"`
	Context        *string  `gorm:"type:text"`
	ContextSummary *string  `gorm:"type:text"`
	SpikeStartedAt *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time

	Genre Genre `gorm:"foreignKey:GenreID;constraint:OnDelete:RESTRICT"`
}

func (Topic) TableName() string {
	return "topics"
}
