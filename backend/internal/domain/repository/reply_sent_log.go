package repository

import (
	"context"

	"github.com/trendbird/backend/internal/domain/entity"
)

// ReplySentLogRepository defines the persistence operations for ReplySentLog entities.
type ReplySentLogRepository interface {
	// Create inserts a new reply sent log record.
	Create(ctx context.Context, log *entity.ReplySentLog) error
	// ExistsByOriginalTweetID checks if a reply has already been sent for the given original tweet.
	ExistsByOriginalTweetID(ctx context.Context, originalTweetID string, originalAuthorID string) (bool, error)
	// ListByUserID returns reply sent logs for the given user, ordered by sent_at DESC.
	ListByUserID(ctx context.Context, userID string, limit int) ([]*entity.ReplySentLog, error)
}
