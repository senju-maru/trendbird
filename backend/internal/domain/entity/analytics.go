package entity

import "time"

// XAnalyticsDaily represents a daily aggregate metrics record from X analytics.
type XAnalyticsDaily struct {
	ID            string
	UserID        string
	Date          time.Time
	Impressions   int32
	Likes         int32
	Engagements   int32
	Bookmarks     int32
	Shares        int32
	NewFollows    int32
	Unfollows     int32
	Replies       int32
	Reposts       int32
	ProfileVisits int32
	PostsCreated  int32
	VideoViews    int32
	MediaViews    int32
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// XAnalyticsPost represents per-post metrics from X analytics.
type XAnalyticsPost struct {
	ID              string
	UserID          string
	PostID          string
	PostedAt        time.Time
	PostText        string
	PostURL         string
	Impressions     int32
	Likes           int32
	Engagements     int32
	Bookmarks       int32
	Shares          int32
	NewFollows      int32
	Replies         int32
	Reposts         int32
	ProfileVisits   int32
	DetailClicks    int32
	URLClicks       int32
	HashtagClicks   int32
	PermalinkClicks int32
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// AnalyticsSummary holds aggregated analytics for a date range.
type AnalyticsSummary struct {
	StartDate        string
	EndDate          string
	TotalImpressions int64
	TotalLikes       int64
	TotalEngagements int64
	TotalNewFollows  int64
	TotalUnfollows   int64
	DaysCount        int32
	PostsCount       int32
	DailyData        []*XAnalyticsDaily
}

// GrowthInsight represents a data-driven insight with a recommended action.
type GrowthInsight struct {
	Category string
	Insight  string
	Action   string
}
