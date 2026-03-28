package entity

import "time"

// UserTopic represents the link between a user and a shared topic.
type UserTopic struct {
	ID                  string
	UserID              string
	TopicID             string
	NotificationEnabled bool
	IsCreator           bool
	CreatedAt           time.Time
}
