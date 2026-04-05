package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// XAnalyticsDailyToEntity converts a GORM model to a domain entity.
func XAnalyticsDailyToEntity(m *model.XAnalyticsDaily) *entity.XAnalyticsDaily {
	return &entity.XAnalyticsDaily{
		ID:            m.ID,
		UserID:        m.UserID,
		Date:          m.Date,
		Impressions:   m.Impressions,
		Likes:         m.Likes,
		Engagements:   m.Engagements,
		Bookmarks:     m.Bookmarks,
		Shares:        m.Shares,
		NewFollows:    m.NewFollows,
		Unfollows:     m.Unfollows,
		Replies:       m.Replies,
		Reposts:       m.Reposts,
		ProfileVisits: m.ProfileVisits,
		PostsCreated:  m.PostsCreated,
		VideoViews:    m.VideoViews,
		MediaViews:    m.MediaViews,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

// XAnalyticsDailyToModel converts a domain entity to a GORM model.
func XAnalyticsDailyToModel(e *entity.XAnalyticsDaily) *model.XAnalyticsDaily {
	return &model.XAnalyticsDaily{
		ID:            e.ID,
		UserID:        e.UserID,
		Date:          e.Date,
		Impressions:   e.Impressions,
		Likes:         e.Likes,
		Engagements:   e.Engagements,
		Bookmarks:     e.Bookmarks,
		Shares:        e.Shares,
		NewFollows:    e.NewFollows,
		Unfollows:     e.Unfollows,
		Replies:       e.Replies,
		Reposts:       e.Reposts,
		ProfileVisits: e.ProfileVisits,
		PostsCreated:  e.PostsCreated,
		VideoViews:    e.VideoViews,
		MediaViews:    e.MediaViews,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}

// XAnalyticsPostToEntity converts a GORM model to a domain entity.
func XAnalyticsPostToEntity(m *model.XAnalyticsPost) *entity.XAnalyticsPost {
	return &entity.XAnalyticsPost{
		ID:              m.ID,
		UserID:          m.UserID,
		PostID:          m.PostID,
		PostedAt:        m.PostedAt,
		PostText:        m.PostText,
		PostURL:         m.PostURL,
		Impressions:     m.Impressions,
		Likes:           m.Likes,
		Engagements:     m.Engagements,
		Bookmarks:       m.Bookmarks,
		Shares:          m.Shares,
		NewFollows:      m.NewFollows,
		Replies:         m.Replies,
		Reposts:         m.Reposts,
		ProfileVisits:   m.ProfileVisits,
		DetailClicks:    m.DetailClicks,
		URLClicks:       m.URLClicks,
		HashtagClicks:   m.HashtagClicks,
		PermalinkClicks: m.PermalinkClicks,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

// XAnalyticsPostToModel converts a domain entity to a GORM model.
func XAnalyticsPostToModel(e *entity.XAnalyticsPost) *model.XAnalyticsPost {
	return &model.XAnalyticsPost{
		ID:              e.ID,
		UserID:          e.UserID,
		PostID:          e.PostID,
		PostedAt:        e.PostedAt,
		PostText:        e.PostText,
		PostURL:         e.PostURL,
		Impressions:     e.Impressions,
		Likes:           e.Likes,
		Engagements:     e.Engagements,
		Bookmarks:       e.Bookmarks,
		Shares:          e.Shares,
		NewFollows:      e.NewFollows,
		Replies:         e.Replies,
		Reposts:         e.Reposts,
		ProfileVisits:   e.ProfileVisits,
		DetailClicks:    e.DetailClicks,
		URLClicks:       e.URLClicks,
		HashtagClicks:   e.HashtagClicks,
		PermalinkClicks: e.PermalinkClicks,
		CreatedAt:       e.CreatedAt,
		UpdatedAt:       e.UpdatedAt,
	}
}
