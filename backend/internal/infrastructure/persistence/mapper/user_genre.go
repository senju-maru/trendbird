package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// UserGenreToEntity converts a GORM UserGenre model to a domain entity.
func UserGenreToEntity(m *model.UserGenre) *entity.UserGenre {
	return &entity.UserGenre{
		ID:        m.ID,
		UserID:    m.UserID,
		GenreID:   m.GenreID,
		CreatedAt: m.CreatedAt,
	}
}

// UserGenreToModel converts a domain UserGenre entity to a GORM model.
func UserGenreToModel(e *entity.UserGenre) *model.UserGenre {
	return &model.UserGenre{
		ID:        e.ID,
		UserID:    e.UserID,
		GenreID:   e.GenreID,
		CreatedAt: e.CreatedAt,
	}
}
