package model

import (
	"time"

	"github.com/lib/pq"
)

// TopicResearch is the GORM model for the topic_research table.
// This table is immutable (INSERT only) so it has no UpdatedAt field.
type TopicResearch struct {
	ID          string         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TopicID     string         `gorm:"type:uuid;not null;index:idx_topic_research_topic_id"`
	Query       string         `gorm:"type:text;not null"`
	Summary     string         `gorm:"type:text;not null"`
	SourceURLs  pq.StringArray `gorm:"type:text[];default:'{}'"`
	TriggerType string         `gorm:"type:text;not null"`
	SearchedAt  time.Time      `gorm:"not null;default:now()"`
	CreatedAt   time.Time

	Topic Topic `gorm:"foreignKey:TopicID;constraint:OnDelete:CASCADE"`
}

func (TopicResearch) TableName() string {
	return "topic_research"
}
