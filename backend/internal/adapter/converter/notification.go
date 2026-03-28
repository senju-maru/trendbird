package converter

import (
	"time"

	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/internal/domain/entity"
)

// NotificationToProto は entity.Notification を Proto Notification に変換する。
func NotificationToProto(e *entity.Notification) *trendbirdv1.Notification {
	n := &trendbirdv1.Notification{
		Id:        e.ID,
		Type:      trendbirdv1.NotificationType(e.Type),
		Title:     e.Title,
		Message:   e.Message,
		Timestamp: e.CreatedAt.Format(time.RFC3339),
		IsRead:    e.IsRead,
		TopicId:   e.TopicID,
		TopicName: e.TopicName,
		ActionUrl: e.ActionURL,
		ActionLabel: e.ActionLabel,
	}

	if e.TopicStatus != nil {
		ts := trendbirdv1.TopicStatus(*e.TopicStatus)
		n.TopicStatus = &ts
	}

	return n
}

// NotificationSliceToProto は Notification スライスを Proto スライスに変換する。
func NotificationSliceToProto(es []*entity.Notification) []*trendbirdv1.Notification {
	result := make([]*trendbirdv1.Notification, len(es))
	for i, e := range es {
		result[i] = NotificationToProto(e)
	}
	return result
}
