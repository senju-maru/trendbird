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

var _ domainrepo.DMSentLogRepository = (*dmSentLogRepository)(nil)

type dmSentLogRepository struct {
	db *gorm.DB
}

func NewDMSentLogRepository(db *gorm.DB) *dmSentLogRepository {
	return &dmSentLogRepository{db: db}
}

func (r *dmSentLogRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

func (r *dmSentLogRepository) Create(ctx context.Context, log *entity.DMSentLog) error {
	m := mapper.DMSentLogToModel(log)
	if err := r.getDB(ctx).Create(m).Error; err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to create dm sent log", err)
	}
	log.ID = m.ID
	log.SentAt = m.SentAt
	return nil
}

func (r *dmSentLogRepository) ExistsByReplyTweetID(ctx context.Context, replyTweetID string, recipientTwitterID string) (bool, error) {
	var exists bool
	err := r.getDB(ctx).Raw(
		"SELECT EXISTS(SELECT 1 FROM dm_sent_logs WHERE reply_tweet_id = ? AND recipient_twitter_id = ?)",
		replyTweetID, recipientTwitterID,
	).Scan(&exists).Error
	if err != nil {
		return false, apperror.Wrap(apperror.CodeInternal, "failed to check dm sent log existence", err)
	}
	return exists, nil
}

func (r *dmSentLogRepository) ListByUserID(ctx context.Context, userID string, limit int) ([]*entity.DMSentLog, error) {
	var models []model.DMSentLog
	if err := r.getDB(ctx).
		Where("user_id = ?", userID).
		Order("sent_at DESC").
		Limit(limit).
		Find(&models).Error; err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list dm sent logs", err)
	}
	entities := make([]*entity.DMSentLog, len(models))
	for i := range models {
		entities[i] = mapper.DMSentLogToEntity(&models[i])
	}
	return entities, nil
}
