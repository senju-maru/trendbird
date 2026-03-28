package repository

import (
	"context"

	"github.com/trendbird/backend/internal/domain/entity"
)

// PostingTipRepository defines the persistence operations for PostingTip entities.
type PostingTipRepository interface {
	FindByTopicID(ctx context.Context, topicID string) (*entity.PostingTip, error)
	Upsert(ctx context.Context, tip *entity.PostingTip) error
}
