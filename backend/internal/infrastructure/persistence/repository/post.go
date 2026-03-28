package repository

import (
	"context"
	"errors"
	"time"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	domainrepo "github.com/trendbird/backend/internal/domain/repository"
	"github.com/trendbird/backend/internal/infrastructure/persistence"
	"github.com/trendbird/backend/internal/infrastructure/persistence/mapper"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
	"gorm.io/gorm"
)

var _ domainrepo.PostRepository = (*postRepository)(nil)

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *postRepository {
	return &postRepository{db: db}
}

func (r *postRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

func (r *postRepository) FindByID(ctx context.Context, id string) (*entity.Post, error) {
	var m model.Post
	if err := r.getDB(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("post not found")
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to find post", err)
	}
	return mapper.PostToEntity(&m), nil
}

func (r *postRepository) ListByUserIDAndStatus(ctx context.Context, userID string, status entity.PostStatus, limit, offset int) ([]*entity.Post, int64, error) {
	var total int64
	q := r.getDB(ctx).Model(&model.Post{}).Where("user_id = ? AND status = ?", userID, int32(status))
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperror.Wrap(apperror.CodeInternal, "failed to count posts", err)
	}

	var models []model.Post
	err := r.getDB(ctx).
		Where("user_id = ? AND status = ?", userID, int32(status)).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&models).Error
	if err != nil {
		return nil, 0, apperror.Wrap(apperror.CodeInternal, "failed to list posts", err)
	}

	entities := make([]*entity.Post, len(models))
	for i := range models {
		entities[i] = mapper.PostToEntity(&models[i])
	}
	return entities, total, nil
}

func (r *postRepository) ListByUserIDAndStatuses(ctx context.Context, userID string, statuses []entity.PostStatus, limit, offset int) ([]*entity.Post, int64, error) {
	statusInts := make([]int32, len(statuses))
	for i, s := range statuses {
		statusInts[i] = int32(s)
	}

	var total int64
	q := r.getDB(ctx).Model(&model.Post{}).Where("user_id = ? AND status IN (?)", userID, statusInts)
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, apperror.Wrap(apperror.CodeInternal, "failed to count posts", err)
	}

	var models []model.Post
	err := r.getDB(ctx).
		Where("user_id = ? AND status IN (?)", userID, statusInts).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&models).Error
	if err != nil {
		return nil, 0, apperror.Wrap(apperror.CodeInternal, "failed to list posts", err)
	}

	entities := make([]*entity.Post, len(models))
	for i := range models {
		entities[i] = mapper.PostToEntity(&models[i])
	}
	return entities, total, nil
}

func (r *postRepository) ListPublishedByUserID(ctx context.Context, userID string, limit, offset int) ([]*entity.Post, int64, error) {
	return r.ListByUserIDAndStatus(ctx, userID, entity.PostPublished, limit, offset)
}

func (r *postRepository) ListScheduled(ctx context.Context) ([]*entity.Post, error) {
	var models []model.Post
	err := r.getDB(ctx).
		Where("status = ? AND scheduled_at <= ?", int32(entity.PostScheduled), time.Now()).
		Find(&models).Error
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list scheduled posts", err)
	}
	entities := make([]*entity.Post, len(models))
	for i := range models {
		entities[i] = mapper.PostToEntity(&models[i])
	}
	return entities, nil
}

func (r *postRepository) Create(ctx context.Context, post *entity.Post) error {
	m := mapper.PostToModel(post)
	if err := r.getDB(ctx).Create(m).Error; err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to create post", err)
	}
	post.ID = m.ID
	post.CreatedAt = m.CreatedAt
	post.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *postRepository) Update(ctx context.Context, post *entity.Post) error {
	m := mapper.PostToModel(post)
	if err := r.getDB(ctx).Save(m).Error; err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to update post", err)
	}
	post.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *postRepository) Delete(ctx context.Context, id string) error {
	result := r.getDB(ctx).Where("id = ?", id).Delete(&model.Post{})
	if result.Error != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to delete post", result.Error)
	}
	if result.RowsAffected == 0 {
		return apperror.NotFound("post not found")
	}
	return nil
}

func (r *postRepository) CountByUserIDAndStatus(ctx context.Context, userID string, status entity.PostStatus) (int64, error) {
	var count int64
	err := r.getDB(ctx).
		Model(&model.Post{}).
		Where("user_id = ? AND status = ?", userID, int32(status)).
		Count(&count).Error
	if err != nil {
		return 0, apperror.Wrap(apperror.CodeInternal, "failed to count posts by status", err)
	}
	return count, nil
}

func (r *postRepository) CountPublishedByUserIDCurrentMonth(ctx context.Context, userID string) (int32, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var count int64
	err := r.getDB(ctx).
		Model(&model.Post{}).
		Where("user_id = ? AND status = ? AND published_at >= ?", userID, int32(entity.PostPublished), monthStart).
		Count(&count).Error
	if err != nil {
		return 0, apperror.Wrap(apperror.CodeInternal, "failed to count published posts", err)
	}
	return int32(count), nil
}
