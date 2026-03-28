package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// GeneratedPostToEntity converts a GORM GeneratedPost model to a domain entity.
func GeneratedPostToEntity(m *model.GeneratedPost) *entity.GeneratedPost {
	return &entity.GeneratedPost{
		ID:              m.ID,
		UserID:          m.UserID,
		TopicID:         m.TopicID,
		GenerationLogID: m.GenerationLogID,
		Style:           entity.PostStyle(m.Style),
		Content:         m.Content,
		CreatedAt:       m.CreatedAt,
	}
}

// GeneratedPostToModel converts a domain GeneratedPost entity to a GORM model.
func GeneratedPostToModel(e *entity.GeneratedPost) *model.GeneratedPost {
	return &model.GeneratedPost{
		ID:              e.ID,
		UserID:          e.UserID,
		TopicID:         e.TopicID,
		GenerationLogID: e.GenerationLogID,
		Style:           int32(e.Style),
		Content:         e.Content,
		CreatedAt:       e.CreatedAt,
	}
}
