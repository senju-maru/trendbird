package entity

import "time"

// UserNotification represents the relationship between a user and a notification,
// including the per-user read status.
type UserNotification struct {
	ID             string
	UserID         string
	NotificationID string
	IsRead         bool
	CreatedAt      time.Time
}
