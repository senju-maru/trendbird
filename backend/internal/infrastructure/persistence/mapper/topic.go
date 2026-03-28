package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// TopicToEntity converts a GORM Topic model to a domain entity.
// GenreSlug は Preload/JOIN で Genre リレーションが読み込まれている場合に設定される。
func TopicToEntity(m *model.Topic) *entity.Topic {
	e := &entity.Topic{
		ID:             m.ID,
		Name:           m.Name,
		Keywords:       JSONToStringSlice(m.Keywords),
		GenreID:        m.GenreID,
		Status:         entity.TopicStatus(m.Status),
		ChangePercent:  m.ChangePercent,
		ZScore:         m.ZScore,
		CurrentVolume:  m.CurrentVolume,
		BaselineVolume: m.BaselineVolume,
		Context:        m.Context,
		ContextSummary: m.ContextSummary,
		SpikeStartedAt: m.SpikeStartedAt,
		CreatedAt:      m.CreatedAt,
		UpdatedAt:      m.UpdatedAt,
	}
	// Genre リレーションが Preload/JOIN されている場合
	if m.Genre.Slug != "" {
		e.GenreSlug = m.Genre.Slug
	}
	return e
}

// TopicToModel converts a domain Topic entity to a GORM model.
func TopicToModel(e *entity.Topic) *model.Topic {
	return &model.Topic{
		ID:             e.ID,
		Name:           e.Name,
		Keywords:       StringSliceToJSON(e.Keywords),
		GenreID:        e.GenreID,
		Status:         int32(e.Status),
		ChangePercent:  e.ChangePercent,
		ZScore:         e.ZScore,
		CurrentVolume:  e.CurrentVolume,
		BaselineVolume: e.BaselineVolume,
		Context:        e.Context,
		ContextSummary: e.ContextSummary,
		SpikeStartedAt: e.SpikeStartedAt,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
	}
}
