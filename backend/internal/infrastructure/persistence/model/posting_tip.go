package model

import "time"

// PostingTip is the GORM model for the posting_tips table.
type PostingTip struct {
	ID                string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TopicID           string    `gorm:"type:uuid;uniqueIndex;not null"`
	PeakDays          string    `gorm:"type:jsonb;not null;default:'[]'"`
	PeakHoursStart    int32     `gorm:"not null;default:0"`
	PeakHoursEnd      int32     `gorm:"not null;default:23"`
	NextSuggestedTime time.Time `gorm:"not null"`
	CreatedAt         time.Time
	UpdatedAt         time.Time

	Topic Topic `gorm:"foreignKey:TopicID;constraint:OnDelete:CASCADE"`
}

func (PostingTip) TableName() string {
	return "posting_tips"
}
