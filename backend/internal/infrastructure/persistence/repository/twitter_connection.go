package repository

import (
	"context"
	"errors"
	"time"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	domainrepo "github.com/trendbird/backend/internal/domain/repository"
	"github.com/trendbird/backend/internal/infrastructure/persistence"
	"github.com/trendbird/backend/internal/infrastructure/persistence/mapper"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ domainrepo.TwitterConnectionRepository = (*twitterConnectionRepository)(nil)

type twitterConnectionRepository struct {
	db *gorm.DB
}

func NewTwitterConnectionRepository(db *gorm.DB) *twitterConnectionRepository {
	return &twitterConnectionRepository{db: db}
}

func (r *twitterConnectionRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

func (r *twitterConnectionRepository) FindByUserID(ctx context.Context, userID string) (*entity.TwitterConnection, error) {
	var m model.TwitterConnection
	if err := r.getDB(ctx).Where("user_id = ?", userID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("twitter connection not found")
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to find twitter connection", err)
	}
	return mapper.TwitterConnectionToEntity(&m), nil
}

func (r *twitterConnectionRepository) Upsert(ctx context.Context, conn *entity.TwitterConnection) error {
	m := mapper.TwitterConnectionToModel(conn)
	err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"access_token", "refresh_token", "token_expires_at",
			"status", "connected_at", "last_tested_at",
			"error_message", "updated_at",
		}),
	}).Create(m).Error
	if err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to upsert twitter connection", err)
	}
	return nil
}

func (r *twitterConnectionRepository) UpdateStatus(ctx context.Context, userID string, status entity.TwitterConnectionStatus, errorMessage *string) error {
	result := r.getDB(ctx).Model(&model.TwitterConnection{}).
		Where("user_id = ?", userID).
		Updates(map[string]any{
			"status":        int32(status),
			"error_message": errorMessage,
		})
	if result.Error != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to update twitter connection status", result.Error)
	}
	if result.RowsAffected == 0 {
		return apperror.NotFound("twitter connection not found")
	}
	return nil
}

func (r *twitterConnectionRepository) UpdateLastTestedAt(ctx context.Context, userID string) error {
	result := r.getDB(ctx).Model(&model.TwitterConnection{}).
		Where("user_id = ?", userID).
		Update("last_tested_at", time.Now())
	if result.Error != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to update last tested at", result.Error)
	}
	if result.RowsAffected == 0 {
		return apperror.NotFound("twitter connection not found")
	}
	return nil
}

func (r *twitterConnectionRepository) DeleteByUserID(ctx context.Context, userID string) error {
	result := r.getDB(ctx).Where("user_id = ?", userID).Delete(&model.TwitterConnection{})
	if result.Error != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to delete twitter connection", result.Error)
	}
	if result.RowsAffected == 0 {
		return apperror.NotFound("twitter connection not found")
	}
	return nil
}
