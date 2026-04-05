package converter

import (
	"time"

	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/internal/domain/entity"
)

// DailyAnalyticsToProto converts a domain entity to Proto message.
func DailyAnalyticsToProto(e *entity.XAnalyticsDaily) *trendbirdv1.DailyAnalytics {
	return &trendbirdv1.DailyAnalytics{
		Id:            e.ID,
		Date:          e.Date.Format("2006-01-02"),
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
	}
}

// DailyAnalyticsSliceToProto converts a slice of entities to Proto messages.
func DailyAnalyticsSliceToProto(es []*entity.XAnalyticsDaily) []*trendbirdv1.DailyAnalytics {
	result := make([]*trendbirdv1.DailyAnalytics, len(es))
	for i, e := range es {
		result[i] = DailyAnalyticsToProto(e)
	}
	return result
}

// PostAnalyticsToProto converts a domain entity to Proto message.
func PostAnalyticsToProto(e *entity.XAnalyticsPost) *trendbirdv1.PostAnalytics {
	return &trendbirdv1.PostAnalytics{
		Id:              e.ID,
		PostId:          e.PostID,
		PostedAt:        e.PostedAt.Format(time.RFC3339),
		PostText:        e.PostText,
		PostUrl:         e.PostURL,
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
		UrlClicks:       e.URLClicks,
		HashtagClicks:   e.HashtagClicks,
		PermalinkClicks: e.PermalinkClicks,
	}
}

// PostAnalyticsSliceToProto converts a slice of entities to Proto messages.
func PostAnalyticsSliceToProto(es []*entity.XAnalyticsPost) []*trendbirdv1.PostAnalytics {
	result := make([]*trendbirdv1.PostAnalytics, len(es))
	for i, e := range es {
		result[i] = PostAnalyticsToProto(e)
	}
	return result
}

// AnalyticsSummaryToProto converts a domain summary to Proto message.
func AnalyticsSummaryToProto(e *entity.AnalyticsSummary) *trendbirdv1.AnalyticsSummary {
	return &trendbirdv1.AnalyticsSummary{
		StartDate:        e.StartDate,
		EndDate:          e.EndDate,
		TotalImpressions: e.TotalImpressions,
		TotalLikes:       e.TotalLikes,
		TotalEngagements: e.TotalEngagements,
		TotalNewFollows:  e.TotalNewFollows,
		TotalUnfollows:   e.TotalUnfollows,
		DaysCount:        e.DaysCount,
		PostsCount:       e.PostsCount,
		DailyData:        DailyAnalyticsSliceToProto(e.DailyData),
	}
}

// GrowthInsightToProto converts a domain insight to Proto message.
func GrowthInsightToProto(e *entity.GrowthInsight) *trendbirdv1.GrowthInsight {
	return &trendbirdv1.GrowthInsight{
		Category: e.Category,
		Insight:  e.Insight,
		Action:   e.Action,
	}
}

// GrowthInsightSliceToProto converts a slice of insights to Proto messages.
func GrowthInsightSliceToProto(es []*entity.GrowthInsight) []*trendbirdv1.GrowthInsight {
	result := make([]*trendbirdv1.GrowthInsight, len(es))
	for i, e := range es {
		result[i] = GrowthInsightToProto(e)
	}
	return result
}

// DailyAnalyticsFromProto converts a Proto message to a domain entity (for import).
func DailyAnalyticsFromProto(p *trendbirdv1.DailyAnalytics) (*entity.XAnalyticsDaily, error) {
	date, err := time.Parse("2006-01-02", p.GetDate())
	if err != nil {
		return nil, err
	}
	return &entity.XAnalyticsDaily{
		Date:          date,
		Impressions:   p.GetImpressions(),
		Likes:         p.GetLikes(),
		Engagements:   p.GetEngagements(),
		Bookmarks:     p.GetBookmarks(),
		Shares:        p.GetShares(),
		NewFollows:    p.GetNewFollows(),
		Unfollows:     p.GetUnfollows(),
		Replies:       p.GetReplies(),
		Reposts:       p.GetReposts(),
		ProfileVisits: p.GetProfileVisits(),
		PostsCreated:  p.GetPostsCreated(),
		VideoViews:    p.GetVideoViews(),
		MediaViews:    p.GetMediaViews(),
	}, nil
}

// PostAnalyticsFromProto converts a Proto message to a domain entity (for import).
func PostAnalyticsFromProto(p *trendbirdv1.PostAnalytics) (*entity.XAnalyticsPost, error) {
	postedAt, err := time.Parse(time.RFC3339, p.GetPostedAt())
	if err != nil {
		return nil, err
	}
	return &entity.XAnalyticsPost{
		PostID:          p.GetPostId(),
		PostedAt:        postedAt,
		PostText:        p.GetPostText(),
		PostURL:         p.GetPostUrl(),
		Impressions:     p.GetImpressions(),
		Likes:           p.GetLikes(),
		Engagements:     p.GetEngagements(),
		Bookmarks:       p.GetBookmarks(),
		Shares:          p.GetShares(),
		NewFollows:      p.GetNewFollows(),
		Replies:         p.GetReplies(),
		Reposts:         p.GetReposts(),
		ProfileVisits:   p.GetProfileVisits(),
		DetailClicks:    p.GetDetailClicks(),
		URLClicks:       p.GetUrlClicks(),
		HashtagClicks:   p.GetHashtagClicks(),
		PermalinkClicks: p.GetPermalinkClicks(),
	}, nil
}
