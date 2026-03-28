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

var _ domainrepo.SpikeHistoryRepository = (*spikeHistoryRepository)(nil)

type spikeHistoryRepository struct {
	db *gorm.DB
}

func NewSpikeHistoryRepository(db *gorm.DB) *spikeHistoryRepository {
	return &spikeHistoryRepository{db: db}
}

func (r *spikeHistoryRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

func (r *spikeHistoryRepository) Create(ctx context.Context, history *entity.SpikeHistory) error {
	m := mapper.SpikeHistoryToModel(history)
	if err := r.getDB(ctx).Create(m).Error; err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to create spike history", err)
	}
	history.ID = m.ID
	history.CreatedAt = m.CreatedAt
	return nil
}

func (r *spikeHistoryRepository) ListByTopicID(ctx context.Context, topicID string) ([]*entity.SpikeHistory, error) {
	var models []model.SpikeHistory
	err := r.getDB(ctx).
		Where("topic_id = ?", topicID).
		Order("timestamp DESC").
		Find(&models).Error
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list spike histories", err)
	}
	entities := make([]*entity.SpikeHistory, len(models))
	for i := range models {
		entities[i] = mapper.SpikeHistoryToEntity(&models[i])
	}
	return entities, nil
}

func (r *spikeHistoryRepository) CountByUserIDCurrentMonth(ctx context.Context, userID string) (int32, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var count int64
	err := r.getDB(ctx).
		Model(&model.SpikeHistory{}).
		Joins("JOIN user_topics ON spike_histories.topic_id = user_topics.topic_id").
		Where("user_topics.user_id = ? AND spike_histories.created_at >= ?", userID, monthStart).
		Count(&count).Error
	if err != nil {
		return 0, apperror.Wrap(apperror.CodeInternal, "failed to count spike histories", err)
	}
	return int32(count), nil
}

// spikeHistoryJoinRow is used for JOIN queries that include topic name.
type spikeHistoryJoinRow struct {
	model.SpikeHistory
	TopicName string `gorm:"column:topic_name"`
}

func (r *spikeHistoryRepository) ListUnnotified(ctx context.Context) ([]*entity.SpikeHistory, error) {
	var rows []spikeHistoryJoinRow
	err := r.getDB(ctx).
		Table("spike_histories").
		Joins("JOIN topics ON spike_histories.topic_id = topics.id").
		Where("spike_histories.notified_at IS NULL").
		Select("spike_histories.*, topics.name AS topic_name").
		Order("spike_histories.created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list unnotified spike histories", err)
	}
	entities := make([]*entity.SpikeHistory, len(rows))
	for i := range rows {
		e := mapper.SpikeHistoryToEntity(&rows[i].SpikeHistory)
		e.TopicName = rows[i].TopicName
		entities[i] = e
	}
	return entities, nil
}

func (r *spikeHistoryRepository) ListUnnotifiedByStatus(ctx context.Context, status entity.TopicStatus) ([]*entity.SpikeHistory, error) {
	var rows []spikeHistoryJoinRow
	err := r.getDB(ctx).
		Table("spike_histories").
		Joins("JOIN topics ON spike_histories.topic_id = topics.id").
		Where("spike_histories.notified_at IS NULL").
		Where("spike_histories.status = ?", int32(status)).
		Select("spike_histories.*, topics.name AS topic_name").
		Order("spike_histories.created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list unnotified spike histories by status", err)
	}
	entities := make([]*entity.SpikeHistory, len(rows))
	for i := range rows {
		e := mapper.SpikeHistoryToEntity(&rows[i].SpikeHistory)
		e.TopicName = rows[i].TopicName
		entities[i] = e
	}
	return entities, nil
}

func (r *spikeHistoryRepository) MarkNotified(ctx context.Context, ids []string, at time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	err := r.getDB(ctx).
		Model(&model.SpikeHistory{}).
		Where("id IN ?", ids).
		Update("notified_at", at).Error
	if err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to mark spike histories as notified", err)
	}
	return nil
}

func (r *spikeHistoryRepository) ListByTopicIDsSince(ctx context.Context, topicIDs []string, since time.Time) ([]*entity.SpikeHistory, error) {
	if len(topicIDs) == 0 {
		return nil, nil
	}
	var rows []spikeHistoryJoinRow
	err := r.getDB(ctx).
		Table("spike_histories").
		Joins("JOIN topics ON spike_histories.topic_id = topics.id").
		Where("spike_histories.topic_id IN ? AND spike_histories.created_at >= ?", topicIDs, since).
		Select("spike_histories.*, topics.name AS topic_name").
		Order("spike_histories.created_at DESC").
		Find(&rows).Error
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list spike histories by topic ids since", err)
	}
	entities := make([]*entity.SpikeHistory, len(rows))
	for i := range rows {
		e := mapper.SpikeHistoryToEntity(&rows[i].SpikeHistory)
		e.TopicName = rows[i].TopicName
		entities[i] = e
	}
	return entities, nil
}
