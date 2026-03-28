package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// PostToEntity converts a GORM Post model to a domain entity.
func PostToEntity(m *model.Post) *entity.Post {
	return &entity.Post{
		ID:           m.ID,
		UserID:       m.UserID,
		Content:      m.Content,
		TopicID:      m.TopicID,
		TopicName:    m.TopicName,
		Status:       entity.PostStatus(m.Status),
		ScheduledAt:  m.ScheduledAt,
		PublishedAt:  m.PublishedAt,
		FailedAt:     m.FailedAt,
		ErrorMessage: m.ErrorMessage,
		TweetURL:     m.TweetURL,
		Likes:        m.Likes,
		Retweets:     m.Retweets,
		Replies:      m.Replies,
		Views:        m.Views,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// PostToModel converts a domain Post entity to a GORM model.
func PostToModel(e *entity.Post) *model.Post {
	return &model.Post{
		ID:           e.ID,
		UserID:       e.UserID,
		Content:      e.Content,
		TopicID:      e.TopicID,
		TopicName:    e.TopicName,
		Status:       int32(e.Status),
		ScheduledAt:  e.ScheduledAt,
		PublishedAt:  e.PublishedAt,
		FailedAt:     e.FailedAt,
		ErrorMessage: e.ErrorMessage,
		TweetURL:     e.TweetURL,
		Likes:        e.Likes,
		Retweets:     e.Retweets,
		Replies:      e.Replies,
		Views:        e.Views,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}
}
