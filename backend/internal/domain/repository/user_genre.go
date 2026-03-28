package repository

import (
	"context"

	"github.com/trendbird/backend/internal/domain/entity"
)

// UserGenreRepository defines the persistence operations for UserGenre entities.
// genre 引数はすべて genre_id (UUID) を受け取る。
type UserGenreRepository interface {
	ListByUserID(ctx context.Context, userID string) ([]*entity.UserGenre, error)
	Create(ctx context.Context, genre *entity.UserGenre) error
	DeleteByUserIDAndGenre(ctx context.Context, userID string, genreID string) error
	CountByUserID(ctx context.Context, userID string) (int, error)
	ExistsByUserIDAndGenre(ctx context.Context, userID string, genreID string) (bool, error)
}
