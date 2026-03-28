package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// PostingTipToEntity converts a GORM PostingTip model to a domain entity.
func PostingTipToEntity(m *model.PostingTip) *entity.PostingTip {
	return &entity.PostingTip{
		ID:                m.ID,
		TopicID:           m.TopicID,
		PeakDays:          JSONToStringSlice(m.PeakDays),
		PeakHoursStart:    m.PeakHoursStart,
		PeakHoursEnd:      m.PeakHoursEnd,
		NextSuggestedTime: m.NextSuggestedTime,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
}

// PostingTipToModel converts a domain PostingTip entity to a GORM model.
func PostingTipToModel(e *entity.PostingTip) *model.PostingTip {
	return &model.PostingTip{
		ID:                e.ID,
		TopicID:           e.TopicID,
		PeakDays:          StringSliceToJSON(e.PeakDays),
		PeakHoursStart:    e.PeakHoursStart,
		PeakHoursEnd:      e.PeakHoursEnd,
		NextSuggestedTime: e.NextSuggestedTime,
		CreatedAt:         e.CreatedAt,
		UpdatedAt:         e.UpdatedAt,
	}
}
