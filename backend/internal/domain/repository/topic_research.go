package repository

import (
	"context"
	"time"

	"github.com/trendbird/backend/internal/domain/entity"
)

// TopicResearchRepository defines the persistence operations for TopicResearch entities.
type TopicResearchRepository interface {
	Create(ctx context.Context, research *entity.TopicResearch) error
	ListByTopicID(ctx context.Context, topicID string, limit int) ([]*entity.TopicResearch, error)
	ListLatestByTopicID(ctx context.Context, topicID string, since time.Time) ([]*entity.TopicResearch, error)
}
