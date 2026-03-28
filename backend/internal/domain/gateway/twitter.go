package gateway

import (
	"context"
	"time"
)

// OAuthTokenResponse holds the tokens returned by the X OAuth 2.0 flow.
type OAuthTokenResponse struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
}

// TwitterUserInfo holds the user profile returned by the X API.
type TwitterUserInfo struct {
	ID       string
	Name     string
	Username string
	Email    string
	Image    string
}

// SearchTweetsInput holds the parameters for searching recent tweets.
type SearchTweetsInput struct {
	Query      string
	MaxResults int
	StartTime  *time.Time
	EndTime    *time.Time
	SinceID    *string // exclude tweets at or before this ID; takes precedence over StartTime when set
	SortOrder  string  // "recency" (default) or "relevancy"
}

// Tweet represents a single tweet returned by the X API.
type Tweet struct {
	ID             string
	Text           string
	AuthorID       string
	AuthorName     string
	AuthorHandle   string
	ConversationID string
	CreatedAt      time.Time
	Metrics        TweetMetrics
}

// TweetMetrics holds engagement metrics for a tweet.
type TweetMetrics struct {
	RetweetCount    int
	ReplyCount      int
	LikeCount       int
	QuoteCount      int
	ImpressionCount int
}

// TweetCountDataPoint represents a single data point of tweet volume over time.
type TweetCountDataPoint struct {
	Start      time.Time
	End        time.Time
	TweetCount int
}

// OAuthStartResult holds the parameters needed to redirect the user to X's authorization page.
type OAuthStartResult struct {
	AuthorizationURL string
	CodeVerifier     string
	State            string
}

// TwitterGateway abstracts the X API v2.
type TwitterGateway interface {
	BuildAuthorizationURL(ctx context.Context) (*OAuthStartResult, error)
	ExchangeCode(ctx context.Context, code string, codeVerifier string) (*OAuthTokenResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*OAuthTokenResponse, error)
	GetUserInfo(ctx context.Context, accessToken string) (*TwitterUserInfo, error)
	SearchRecentTweets(ctx context.Context, accessToken string, input SearchTweetsInput) ([]Tweet, error)
	GetTweetCounts(ctx context.Context, accessToken string, query string, startTime time.Time) ([]TweetCountDataPoint, error)
	PostTweet(ctx context.Context, accessToken string, text string) (string, error)
	PostReply(ctx context.Context, accessToken string, text string, inReplyToTweetID string) (string, error)
	DeleteTweet(ctx context.Context, accessToken string, tweetID string) error
	VerifyCredentials(ctx context.Context, accessToken string) error
	SendDirectMessage(ctx context.Context, accessToken string, recipientID string, text string) error
	RevokeToken(ctx context.Context, accessToken string, refreshToken string) error
}
