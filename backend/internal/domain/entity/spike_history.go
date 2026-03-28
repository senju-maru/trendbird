package entity

import "time"

// SpikeHistory records a past spike event for a topic.
// This entity is immutable (INSERT only), except for NotifiedAt which is updated by batch jobs.
type SpikeHistory struct {
	ID              string
	TopicID         string
	TopicName       string // JOIN で取得（バッチ用）
	Timestamp       time.Time
	PeakZScore      float64
	Status          TopicStatus
	Summary         string
	DurationMinutes int32
	NotifiedAt      *time.Time
	CreatedAt       time.Time
}
