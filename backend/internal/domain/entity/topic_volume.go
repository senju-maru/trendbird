package entity

import "time"

// TopicVolume represents a single time-series data point for topic mention volume.
// This entity is immutable (INSERT only).
type TopicVolume struct {
	ID        string
	TopicID   string
	Timestamp time.Time
	Value     int32
	CreatedAt time.Time
}
