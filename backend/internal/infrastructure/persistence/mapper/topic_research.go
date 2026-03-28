package mapper

import (
	"github.com/lib/pq"
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// TopicResearchToEntity converts a GORM TopicResearch model to a domain entity.
func TopicResearchToEntity(m *model.TopicResearch) *entity.TopicResearch {
	return &entity.TopicResearch{
		ID:          m.ID,
		TopicID:     m.TopicID,
		Query:       m.Query,
		Summary:     m.Summary,
		SourceURLs:  []string(m.SourceURLs),
		TriggerType: entity.TriggerType(m.TriggerType),
		SearchedAt:  m.SearchedAt,
		CreatedAt:   m.CreatedAt,
	}
}

// TopicResearchToModel converts a domain TopicResearch entity to a GORM model.
func TopicResearchToModel(e *entity.TopicResearch) *model.TopicResearch {
	return &model.TopicResearch{
		ID:          e.ID,
		TopicID:     e.TopicID,
		Query:       e.Query,
		Summary:     e.Summary,
		SourceURLs:  pq.StringArray(e.SourceURLs),
		TriggerType: string(e.TriggerType),
		SearchedAt:  e.SearchedAt,
		CreatedAt:   e.CreatedAt,
	}
}
