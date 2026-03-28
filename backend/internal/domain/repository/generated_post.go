package repository

import (
	"context"

	"github.com/trendbird/backend/internal/domain/entity"
)

// GeneratedPostRepository defines the persistence operations for GeneratedPost entities.
type GeneratedPostRepository interface {
	BulkCreate(ctx context.Context, posts []*entity.GeneratedPost) error
	ListByUserIDAndTopicID(ctx context.Context, userID, topicID string) ([]*entity.GeneratedPost, error)
}
