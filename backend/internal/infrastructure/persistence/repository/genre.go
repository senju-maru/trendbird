package repository

import (
	"context"
	"errors"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	domainrepo "github.com/trendbird/backend/internal/domain/repository"
	"github.com/trendbird/backend/internal/infrastructure/persistence"
	"github.com/trendbird/backend/internal/infrastructure/persistence/mapper"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
	"gorm.io/gorm"
)

var _ domainrepo.GenreRepository = (*genreRepository)(nil)

type genreRepository struct {
	db *gorm.DB
}

func NewGenreRepository(db *gorm.DB) *genreRepository {
	return &genreRepository{db: db}
}

func (r *genreRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

func (r *genreRepository) List(ctx context.Context) ([]*entity.Genre, error) {
	var models []model.Genre
	if err := r.getDB(ctx).Order("sort_order ASC").Find(&models).Error; err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list genres", err)
	}
	entities := make([]*entity.Genre, len(models))
	for i := range models {
		entities[i] = mapper.GenreToEntity(&models[i])
	}
	return entities, nil
}

func (r *genreRepository) FindBySlug(ctx context.Context, slug string) (*entity.Genre, error) {
	var m model.Genre
	if err := r.getDB(ctx).Where("slug = ?", slug).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("genre not found")
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to find genre", err)
	}
	return mapper.GenreToEntity(&m), nil
}

func (r *genreRepository) FindByID(ctx context.Context, id string) (*entity.Genre, error) {
	var m model.Genre
	if err := r.getDB(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("genre not found")
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to find genre", err)
	}
	return mapper.GenreToEntity(&m), nil
}
