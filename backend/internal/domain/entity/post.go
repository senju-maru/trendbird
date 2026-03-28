package entity

import "time"

// PostStatus represents the lifecycle state of a post.
type PostStatus int

const (
	PostDraft     PostStatus = 1
	PostScheduled PostStatus = 2
	PostPublished PostStatus = 3
	PostFailed    PostStatus = 4
)

// Post represents a user's post (draft, scheduled, published, or failed).
type Post struct {
	ID           string
	UserID       string
	Content      string
	TopicID      *string
	TopicName    *string
	Status       PostStatus
	ScheduledAt  *time.Time
	PublishedAt  *time.Time
	FailedAt     *time.Time
	ErrorMessage *string
	TweetURL     *string
	Likes        int32
	Retweets     int32
	Replies      int32
	Views        int32
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
