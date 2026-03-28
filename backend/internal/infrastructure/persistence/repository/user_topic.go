package repository

import (
	"context"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	domainrepo "github.com/trendbird/backend/internal/domain/repository"
	"github.com/trendbird/backend/internal/infrastructure/persistence"
	"github.com/trendbird/backend/internal/infrastructure/persistence/mapper"
	"gorm.io/gorm"
)

var _ domainrepo.UserTopicRepository = (*userTopicRepository)(nil)

type userTopicRepository struct {
	db *gorm.DB
}

func NewUserTopicRepository(db *gorm.DB) *userTopicRepository {
	return &userTopicRepository{db: db}
}

func (r *userTopicRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

// Create inserts a user_topics record. Idempotent: ON CONFLICT DO NOTHING.
func (r *userTopicRepository) Create(ctx context.Context, userTopic *entity.UserTopic) error {
	m := mapper.UserTopicToModel(userTopic)
	result := r.getDB(ctx).Exec(
		`INSERT INTO user_topics (user_id, topic_id, notification_enabled, is_creator)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT (user_id, topic_id) DO NOTHING
		 RETURNING id, created_at`,
		m.UserID, m.TopicID, m.NotificationEnabled, m.IsCreator,
	)
	if result.Error != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to create user topic", result.Error)
	}
	return nil
}

// Delete removes the user-topic link.
func (r *userTopicRepository) Delete(ctx context.Context, userID string, topicID string) error {
	if err := r.getDB(ctx).Exec(
		"DELETE FROM user_topics WHERE user_id = ? AND topic_id = ?", userID, topicID,
	).Error; err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to delete user topic", err)
	}
	return nil
}

// DeleteByUserIDAndGenre removes all user-topic links for a specific genre.
func (r *userTopicRepository) DeleteByUserIDAndGenre(ctx context.Context, userID string, genreID string) error {
	if err := r.getDB(ctx).Exec(
		"DELETE FROM user_topics WHERE user_id = ? AND topic_id IN (SELECT id FROM topics WHERE genre_id = ?)",
		userID, genreID,
	).Error; err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to delete user topics by genre", err)
	}
	return nil
}

// Exists checks if the user-topic link exists.
func (r *userTopicRepository) Exists(ctx context.Context, userID string, topicID string) (bool, error) {
	var exists bool
	err := r.getDB(ctx).Raw(
		"SELECT EXISTS(SELECT 1 FROM user_topics WHERE user_id = ? AND topic_id = ?)",
		userID, topicID,
	).Scan(&exists).Error
	if err != nil {
		return false, apperror.Wrap(apperror.CodeInternal, "failed to check user topic existence", err)
	}
	return exists, nil
}

// CountCreatorByUserID counts the number of topics created (is_creator=true) by the user.
func (r *userTopicRepository) CountCreatorByUserID(ctx context.Context, userID string) (int, error) {
	var count int64
	err := r.getDB(ctx).Raw(
		"SELECT COUNT(*) FROM user_topics WHERE user_id = ? AND is_creator = true", userID,
	).Scan(&count).Error
	if err != nil {
		return 0, apperror.Wrap(apperror.CodeInternal, "failed to count creator user topics", err)
	}
	return int(count), nil
}

// CountByUserID counts the number of topics linked to the user.
func (r *userTopicRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	var count int64
	err := r.getDB(ctx).Raw(
		"SELECT COUNT(*) FROM user_topics WHERE user_id = ?", userID,
	).Scan(&count).Error
	if err != nil {
		return 0, apperror.Wrap(apperror.CodeInternal, "failed to count user topics", err)
	}
	return int(count), nil
}

// UpdateNotificationEnabled updates the notification_enabled flag for a user-topic link.
func (r *userTopicRepository) UpdateNotificationEnabled(ctx context.Context, userID string, topicID string, enabled bool) error {
	result := r.getDB(ctx).Exec(
		"UPDATE user_topics SET notification_enabled = ? WHERE user_id = ? AND topic_id = ?",
		enabled, userID, topicID,
	)
	if result.Error != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to update user topic notification", result.Error)
	}
	if result.RowsAffected == 0 {
		return apperror.NotFound("user topic not found")
	}
	return nil
}

// ListUserIDsByTopicID returns user IDs subscribed to the given topic.
func (r *userTopicRepository) ListUserIDsByTopicID(ctx context.Context, topicID string, notificationEnabledOnly bool) ([]string, error) {
	query := "SELECT user_id FROM user_topics WHERE topic_id = ?"
	if notificationEnabledOnly {
		query += " AND notification_enabled = true"
	}
	var userIDs []string
	err := r.getDB(ctx).Raw(query, topicID).Scan(&userIDs).Error
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list user ids by topic id", err)
	}
	return userIDs, nil
}

// ListTopicIDsByUserID returns topic IDs subscribed by the given user.
func (r *userTopicRepository) ListTopicIDsByUserID(ctx context.Context, userID string) ([]string, error) {
	var topicIDs []string
	err := r.getDB(ctx).Raw(
		"SELECT topic_id FROM user_topics WHERE user_id = ?", userID,
	).Scan(&topicIDs).Error
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list topic ids by user id", err)
	}
	return topicIDs, nil
}
