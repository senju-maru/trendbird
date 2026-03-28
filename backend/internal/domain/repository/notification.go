package repository

import (
	"context"
	"time"

	"github.com/trendbird/backend/internal/domain/entity"
)

// NotificationRepository defines the persistence operations for Notification entities.
type NotificationRepository interface {
	// Create creates a notification and delivers it to a single user.
	Create(ctx context.Context, notification *entity.Notification) error
	// CreateForUsers creates a notification and delivers it to multiple users.
	CreateForUsers(ctx context.Context, notification *entity.Notification, userIDs []string) error
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*entity.Notification, int64, error)
	MarkAsRead(ctx context.Context, userID, id string) error
	MarkAllAsReadByUserID(ctx context.Context, userID string) error
	CountUnreadByUserID(ctx context.Context, userID string) (int32, error)
	// GetLastNotifiedAt returns the most recent notification timestamp for the user
	// filtered by notification type. Returns zero time if no notification exists.
	GetLastNotifiedAt(ctx context.Context, userID string, notificationType entity.NotificationType) (time.Time, error)
}
