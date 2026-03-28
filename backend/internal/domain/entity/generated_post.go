package entity

import "time"

// PostStyle represents the writing style of an AI-generated post.
type PostStyle int

const (
	PostStyleCasual   PostStyle = 1
	PostStyleBreaking PostStyle = 2
	PostStyleAnalysis PostStyle = 3
)

// GeneratedPost represents an AI-generated post content.
// This entity is immutable (INSERT only).
type GeneratedPost struct {
	ID              string
	UserID          string
	TopicID         *string
	GenerationLogID *string
	Style           PostStyle
	Content         string
	CreatedAt       time.Time
}
