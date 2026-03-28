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
	"gorm.io/gorm/clause"
)

var _ domainrepo.ReplyPendingQueueRepository = (*replyPendingQueueRepository)(nil)

type replyPendingQueueRepository struct {
	db *gorm.DB
}

func NewReplyPendingQueueRepository(db *gorm.DB) *replyPendingQueueRepository {
	return &replyPendingQueueRepository{db: db}
}

func (r *replyPendingQueueRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

// Create inserts a new pending reply. Idempotent: ON CONFLICT DO NOTHING.
func (r *replyPendingQueueRepository) Create(ctx context.Context, item *entity.ReplyPendingQueue) error {
	m := mapper.ReplyPendingQueueToModel(item)
	if err := r.getDB(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(m).Error; err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to create reply pending queue item", err)
	}
	if m.ID != "" {
		item.ID = m.ID
		item.CreatedAt = m.CreatedAt
	}
	return nil
}

func (r *replyPendingQueueRepository) ListPendingGroupedByUser(ctx context.Context) (map[string][]*entity.ReplyPendingQueue, error) {
	var models []model.ReplyPendingQueue
	if err := r.getDB(ctx).
		Where("status = ?", int(entity.ReplyPendingStatusPending)).
		Order("created_at ASC").
		Find(&models).Error; err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list pending reply queue grouped", err)
	}
	result := make(map[string][]*entity.ReplyPendingQueue)
	for i := range models {
		e := mapper.ReplyPendingQueueToEntity(&models[i])
		result[e.UserID] = append(result[e.UserID], e)
	}
	return result, nil
}

func (r *replyPendingQueueRepository) UpdateStatus(ctx context.Context, id string, status entity.ReplyPendingStatus) error {
	result := r.getDB(ctx).Exec(
		"UPDATE reply_pending_queue SET status = ? WHERE id = ?",
		int(status), id,
	)
	if result.Error != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to update reply pending queue status", result.Error)
	}
	return nil
}

func (r *replyPendingQueueRepository) DeleteByID(ctx context.Context, id string) error {
	if err := r.getDB(ctx).Exec("DELETE FROM reply_pending_queue WHERE id = ?", id).Error; err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to delete reply pending queue item", err)
	}
	return nil
}
