package entity

import "time"

// NotificationType represents the category of a notification.
type NotificationType int

const (
	NotificationTrend  NotificationType = 1
	NotificationSystem NotificationType = 2
)

// Notification represents a user notification.
// UserID と IsRead は user_notifications テーブルとの JOIN から取得される。
// notifications テーブル自体にはこれらのカラムは存在しない。
type Notification struct {
	ID          string
	UserID      string // user_notifications から JOIN で取得
	Type        NotificationType
	Title       string
	Message     string
	IsRead      bool // user_notifications から JOIN で取得
	TopicID     *string
	TopicName   *string
	TopicStatus *TopicStatus
	ActionURL   *string
	ActionLabel *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
