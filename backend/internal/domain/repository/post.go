package repository

import (
	"context"

	"github.com/trendbird/backend/internal/domain/entity"
)

// PostRepository defines the persistence operations for Post entities.
type PostRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Post, error)
	ListByUserIDAndStatus(ctx context.Context, userID string, status entity.PostStatus, limit, offset int) ([]*entity.Post, int64, error)
	ListByUserIDAndStatuses(ctx context.Context, userID string, statuses []entity.PostStatus, limit, offset int) ([]*entity.Post, int64, error)
	ListPublishedByUserID(ctx context.Context, userID string, limit, offset int) ([]*entity.Post, int64, error)
	ListScheduled(ctx context.Context) ([]*entity.Post, error)
	Create(ctx context.Context, post *entity.Post) error
	Update(ctx context.Context, post *entity.Post) error
	Delete(ctx context.Context, id string) error
	CountByUserIDAndStatus(ctx context.Context, userID string, status entity.PostStatus) (int64, error)
	CountPublishedByUserIDCurrentMonth(ctx context.Context, userID string) (int32, error)
}
