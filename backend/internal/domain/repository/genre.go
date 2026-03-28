package repository

import (
	"context"

	"github.com/trendbird/backend/internal/domain/entity"
)

// GenreRepository defines the persistence operations for Genre entities.
type GenreRepository interface {
	List(ctx context.Context) ([]*entity.Genre, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Genre, error)
	FindByID(ctx context.Context, id string) (*entity.Genre, error)
}
