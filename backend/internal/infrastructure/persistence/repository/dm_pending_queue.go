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

var _ domainrepo.DMPendingQueueRepository = (*dmPendingQueueRepository)(nil)

type dmPendingQueueRepository struct {
	db *gorm.DB
}

func NewDMPendingQueueRepository(db *gorm.DB) *dmPendingQueueRepository {
	return &dmPendingQueueRepository{db: db}
}

func (r *dmPendingQueueRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

// Create inserts a new pending DM. Idempotent: ON CONFLICT DO NOTHING.
func (r *dmPendingQueueRepository) Create(ctx context.Context, item *entity.DMPendingQueue) error {
	m := mapper.DMPendingQueueToModel(item)
	if err := r.getDB(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(m).Error; err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to create dm pending queue item", err)
	}
	if m.ID != "" {
		item.ID = m.ID
		item.CreatedAt = m.CreatedAt
	}
	return nil
}

func (r *dmPendingQueueRepository) ListPendingByUserID(ctx context.Context, userID string, limit int) ([]*entity.DMPendingQueue, error) {
	var models []model.DMPendingQueue
	if err := r.getDB(ctx).
		Where("user_id = ? AND status = ?", userID, int(entity.DMPendingStatusPending)).
		Order("created_at ASC").
		Limit(limit).
		Find(&models).Error; err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list pending dm queue", err)
	}
	entities := make([]*entity.DMPendingQueue, len(models))
	for i := range models {
		entities[i] = mapper.DMPendingQueueToEntity(&models[i])
	}
	return entities, nil
}

func (r *dmPendingQueueRepository) ListPendingGroupedByUser(ctx context.Context) (map[string][]*entity.DMPendingQueue, error) {
	var models []model.DMPendingQueue
	if err := r.getDB(ctx).
		Where("status = ?", int(entity.DMPendingStatusPending)).
		Order("created_at ASC").
		Find(&models).Error; err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list pending dm queue grouped", err)
	}
	result := make(map[string][]*entity.DMPendingQueue)
	for i := range models {
		e := mapper.DMPendingQueueToEntity(&models[i])
		result[e.UserID] = append(result[e.UserID], e)
	}
	return result, nil
}

func (r *dmPendingQueueRepository) UpdateStatus(ctx context.Context, id string, status entity.DMPendingStatus) error {
	result := r.getDB(ctx).Exec(
		"UPDATE dm_pending_queue SET status = ? WHERE id = ?",
		int(status), id,
	)
	if result.Error != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to update dm pending queue status", result.Error)
	}
	return nil
}

func (r *dmPendingQueueRepository) DeleteByID(ctx context.Context, id string) error {
	if err := r.getDB(ctx).Exec("DELETE FROM dm_pending_queue WHERE id = ?", id).Error; err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to delete dm pending queue item", err)
	}
	return nil
}
