package entity

import "time"

// AutoDMRule represents a user's automatic DM reply rule.
type AutoDMRule struct {
	ID                  string
	UserID              string
	Enabled             bool
	TriggerKeywords     []string
	TemplateMessage     string
	LastCheckedReplyID  *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
