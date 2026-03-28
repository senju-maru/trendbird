package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// NotificationToEntity converts a GORM Notification model to a domain entity.
// UserID と IsRead は含まれない（user_notifications との JOIN 結果から別途設定する）。
func NotificationToEntity(m *model.Notification) *entity.Notification {
	var topicStatus *entity.TopicStatus
	if m.TopicStatus != nil {
		ts := entity.TopicStatus(*m.TopicStatus)
		topicStatus = &ts
	}
	return &entity.Notification{
		ID:          m.ID,
		Type:        entity.NotificationType(m.Type),
		Title:       m.Title,
		Message:     m.Message,
		TopicID:     m.TopicID,
		TopicName:   m.TopicName,
		TopicStatus: topicStatus,
		ActionURL:   m.ActionURL,
		ActionLabel: m.ActionLabel,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// NotificationToModel converts a domain Notification entity to a GORM model.
// UserID と IsRead は notifications テーブルに含まれないため変換しない。
func NotificationToModel(e *entity.Notification) *model.Notification {
	var topicStatus *int32
	if e.TopicStatus != nil {
		ts := int32(*e.TopicStatus)
		topicStatus = &ts
	}
	return &model.Notification{
		ID:          e.ID,
		Type:        int32(e.Type),
		Title:       e.Title,
		Message:     e.Message,
		TopicID:     e.TopicID,
		TopicName:   e.TopicName,
		TopicStatus: topicStatus,
		ActionURL:   e.ActionURL,
		ActionLabel: e.ActionLabel,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}

// notificationJoinRow は notifications + user_notifications の JOIN 結果を表す。
type NotificationJoinRow struct {
	model.Notification
	UserID string `gorm:"column:user_id"`
	IsRead bool   `gorm:"column:is_read"`
}

// NotificationJoinToEntity converts a JOIN result row to a domain entity with UserID and IsRead.
func NotificationJoinToEntity(row *NotificationJoinRow) *entity.Notification {
	e := NotificationToEntity(&row.Notification)
	e.UserID = row.UserID
	e.IsRead = row.IsRead
	return e
}
