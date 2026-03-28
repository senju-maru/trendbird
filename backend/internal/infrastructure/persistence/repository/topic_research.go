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

var _ domainrepo.TopicResearchRepository = (*topicResearchRepository)(nil)

type topicResearchRepository struct {
	db *gorm.DB
}

func NewTopicResearchRepository(db *gorm.DB) *topicResearchRepository {
	return &topicResearchRepository{db: db}
}

func (r *topicResearchRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

func (r *topicResearchRepository) Create(ctx context.Context, research *entity.TopicResearch) error {
	m := mapper.TopicResearchToModel(research)
	if err := r.getDB(ctx).Create(m).Error; err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to create topic research", err)
	}
	research.ID = m.ID
	research.CreatedAt = m.CreatedAt
	return nil
}

func (r *topicResearchRepository) ListByTopicID(ctx context.Context, topicID string, limit int) ([]*entity.TopicResearch, error) {
	var models []model.TopicResearch
	err := r.getDB(ctx).
		Where("topic_id = ?", topicID).
		Order("searched_at DESC").
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list topic research", err)
	}
	entities := make([]*entity.TopicResearch, len(models))
	for i := range models {
		entities[i] = mapper.TopicResearchToEntity(&models[i])
	}
	return entities, nil
}

func (r *topicResearchRepository) ListLatestByTopicID(ctx context.Context, topicID string, since time.Time) ([]*entity.TopicResearch, error) {
	var models []model.TopicResearch
	err := r.getDB(ctx).
		Where("topic_id = ? AND searched_at >= ?", topicID, since).
		Order("searched_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list latest topic research", err)
	}
	entities := make([]*entity.TopicResearch, len(models))
	for i := range models {
		entities[i] = mapper.TopicResearchToEntity(&models[i])
	}
	return entities, nil
}
