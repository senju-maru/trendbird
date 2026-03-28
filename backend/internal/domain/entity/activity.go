package entity

import "time"

// ActivityType represents the kind of activity event.
type ActivityType int

const (
	ActivitySpike        ActivityType = 1
	ActivityRising       ActivityType = 2
	ActivityAIGenerated  ActivityType = 3
	ActivityPosted       ActivityType = 4
	ActivityTopicAdded   ActivityType = 5
	ActivityTopicRemoved ActivityType = 6
	ActivityLogin        ActivityType = 7
)

// Activity represents a user activity feed entry.
// This entity is immutable (INSERT only).
type Activity struct {
	ID          string
	UserID      string
	Type        ActivityType
	TopicName   string
	Description string
	Timestamp   time.Time
	CreatedAt   time.Time
}
