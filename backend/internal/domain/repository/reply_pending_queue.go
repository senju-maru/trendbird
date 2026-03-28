package repository

import (
	"context"

	"github.com/trendbird/backend/internal/domain/entity"
)

// ReplyPendingQueueRepository defines the persistence operations for ReplyPendingQueue entities.
type ReplyPendingQueueRepository interface {
	// Create inserts a new pending reply. Idempotent: ON CONFLICT DO NOTHING.
	Create(ctx context.Context, item *entity.ReplyPendingQueue) error
	// ListPendingGroupedByUser returns all pending items grouped by user_id.
	ListPendingGroupedByUser(ctx context.Context) (map[string][]*entity.ReplyPendingQueue, error)
	// UpdateStatus updates the status of a pending reply item.
	UpdateStatus(ctx context.Context, id string, status entity.ReplyPendingStatus) error
	// DeleteByID removes a pending reply item.
	DeleteByID(ctx context.Context, id string) error
}
