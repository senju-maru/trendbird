package entity

import "time"

// AIGenerationLog records a single AI post generation event.
// This table is immutable (INSERT only).
type AIGenerationLog struct {
	ID        string
	UserID    string
	TopicID   *string
	Style     PostStyle
	Count     int32
	CreatedAt time.Time
}
