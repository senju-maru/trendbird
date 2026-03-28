package repository

import (
	"context"
	"time"

	"github.com/trendbird/backend/internal/domain/entity"
)

// TopicRepository defines the persistence operations for Topic entities.
type TopicRepository interface {
	FindByID(ctx context.Context, id string) (*entity.Topic, error)
	FindByIDForUser(ctx context.Context, id string, userID string) (*entity.Topic, error)
	FindByNameAndGenre(ctx context.Context, name string, genre string) (*entity.Topic, error)
	ListAll(ctx context.Context) ([]*entity.Topic, error)
	ListByUserID(ctx context.Context, userID string) ([]*entity.Topic, error)
	GetLatestUpdatedAtByUserID(ctx context.Context, userID string) (*time.Time, error)
	Create(ctx context.Context, topic *entity.Topic) error
	Update(ctx context.Context, topic *entity.Topic) error
	Delete(ctx context.Context, id string) error
	SuggestByName(ctx context.Context, query string, excludeIDs []string, limit int) ([]*entity.TopicSuggestion, error)
	ListByGenreExcluding(ctx context.Context, genreSlug string, excludeIDs []string, limit int) ([]*entity.TopicSuggestion, error)
}
