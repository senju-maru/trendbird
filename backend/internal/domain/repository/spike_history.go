package repository

import (
	"context"
	"time"

	"github.com/trendbird/backend/internal/domain/entity"
)

// SpikeHistoryRepository defines the persistence operations for SpikeHistory entities.
type SpikeHistoryRepository interface {
	Create(ctx context.Context, history *entity.SpikeHistory) error
	ListByTopicID(ctx context.Context, topicID string) ([]*entity.SpikeHistory, error)
	CountByUserIDCurrentMonth(ctx context.Context, userID string) (int32, error)
	// ListUnnotified returns spike histories where notified_at IS NULL.
	ListUnnotified(ctx context.Context) ([]*entity.SpikeHistory, error)
	// ListUnnotifiedByStatus returns spike histories where notified_at IS NULL and status matches.
	ListUnnotifiedByStatus(ctx context.Context, status entity.TopicStatus) ([]*entity.SpikeHistory, error)
	// MarkNotified sets notified_at for the given IDs.
	MarkNotified(ctx context.Context, ids []string, at time.Time) error
	// ListByTopicIDsSince returns spike histories for the given topic IDs created after since.
	ListByTopicIDsSince(ctx context.Context, topicIDs []string, since time.Time) ([]*entity.SpikeHistory, error)
}
