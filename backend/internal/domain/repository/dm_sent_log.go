package repository

import (
	"context"

	"github.com/trendbird/backend/internal/domain/entity"
)

// DMSentLogRepository defines the persistence operations for DMSentLog entities.
type DMSentLogRepository interface {
	// Create inserts a new DM sent log record.
	Create(ctx context.Context, log *entity.DMSentLog) error
	// ExistsByReplyTweetID checks if a DM has already been sent for the given reply tweet.
	ExistsByReplyTweetID(ctx context.Context, replyTweetID string, recipientTwitterID string) (bool, error)
	// ListByUserID returns DM sent logs for the given user, ordered by sent_at DESC.
	ListByUserID(ctx context.Context, userID string, limit int) ([]*entity.DMSentLog, error)
}
