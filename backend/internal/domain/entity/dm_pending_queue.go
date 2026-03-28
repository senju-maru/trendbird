package entity

import "time"

// DMPendingStatus represents the status of a pending DM.
type DMPendingStatus int

const (
	DMPendingStatusPending DMPendingStatus = 1
	DMPendingStatusSent    DMPendingStatus = 2
	DMPendingStatusFailed  DMPendingStatus = 3
)

// DMPendingQueue represents a DM waiting to be sent (for rate limit overflow).
type DMPendingQueue struct {
	ID                 string
	UserID             string
	RuleID             string
	RecipientTwitterID string
	ReplyTweetID       string
	TriggerKeyword     string
	Status             DMPendingStatus
	CreatedAt          time.Time
}
