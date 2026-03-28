package entity

import "time"

// NotificationSetting holds per-user notification preferences.
type NotificationSetting struct {
	ID            string
	UserID        string
	SpikeEnabled  bool
	RisingEnabled bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
