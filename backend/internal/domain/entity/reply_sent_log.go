package entity

import "time"

// ReplySentLog represents a record of a reply that was sent automatically.
type ReplySentLog struct {
	ID               string
	UserID           string
	RuleID           string
	OriginalTweetID  string
	OriginalAuthorID string
	ReplyTweetID     string
	TriggerKeyword   string
	ReplyText        string
	SentAt           time.Time
}
