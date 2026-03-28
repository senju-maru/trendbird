package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// SpikeHistoryToEntity converts a GORM SpikeHistory model to a domain entity.
func SpikeHistoryToEntity(m *model.SpikeHistory) *entity.SpikeHistory {
	return &entity.SpikeHistory{
		ID:              m.ID,
		TopicID:         m.TopicID,
		Timestamp:       m.Timestamp,
		PeakZScore:      m.PeakZScore,
		Status:          entity.TopicStatus(m.Status),
		Summary:         m.Summary,
		DurationMinutes: m.DurationMinutes,
		NotifiedAt: m.NotifiedAt,
		CreatedAt:       m.CreatedAt,
	}
}

// SpikeHistoryToModel converts a domain SpikeHistory entity to a GORM model.
func SpikeHistoryToModel(e *entity.SpikeHistory) *model.SpikeHistory {
	return &model.SpikeHistory{
		ID:              e.ID,
		TopicID:         e.TopicID,
		Timestamp:       e.Timestamp,
		PeakZScore:      e.PeakZScore,
		Status:          int32(e.Status),
		Summary:         e.Summary,
		DurationMinutes: e.DurationMinutes,
		NotifiedAt: e.NotifiedAt,
		CreatedAt:       e.CreatedAt,
	}
}
