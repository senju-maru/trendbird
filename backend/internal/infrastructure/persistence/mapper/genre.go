package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// GenreToEntity converts a GORM Genre model to a domain entity.
func GenreToEntity(m *model.Genre) *entity.Genre {
	return &entity.Genre{
		ID:          m.ID,
		Slug:        m.Slug,
		Label:       m.Label,
		Description: m.Description,
		SortOrder:   m.SortOrder,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// GenreToModel converts a domain Genre entity to a GORM model.
func GenreToModel(e *entity.Genre) *model.Genre {
	return &model.Genre{
		ID:          e.ID,
		Slug:        e.Slug,
		Label:       e.Label,
		Description: e.Description,
		SortOrder:   e.SortOrder,
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
	}
}
