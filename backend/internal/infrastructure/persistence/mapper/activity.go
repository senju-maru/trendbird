package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// ActivityToEntity converts a GORM Activity model to a domain entity.
func ActivityToEntity(m *model.Activity) *entity.Activity {
	return &entity.Activity{
		ID:          m.ID,
		UserID:      m.UserID,
		Type:        entity.ActivityType(m.Type),
		TopicName:   m.TopicName,
		Description: m.Description,
		Timestamp:   m.Timestamp,
		CreatedAt:   m.CreatedAt,
	}
}

// ActivityToModel converts a domain Activity entity to a GORM model.
func ActivityToModel(e *entity.Activity) *model.Activity {
	return &model.Activity{
		ID:          e.ID,
		UserID:      e.UserID,
		Type:        int32(e.Type),
		TopicName:   e.TopicName,
		Description: e.Description,
		Timestamp:   e.Timestamp,
		CreatedAt:   e.CreatedAt,
	}
}
