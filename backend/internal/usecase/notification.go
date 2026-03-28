package usecase

import (
	"context"

	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/domain/repository"
)

// NotificationUsecase handles notification operations.
type NotificationUsecase struct {
	notiRepo repository.NotificationRepository
}

// NewNotificationUsecase creates a new NotificationUsecase.
func NewNotificationUsecase(notiRepo repository.NotificationRepository) *NotificationUsecase {
	return &NotificationUsecase{
		notiRepo: notiRepo,
	}
}

// ListNotifications returns a paginated list of notifications for the user.
func (u *NotificationUsecase) ListNotifications(ctx context.Context, userID string, limit, offset int) ([]*entity.Notification, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return u.notiRepo.ListByUserID(ctx, userID, limit, offset)
}

// MarkAsRead marks a single notification as read.
func (u *NotificationUsecase) MarkAsRead(ctx context.Context, userID string, notificationID string) error {
	return u.notiRepo.MarkAsRead(ctx, userID, notificationID)
}

// MarkAllAsRead marks all notifications for the user as read.
func (u *NotificationUsecase) MarkAllAsRead(ctx context.Context, userID string) error {
	return u.notiRepo.MarkAllAsReadByUserID(ctx, userID)
}
