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

var _ domainrepo.GeneratedPostRepository = (*generatedPostRepository)(nil)

type generatedPostRepository struct {
	db *gorm.DB
}

func NewGeneratedPostRepository(db *gorm.DB) *generatedPostRepository {
	return &generatedPostRepository{db: db}
}

func (r *generatedPostRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

func (r *generatedPostRepository) BulkCreate(ctx context.Context, posts []*entity.GeneratedPost) error {
	if len(posts) == 0 {
		return nil
	}
	models := make([]*model.GeneratedPost, len(posts))
	for i, p := range posts {
		models[i] = mapper.GeneratedPostToModel(p)
	}
	if err := r.getDB(ctx).CreateInBatches(models, 100).Error; err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to bulk create generated posts", err)
	}
	// Write back DB-generated IDs to entities
	for i, m := range models {
		posts[i].ID = m.ID
	}
	return nil
}

func (r *generatedPostRepository) ListByUserIDAndTopicID(ctx context.Context, userID, topicID string) ([]*entity.GeneratedPost, error) {
	var models []model.GeneratedPost
	err := r.getDB(ctx).
		Where("user_id = ? AND topic_id = ?", userID, topicID).
		Order("created_at DESC").
		Find(&models).Error
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list generated posts", err)
	}
	entities := make([]*entity.GeneratedPost, len(models))
	for i := range models {
		entities[i] = mapper.GeneratedPostToEntity(&models[i])
	}
	return entities, nil
}
