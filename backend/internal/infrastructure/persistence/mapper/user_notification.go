package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// UserNotificationToEntity converts a GORM UserNotification model to a domain entity.
func UserNotificationToEntity(m *model.UserNotification) *entity.UserNotification {
	return &entity.UserNotification{
		ID:             m.ID,
		UserID:         m.UserID,
		NotificationID: m.NotificationID,
		IsRead:         m.IsRead,
		CreatedAt:      m.CreatedAt,
	}
}

// UserNotificationToModel converts a domain UserNotification entity to a GORM model.
func UserNotificationToModel(e *entity.UserNotification) *model.UserNotification {
	return &model.UserNotification{
		ID:             e.ID,
		UserID:         e.UserID,
		NotificationID: e.NotificationID,
		IsRead:         e.IsRead,
		CreatedAt:      e.CreatedAt,
	}
}
