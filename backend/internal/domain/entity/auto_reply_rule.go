package entity

import "time"

// AutoReplyRule represents a user's automatic reply rule for a specific tweet.
type AutoReplyRule struct {
	ID                 string
	UserID             string
	Enabled            bool
	TargetTweetID      string
	TargetTweetText    string
	TriggerKeywords    []string
	ReplyTemplate      string
	LastCheckedReplyID *string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
