package repository

import (
	"context"

	"github.com/trendbird/backend/internal/domain/entity"
)

// UserTopicRepository defines the persistence operations for UserTopic entities.
type UserTopicRepository interface {
	Create(ctx context.Context, userTopic *entity.UserTopic) error
	Delete(ctx context.Context, userID string, topicID string) error
	DeleteByUserIDAndGenre(ctx context.Context, userID string, genreID string) error
	Exists(ctx context.Context, userID string, topicID string) (bool, error)
	CountByUserID(ctx context.Context, userID string) (int, error)
	CountCreatorByUserID(ctx context.Context, userID string) (int, error)
	UpdateNotificationEnabled(ctx context.Context, userID string, topicID string, enabled bool) error
	// ListUserIDsByTopicID returns user IDs subscribed to the given topic.
	// If notificationEnabledOnly is true, only users with notification_enabled = true are returned.
	ListUserIDsByTopicID(ctx context.Context, topicID string, notificationEnabledOnly bool) ([]string, error)
	// ListTopicIDsByUserID returns topic IDs subscribed by the given user.
	ListTopicIDsByUserID(ctx context.Context, userID string) ([]string, error)
}
