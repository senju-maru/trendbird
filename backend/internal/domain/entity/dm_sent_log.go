package entity

import "time"

// DMSentLog represents a record of a DM that was sent automatically.
type DMSentLog struct {
	ID                 string
	UserID             string
	RuleID             string
	RecipientTwitterID string
	ReplyTweetID       string
	TriggerKeyword     string
	DMText             string
	SentAt             time.Time
}
