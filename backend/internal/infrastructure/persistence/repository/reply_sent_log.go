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

var _ domainrepo.ReplySentLogRepository = (*replySentLogRepository)(nil)

type replySentLogRepository struct {
	db *gorm.DB
}

func NewReplySentLogRepository(db *gorm.DB) *replySentLogRepository {
	return &replySentLogRepository{db: db}
}

func (r *replySentLogRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

func (r *replySentLogRepository) Create(ctx context.Context, log *entity.ReplySentLog) error {
	m := mapper.ReplySentLogToModel(log)
	if err := r.getDB(ctx).Create(m).Error; err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to create reply sent log", err)
	}
	log.ID = m.ID
	log.SentAt = m.SentAt
	return nil
}

func (r *replySentLogRepository) ExistsByOriginalTweetID(ctx context.Context, originalTweetID string, originalAuthorID string) (bool, error) {
	var exists bool
	err := r.getDB(ctx).Raw(
		"SELECT EXISTS(SELECT 1 FROM reply_sent_logs WHERE original_tweet_id = ? AND original_author_id = ?)",
		originalTweetID, originalAuthorID,
	).Scan(&exists).Error
	if err != nil {
		return false, apperror.Wrap(apperror.CodeInternal, "failed to check reply sent log existence", err)
	}
	return exists, nil
}

func (r *replySentLogRepository) ListByUserID(ctx context.Context, userID string, limit int) ([]*entity.ReplySentLog, error) {
	var models []model.ReplySentLog
	if err := r.getDB(ctx).
		Where("user_id = ?", userID).
		Order("sent_at DESC").
		Limit(limit).
		Find(&models).Error; err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list reply sent logs", err)
	}
	entities := make([]*entity.ReplySentLog, len(models))
	for i := range models {
		entities[i] = mapper.ReplySentLogToEntity(&models[i])
	}
	return entities, nil
}
