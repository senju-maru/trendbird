package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// UserTopicToEntity converts a GORM UserTopic model to a domain entity.
func UserTopicToEntity(m *model.UserTopic) *entity.UserTopic {
	return &entity.UserTopic{
		ID:                  m.ID,
		UserID:              m.UserID,
		TopicID:             m.TopicID,
		NotificationEnabled: m.NotificationEnabled,
		IsCreator:           m.IsCreator,
		CreatedAt:           m.CreatedAt,
	}
}

// UserTopicToModel converts a domain UserTopic entity to a GORM model.
func UserTopicToModel(e *entity.UserTopic) *model.UserTopic {
	return &model.UserTopic{
		ID:                  e.ID,
		UserID:              e.UserID,
		TopicID:             e.TopicID,
		NotificationEnabled: e.NotificationEnabled,
		IsCreator:           e.IsCreator,
		CreatedAt:           e.CreatedAt,
	}
}
