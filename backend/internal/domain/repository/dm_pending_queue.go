package repository

import (
	"context"

	"github.com/trendbird/backend/internal/domain/entity"
)

// DMPendingQueueRepository defines the persistence operations for DMPendingQueue entities.
type DMPendingQueueRepository interface {
	// Create inserts a new pending DM. Idempotent: ON CONFLICT DO NOTHING.
	Create(ctx context.Context, item *entity.DMPendingQueue) error
	// ListPendingByUserID returns pending items for the given user, ordered by created_at ASC.
	ListPendingByUserID(ctx context.Context, userID string, limit int) ([]*entity.DMPendingQueue, error)
	// ListPendingGroupedByUser returns all pending items grouped by user_id.
	ListPendingGroupedByUser(ctx context.Context) (map[string][]*entity.DMPendingQueue, error)
	// UpdateStatus updates the status of a pending DM item.
	UpdateStatus(ctx context.Context, id string, status entity.DMPendingStatus) error
	// DeleteByID removes a pending DM item.
	DeleteByID(ctx context.Context, id string) error
}
