package entity

import "time"

// TriggerType represents how a topic research was triggered.
type TriggerType string

const (
	TriggerTypeBatch  TriggerType = "batch"
	TriggerTypeSpike  TriggerType = "spike"
	TriggerTypeRising TriggerType = "rising"
)

// TopicResearch stores a web search result for a topic.
// This entity is immutable (INSERT only).
type TopicResearch struct {
	ID          string
	TopicID     string
	Query       string
	Summary     string
	SourceURLs  []string
	TriggerType TriggerType
	SearchedAt  time.Time
	CreatedAt   time.Time
}
