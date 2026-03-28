package repository

import (
	"context"
	"time"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	domainrepo "github.com/trendbird/backend/internal/domain/repository"
	"github.com/trendbird/backend/internal/infrastructure/persistence"
	"github.com/trendbird/backend/internal/infrastructure/persistence/mapper"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
	"gorm.io/gorm"
)

var _ domainrepo.AIGenerationLogRepository = (*aiGenerationLogRepository)(nil)

type aiGenerationLogRepository struct {
	db *gorm.DB
}

func NewAIGenerationLogRepository(db *gorm.DB) *aiGenerationLogRepository {
	return &aiGenerationLogRepository{db: db}
}

func (r *aiGenerationLogRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

func (r *aiGenerationLogRepository) Create(ctx context.Context, log *entity.AIGenerationLog) error {
	m := mapper.AIGenerationLogToModel(log)
	if err := r.getDB(ctx).Create(m).Error; err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to create ai generation log", err)
	}
	log.ID = m.ID
	log.CreatedAt = m.CreatedAt
	return nil
}

func (r *aiGenerationLogRepository) CountByUserIDCurrentMonth(ctx context.Context, userID string) (int32, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var count int64
	err := r.getDB(ctx).
		Model(&model.AIGenerationLog{}).
		Where("user_id = ? AND created_at >= ?", userID, monthStart).
		Count(&count).Error
	if err != nil {
		return 0, apperror.Wrap(apperror.CodeInternal, "failed to count ai generation logs", err)
	}
	return int32(count), nil
}
