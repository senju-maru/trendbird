package e2etest

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/trendbird/backend/internal/domain/gateway"
)

// ---------------------------------------------------------------------------
// mockTwitterGateway
// ---------------------------------------------------------------------------

var _ gateway.TwitterGateway = (*mockTwitterGateway)(nil)

type mockTwitterGateway struct {
	BuildAuthorizationURLFn func(ctx context.Context) (*gateway.OAuthStartResult, error)
	ExchangeCodeFn          func(ctx context.Context, code string, codeVerifier string) (*gateway.OAuthTokenResponse, error)
	RefreshTokenFn          func(ctx context.Context, refreshToken string) (*gateway.OAuthTokenResponse, error)
	GetUserInfoFn           func(ctx context.Context, accessToken string) (*gateway.TwitterUserInfo, error)
	SearchRecentTweetsFn    func(ctx context.Context, accessToken string, input gateway.SearchTweetsInput) ([]gateway.Tweet, error)
	GetTweetCountsFn        func(ctx context.Context, accessToken string, query string, startTime time.Time) ([]gateway.TweetCountDataPoint, error)
	PostTweetFn             func(ctx context.Context, accessToken string, text string) (string, error)
	PostReplyFn             func(ctx context.Context, accessToken string, text string, inReplyToTweetID string) (string, error)
	DeleteTweetFn           func(ctx context.Context, accessToken string, tweetID string) error
	VerifyCredentialsFn     func(ctx context.Context, accessToken string) error
	SendDirectMessageFn     func(ctx context.Context, accessToken string, recipientID string, text string) error
	RevokeTokenFn           func(ctx context.Context, accessToken string, refreshToken string) error

	Calls atomic.Int64
}

func newMockTwitterGateway() *mockTwitterGateway {
	return &mockTwitterGateway{
		BuildAuthorizationURLFn: func(_ context.Context) (*gateway.OAuthStartResult, error) {
			return &gateway.OAuthStartResult{
				AuthorizationURL: "https://x.com/authorize?test=1",
				CodeVerifier:     "test-code-verifier",
				State:            "test-state",
			}, nil
		},
		ExchangeCodeFn: func(_ context.Context, _ string, _ string) (*gateway.OAuthTokenResponse, error) {
			return &gateway.OAuthTokenResponse{
				AccessToken:  "test-access-token",
				RefreshToken: "test-refresh-token",
				ExpiresIn:    7200,
			}, nil
		},
		RefreshTokenFn: func(_ context.Context, _ string) (*gateway.OAuthTokenResponse, error) {
			return &gateway.OAuthTokenResponse{
				AccessToken:  "refreshed-access-token",
				RefreshToken: "refreshed-refresh-token",
				ExpiresIn:    7200,
			}, nil
		},
		GetUserInfoFn: func(_ context.Context, _ string) (*gateway.TwitterUserInfo, error) {
			return &gateway.TwitterUserInfo{
				ID:       "x-user-123",
				Name:     "Test User",
				Username: "testuser",
				Email:    "test@example.com",
				Image:    "https://pbs.twimg.com/test.jpg",
			}, nil
		},
		SearchRecentTweetsFn: func(_ context.Context, _ string, _ gateway.SearchTweetsInput) ([]gateway.Tweet, error) {
			return []gateway.Tweet{}, nil
		},
		GetTweetCountsFn: func(_ context.Context, _ string, _ string, _ time.Time) ([]gateway.TweetCountDataPoint, error) {
			return []gateway.TweetCountDataPoint{}, nil
		},
		PostTweetFn: func(_ context.Context, _ string, _ string) (string, error) {
			return "tweet-id-123", nil
		},
		PostReplyFn: func(_ context.Context, _ string, _ string, _ string) (string, error) {
			return "reply-tweet-id-123", nil
		},
		DeleteTweetFn: func(_ context.Context, _ string, _ string) error {
			return nil
		},
		VerifyCredentialsFn: func(_ context.Context, _ string) error {
			return nil
		},
		SendDirectMessageFn: func(_ context.Context, _ string, _ string, _ string) error {
			return nil
		},
		RevokeTokenFn: func(_ context.Context, _ string, _ string) error {
			return nil
		},
	}
}

func (m *mockTwitterGateway) BuildAuthorizationURL(ctx context.Context) (*gateway.OAuthStartResult, error) {
	m.Calls.Add(1)
	return m.BuildAuthorizationURLFn(ctx)
}

func (m *mockTwitterGateway) ExchangeCode(ctx context.Context, code string, codeVerifier string) (*gateway.OAuthTokenResponse, error) {
	m.Calls.Add(1)
	return m.ExchangeCodeFn(ctx, code, codeVerifier)
}

func (m *mockTwitterGateway) RefreshToken(ctx context.Context, refreshToken string) (*gateway.OAuthTokenResponse, error) {
	m.Calls.Add(1)
	return m.RefreshTokenFn(ctx, refreshToken)
}

func (m *mockTwitterGateway) GetUserInfo(ctx context.Context, accessToken string) (*gateway.TwitterUserInfo, error) {
	m.Calls.Add(1)
	return m.GetUserInfoFn(ctx, accessToken)
}

func (m *mockTwitterGateway) SearchRecentTweets(ctx context.Context, accessToken string, input gateway.SearchTweetsInput) ([]gateway.Tweet, error) {
	m.Calls.Add(1)
	return m.SearchRecentTweetsFn(ctx, accessToken, input)
}

func (m *mockTwitterGateway) GetTweetCounts(ctx context.Context, accessToken string, query string, startTime time.Time) ([]gateway.TweetCountDataPoint, error) {
	m.Calls.Add(1)
	return m.GetTweetCountsFn(ctx, accessToken, query, startTime)
}

func (m *mockTwitterGateway) PostTweet(ctx context.Context, accessToken string, text string) (string, error) {
	m.Calls.Add(1)
	return m.PostTweetFn(ctx, accessToken, text)
}

func (m *mockTwitterGateway) PostReply(ctx context.Context, accessToken string, text string, inReplyToTweetID string) (string, error) {
	m.Calls.Add(1)
	return m.PostReplyFn(ctx, accessToken, text, inReplyToTweetID)
}

func (m *mockTwitterGateway) DeleteTweet(ctx context.Context, accessToken string, tweetID string) error {
	m.Calls.Add(1)
	return m.DeleteTweetFn(ctx, accessToken, tweetID)
}

func (m *mockTwitterGateway) VerifyCredentials(ctx context.Context, accessToken string) error {
	m.Calls.Add(1)
	return m.VerifyCredentialsFn(ctx, accessToken)
}

func (m *mockTwitterGateway) SendDirectMessage(ctx context.Context, accessToken string, recipientID string, text string) error {
	m.Calls.Add(1)
	return m.SendDirectMessageFn(ctx, accessToken, recipientID, text)
}

func (m *mockTwitterGateway) RevokeToken(ctx context.Context, accessToken string, refreshToken string) error {
	m.Calls.Add(1)
	return m.RevokeTokenFn(ctx, accessToken, refreshToken)
}

// ---------------------------------------------------------------------------
// mockAIGateway
// ---------------------------------------------------------------------------

var _ gateway.AIGateway = (*mockAIGateway)(nil)

type mockAIGateway struct {
	GeneratePostsFn  func(ctx context.Context, input gateway.GeneratePostsInput) (*gateway.GeneratePostsOutput, error)
	ResearchTopicFn  func(ctx context.Context, input gateway.ResearchTopicInput) (*gateway.ResearchTopicOutput, error)

	Calls atomic.Int64
}

func newMockAIGateway() *mockAIGateway {
	return &mockAIGateway{
		GeneratePostsFn: func(_ context.Context, input gateway.GeneratePostsInput) (*gateway.GeneratePostsOutput, error) {
			posts := make([]string, input.Count)
			for i := range posts {
				posts[i] = "AI generated post about " + input.TopicName
			}
			return &gateway.GeneratePostsOutput{Posts: posts}, nil
		},
		ResearchTopicFn: func(_ context.Context, input gateway.ResearchTopicInput) (*gateway.ResearchTopicOutput, error) {
			return &gateway.ResearchTopicOutput{
				Summary:    "Mock research summary for " + input.TopicName,
				SourceURLs: []string{"https://example.com/article1"},
			}, nil
		},
	}
}

func (m *mockAIGateway) GeneratePosts(ctx context.Context, input gateway.GeneratePostsInput) (*gateway.GeneratePostsOutput, error) {
	m.Calls.Add(1)
	return m.GeneratePostsFn(ctx, input)
}

func (m *mockAIGateway) ResearchTopic(ctx context.Context, input gateway.ResearchTopicInput) (*gateway.ResearchTopicOutput, error) {
	m.Calls.Add(1)
	return m.ResearchTopicFn(ctx, input)
}

