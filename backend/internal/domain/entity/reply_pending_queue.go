package entity

import "time"

// ReplyPendingStatus represents the status of a pending reply.
type ReplyPendingStatus int

const (
	ReplyPendingStatusPending ReplyPendingStatus = 1
	ReplyPendingStatusSent    ReplyPendingStatus = 2
	ReplyPendingStatusFailed  ReplyPendingStatus = 3
)

// ReplyPendingQueue represents a reply waiting to be sent (for rate limit overflow).
type ReplyPendingQueue struct {
	ID               string
	UserID           string
	RuleID           string
	OriginalTweetID  string
	OriginalAuthorID string
	TriggerKeyword   string
	Status           ReplyPendingStatus
	CreatedAt        time.Time
}
