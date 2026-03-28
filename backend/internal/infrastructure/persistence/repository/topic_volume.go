package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	domainrepo "github.com/trendbird/backend/internal/domain/repository"
	"github.com/trendbird/backend/internal/infrastructure/persistence"
	"github.com/trendbird/backend/internal/infrastructure/persistence/mapper"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
	"gorm.io/gorm"
)

var _ domainrepo.TopicVolumeRepository = (*topicVolumeRepository)(nil)

type topicVolumeRepository struct {
	db *gorm.DB
}

func NewTopicVolumeRepository(db *gorm.DB) *topicVolumeRepository {
	return &topicVolumeRepository{db: db}
}

func (r *topicVolumeRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

func (r *topicVolumeRepository) BulkCreate(ctx context.Context, volumes []*entity.TopicVolume) error {
	if len(volumes) == 0 {
		return nil
	}
	models := make([]model.TopicVolume, len(volumes))
	for i, v := range volumes {
		models[i] = *mapper.TopicVolumeToModel(v)
	}

	// ON CONFLICT (topic_id, timestamp) DO NOTHING で冪等性を保証
	const batchSize = 100
	for i := 0; i < len(models); i += batchSize {
		end := min(i+batchSize, len(models))
		batch := models[i:end]

		placeholders := make([]string, len(batch))
		args := make([]any, 0, len(batch)*3)
		for j, m := range batch {
			placeholders[j] = fmt.Sprintf("(gen_random_uuid(), $%d, $%d, $%d)", j*3+1, j*3+2, j*3+3)
			args = append(args, m.TopicID, m.Timestamp, m.Value)
		}

		query := fmt.Sprintf(
			`INSERT INTO topic_volumes (id, topic_id, timestamp, value) VALUES %s ON CONFLICT (topic_id, timestamp) DO NOTHING`,
			strings.Join(placeholders, ", "),
		)
		if err := r.getDB(ctx).Exec(query, args...).Error; err != nil {
			return apperror.Wrap(apperror.CodeInternal, "failed to bulk create topic volumes", err)
		}
	}
	return nil
}

func (r *topicVolumeRepository) ListByTopicIDAndRange(ctx context.Context, topicID string, from, to time.Time) ([]*entity.TopicVolume, error) {
	var models []model.TopicVolume
	err := r.getDB(ctx).
		Where("topic_id = ? AND timestamp BETWEEN ? AND ?", topicID, from, to).
		Order("timestamp ASC").
		Find(&models).Error
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list topic volumes", err)
	}
	entities := make([]*entity.TopicVolume, len(models))
	for i := range models {
		entities[i] = mapper.TopicVolumeToEntity(&models[i])
	}
	return entities, nil
}
