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

var _ domainrepo.ActivityRepository = (*activityRepository)(nil)

type activityRepository struct {
	db *gorm.DB
}

func NewActivityRepository(db *gorm.DB) *activityRepository {
	return &activityRepository{db: db}
}

func (r *activityRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

func (r *activityRepository) Create(ctx context.Context, activity *entity.Activity) error {
	m := mapper.ActivityToModel(activity)
	if err := r.getDB(ctx).Create(m).Error; err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to create activity", err)
	}
	activity.ID = m.ID
	activity.CreatedAt = m.CreatedAt
	return nil
}

func (r *activityRepository) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*entity.Activity, error) {
	var models []model.Activity
	err := r.getDB(ctx).
		Where("user_id = ?", userID).
		Order("timestamp DESC").
		Limit(limit).Offset(offset).
		Find(&models).Error
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list activities", err)
	}
	entities := make([]*entity.Activity, len(models))
	for i := range models {
		entities[i] = mapper.ActivityToEntity(&models[i])
	}
	return entities, nil
}
