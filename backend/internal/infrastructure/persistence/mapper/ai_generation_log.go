package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// AIGenerationLogToEntity converts a GORM AIGenerationLog model to a domain entity.
func AIGenerationLogToEntity(m *model.AIGenerationLog) *entity.AIGenerationLog {
	return &entity.AIGenerationLog{
		ID:        m.ID,
		UserID:    m.UserID,
		TopicID:   m.TopicID,
		Style:     entity.PostStyle(m.Style),
		Count:     m.Count,
		CreatedAt: m.CreatedAt,
	}
}

// AIGenerationLogToModel converts a domain AIGenerationLog entity to a GORM model.
func AIGenerationLogToModel(e *entity.AIGenerationLog) *model.AIGenerationLog {
	return &model.AIGenerationLog{
		ID:        e.ID,
		UserID:    e.UserID,
		TopicID:   e.TopicID,
		Style:     int32(e.Style),
		Count:     e.Count,
		CreatedAt: e.CreatedAt,
	}
}
