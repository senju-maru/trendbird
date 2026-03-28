package repository

import (
	"context"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	domainrepo "github.com/trendbird/backend/internal/domain/repository"
	"github.com/trendbird/backend/internal/infrastructure/persistence"
	"github.com/trendbird/backend/internal/infrastructure/persistence/mapper"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
	"gorm.io/gorm"
)

var _ domainrepo.UserGenreRepository = (*userGenreRepository)(nil)

type userGenreRepository struct {
	db *gorm.DB
}

func NewUserGenreRepository(db *gorm.DB) *userGenreRepository {
	return &userGenreRepository{db: db}
}

func (r *userGenreRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

func (r *userGenreRepository) ListByUserID(ctx context.Context, userID string) ([]*entity.UserGenre, error) {
	type userGenreWithSlug struct {
		model.UserGenre
		GenreSlug string `gorm:"column:genre_slug"`
	}
	var results []userGenreWithSlug
	if err := r.getDB(ctx).
		Table("user_genres").
		Select("user_genres.*, genres.slug as genre_slug").
		Joins("JOIN genres ON user_genres.genre_id = genres.id").
		Where("user_genres.user_id = ?", userID).
		Order("user_genres.created_at ASC").
		Find(&results).Error; err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list user genres", err)
	}
	entities := make([]*entity.UserGenre, len(results))
	for i := range results {
		entities[i] = mapper.UserGenreToEntity(&results[i].UserGenre)
		entities[i].GenreSlug = results[i].GenreSlug
	}
	return entities, nil
}

func (r *userGenreRepository) Create(ctx context.Context, genre *entity.UserGenre) error {
	m := mapper.UserGenreToModel(genre)
	// ON CONFLICT DO NOTHING for idempotency
	result := r.getDB(ctx).Exec(
		`INSERT INTO user_genres (user_id, genre_id) VALUES (?, ?) ON CONFLICT (user_id, genre_id) DO NOTHING`,
		m.UserID, m.GenreID,
	)
	if result.Error != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to create user genre", result.Error)
	}
	return nil
}

func (r *userGenreRepository) DeleteByUserIDAndGenre(ctx context.Context, userID string, genreID string) error {
	result := r.getDB(ctx).Where("user_id = ? AND genre_id = ?", userID, genreID).Delete(&model.UserGenre{})
	if result.Error != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to delete user genre", result.Error)
	}
	return nil
}

func (r *userGenreRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	var count int64
	if err := r.getDB(ctx).Model(&model.UserGenre{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, apperror.Wrap(apperror.CodeInternal, "failed to count user genres", err)
	}
	return int(count), nil
}

func (r *userGenreRepository) ExistsByUserIDAndGenre(ctx context.Context, userID string, genreID string) (bool, error) {
	var count int64
	if err := r.getDB(ctx).Model(&model.UserGenre{}).Where("user_id = ? AND genre_id = ?", userID, genreID).Count(&count).Error; err != nil {
		return false, apperror.Wrap(apperror.CodeInternal, "failed to check user genre existence", err)
	}
	return count > 0, nil
}
