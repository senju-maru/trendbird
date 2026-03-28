package entity

import "time"

// PostingTip holds AI-calculated posting time recommendations for a topic.
type PostingTip struct {
	ID                string
	TopicID           string
	PeakDays          []string
	PeakHoursStart    int32
	PeakHoursEnd      int32
	NextSuggestedTime time.Time
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
