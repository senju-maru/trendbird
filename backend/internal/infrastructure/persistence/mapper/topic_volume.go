package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// TopicVolumeToEntity converts a GORM TopicVolume model to a domain entity.
func TopicVolumeToEntity(m *model.TopicVolume) *entity.TopicVolume {
	return &entity.TopicVolume{
		ID:        m.ID,
		TopicID:   m.TopicID,
		Timestamp: m.Timestamp,
		Value:     m.Value,
		CreatedAt: m.CreatedAt,
	}
}

// TopicVolumeToModel converts a domain TopicVolume entity to a GORM model.
func TopicVolumeToModel(e *entity.TopicVolume) *model.TopicVolume {
	return &model.TopicVolume{
		ID:        e.ID,
		TopicID:   e.TopicID,
		Timestamp: e.Timestamp,
		Value:     e.Value,
		CreatedAt: e.CreatedAt,
	}
}
