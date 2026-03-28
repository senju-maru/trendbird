package repository

import (
	"context"
	"time"

	"github.com/trendbird/backend/internal/domain/entity"
)

// TopicVolumeRepository defines the persistence operations for TopicVolume entities.
type TopicVolumeRepository interface {
	BulkCreate(ctx context.Context, volumes []*entity.TopicVolume) error
	ListByTopicIDAndRange(ctx context.Context, topicID string, from, to time.Time) ([]*entity.TopicVolume, error)
}
