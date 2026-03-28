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
	"gorm.io/gorm/clause"
)

var _ domainrepo.NotificationSettingRepository = (*notificationSettingRepository)(nil)

type notificationSettingRepository struct {
	db *gorm.DB
}

func NewNotificationSettingRepository(db *gorm.DB) *notificationSettingRepository {
	return &notificationSettingRepository{db: db}
}

func (r *notificationSettingRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

func (r *notificationSettingRepository) FindByUserID(ctx context.Context, userID string) (*entity.NotificationSetting, error) {
	var m model.NotificationSetting
	if err := r.getDB(ctx).Where("user_id = ?", userID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("notification setting not found")
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to find notification setting", err)
	}
	return mapper.NotificationSettingToEntity(&m), nil
}

func (r *notificationSettingRepository) Upsert(ctx context.Context, setting *entity.NotificationSetting) error {
	m := mapper.NotificationSettingToModel(setting)
	// bool の false がゼロ値として GORM の Create でスキップされるのを防ぐため、
	// map を使って全フィールドを明示的に指定する。
	values := map[string]any{
		"user_id":        m.UserID,
		"spike_enabled":  m.SpikeEnabled,
		"rising_enabled": m.RisingEnabled,
		"updated_at":     m.UpdatedAt,
	}
	err := r.getDB(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"spike_enabled", "rising_enabled", "updated_at",
		}),
	}).Model(m).Create(values).Error
	if err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to upsert notification setting", err)
	}
	return nil
}
