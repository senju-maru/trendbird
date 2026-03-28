package repository

import (
	"context"
	"time"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	domainrepo "github.com/trendbird/backend/internal/domain/repository"
	"github.com/trendbird/backend/internal/infrastructure/persistence"
	"github.com/trendbird/backend/internal/infrastructure/persistence/mapper"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
	"gorm.io/gorm"
)

var _ domainrepo.NotificationRepository = (*notificationRepository)(nil)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *notificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

// Create creates a notification and delivers it to a single user.
func (r *notificationRepository) Create(ctx context.Context, notification *entity.Notification) error {
	m := mapper.NotificationToModel(notification)
	if err := r.getDB(ctx).Create(m).Error; err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to create notification", err)
	}
	notification.ID = m.ID
	notification.CreatedAt = m.CreatedAt
	notification.UpdatedAt = m.UpdatedAt

	// user_notifications にレコードを作成
	un := &model.UserNotification{
		UserID:         notification.UserID,
		NotificationID: m.ID,
		IsRead:         false,
	}
	if err := r.getDB(ctx).Create(un).Error; err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to create user_notification", err)
	}
	return nil
}

// CreateForUsers creates a notification and delivers it to multiple users.
func (r *notificationRepository) CreateForUsers(ctx context.Context, notification *entity.Notification, userIDs []string) error {
	m := mapper.NotificationToModel(notification)
	if err := r.getDB(ctx).Create(m).Error; err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to create notification", err)
	}
	notification.ID = m.ID
	notification.CreatedAt = m.CreatedAt
	notification.UpdatedAt = m.UpdatedAt

	// user_notifications にバッチ INSERT
	userNotifications := make([]model.UserNotification, len(userIDs))
	for i, uid := range userIDs {
		userNotifications[i] = model.UserNotification{
			UserID:         uid,
			NotificationID: m.ID,
			IsRead:         false,
		}
	}
	if len(userNotifications) > 0 {
		if err := r.getDB(ctx).CreateInBatches(userNotifications, 100).Error; err != nil {
			return apperror.Wrap(apperror.CodeInternal, "failed to create user_notifications", err)
		}
	}
	return nil
}

// ListByUserID returns notifications for a user via JOIN with user_notifications.
func (r *notificationRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*entity.Notification, int64, error) {
	db := r.getDB(ctx)

	var total int64
	if err := db.Table("user_notifications").
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, apperror.Wrap(apperror.CodeInternal, "failed to count notifications", err)
	}

	var rows []mapper.NotificationJoinRow
	err := db.Table("notifications").
		Select("notifications.*, user_notifications.user_id, user_notifications.is_read").
		Joins("JOIN user_notifications ON notifications.id = user_notifications.notification_id").
		Where("user_notifications.user_id = ?", userID).
		Order("notifications.created_at DESC").
		Limit(limit).Offset(offset).
		Find(&rows).Error
	if err != nil {
		return nil, 0, apperror.Wrap(apperror.CodeInternal, "failed to list notifications", err)
	}

	entities := make([]*entity.Notification, len(rows))
	for i := range rows {
		entities[i] = mapper.NotificationJoinToEntity(&rows[i])
	}
	return entities, total, nil
}

// MarkAsRead marks a single notification as read for the user.
func (r *notificationRepository) MarkAsRead(ctx context.Context, userID, id string) error {
	result := r.getDB(ctx).Model(&model.UserNotification{}).
		Where("user_id = ? AND notification_id = ?", userID, id).
		Update("is_read", true)
	if result.Error != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to mark notification as read", result.Error)
	}
	if result.RowsAffected == 0 {
		return apperror.NotFound("notification not found")
	}
	return nil
}

// MarkAllAsReadByUserID marks all notifications as read for the user.
func (r *notificationRepository) MarkAllAsReadByUserID(ctx context.Context, userID string) error {
	err := r.getDB(ctx).Model(&model.UserNotification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
	if err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to mark all notifications as read", err)
	}
	return nil
}

// CountUnreadByUserID counts unread notifications for the user.
func (r *notificationRepository) CountUnreadByUserID(ctx context.Context, userID string) (int32, error) {
	var count int64
	err := r.getDB(ctx).
		Model(&model.UserNotification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	if err != nil {
		return 0, apperror.Wrap(apperror.CodeInternal, "failed to count unread notifications", err)
	}
	return int32(count), nil
}

// GetLastNotifiedAt returns the most recent notification timestamp for the user
// filtered by notification type. Returns zero time if no notification exists.
func (r *notificationRepository) GetLastNotifiedAt(ctx context.Context, userID string, notificationType entity.NotificationType) (time.Time, error) {
	var result struct {
		MaxCreatedAt *time.Time
	}
	err := r.getDB(ctx).
		Table("notifications").
		Select("MAX(notifications.created_at) as max_created_at").
		Joins("JOIN user_notifications ON notifications.id = user_notifications.notification_id").
		Where("user_notifications.user_id = ? AND notifications.type = ?", userID, notificationType).
		Scan(&result).Error
	if err != nil {
		return time.Time{}, apperror.Wrap(apperror.CodeInternal, "failed to get last notified at", err)
	}
	if result.MaxCreatedAt == nil {
		return time.Time{}, nil
	}
	return *result.MaxCreatedAt, nil
}
