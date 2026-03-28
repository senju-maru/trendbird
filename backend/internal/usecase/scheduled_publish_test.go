package usecase

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/domain/gateway"
)

// ---------------------------------------------------------------------------
// Mock implementations for ScheduledPublishUsecase tests
// ---------------------------------------------------------------------------

type mockPostRepo struct {
	ListScheduledFn func(ctx context.Context) ([]*entity.Post, error)
	UpdateFn        func(ctx context.Context, post *entity.Post) error

	mu           sync.Mutex
	UpdatedPosts []*entity.Post
}

func (m *mockPostRepo) FindByID(_ context.Context, _ string) (*entity.Post, error) {
	return nil, nil
}
func (m *mockPostRepo) ListByUserIDAndStatus(_ context.Context, _ string, _ entity.PostStatus, _, _ int) ([]*entity.Post, int64, error) {
	return nil, 0, nil
}
func (m *mockPostRepo) ListByUserIDAndStatuses(_ context.Context, _ string, _ []entity.PostStatus, _, _ int) ([]*entity.Post, int64, error) {
	return nil, 0, nil
}
func (m *mockPostRepo) ListPublishedByUserID(_ context.Context, _ string, _, _ int) ([]*entity.Post, int64, error) {
	return nil, 0, nil
}
func (m *mockPostRepo) ListScheduled(ctx context.Context) ([]*entity.Post, error) {
	if m.ListScheduledFn != nil {
		return m.ListScheduledFn(ctx)
	}
	return nil, nil
}
func (m *mockPostRepo) Create(_ context.Context, _ *entity.Post) error { return nil }
func (m *mockPostRepo) Update(ctx context.Context, post *entity.Post) error {
	m.mu.Lock()
	m.UpdatedPosts = append(m.UpdatedPosts, post)
	m.mu.Unlock()
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, post)
	}
	return nil
}
func (m *mockPostRepo) Delete(_ context.Context, _ string) error { return nil }
func (m *mockPostRepo) CountByUserIDAndStatus(_ context.Context, _ string, _ entity.PostStatus) (int64, error) {
	return 0, nil
}
func (m *mockPostRepo) CountPublishedByUserIDCurrentMonth(_ context.Context, _ string) (int32, error) {
	return 0, nil
}

type mockTweetConnRepo struct {
	FindByUserIDFn func(ctx context.Context, userID string) (*entity.TwitterConnection, error)
	UpsertFn       func(ctx context.Context, conn *entity.TwitterConnection) error

	mu              sync.Mutex
	UpdatedStatuses []entity.TwitterConnectionStatus
}

func (m *mockTweetConnRepo) FindByUserID(ctx context.Context, userID string) (*entity.TwitterConnection, error) {
	if m.FindByUserIDFn != nil {
		return m.FindByUserIDFn(ctx, userID)
	}
	return nil, apperror.NotFound("not found")
}
func (m *mockTweetConnRepo) Upsert(ctx context.Context, conn *entity.TwitterConnection) error {
	if m.UpsertFn != nil {
		return m.UpsertFn(ctx, conn)
	}
	return nil
}
func (m *mockTweetConnRepo) UpdateStatus(_ context.Context, _ string, status entity.TwitterConnectionStatus, _ *string) error {
	m.mu.Lock()
	m.UpdatedStatuses = append(m.UpdatedStatuses, status)
	m.mu.Unlock()
	return nil
}
func (m *mockTweetConnRepo) UpdateLastTestedAt(_ context.Context, _ string) error { return nil }
func (m *mockTweetConnRepo) DeleteByUserID(_ context.Context, _ string) error     { return nil }

type mockActivityRepo struct {
	mu         sync.Mutex
	Activities []*entity.Activity
}

func (m *mockActivityRepo) Create(_ context.Context, a *entity.Activity) error {
	m.mu.Lock()
	m.Activities = append(m.Activities, a)
	m.mu.Unlock()
	return nil
}
func (m *mockActivityRepo) ListByUserID(_ context.Context, _ string, _, _ int) ([]*entity.Activity, error) {
	return nil, nil
}

type mockScheduledTwitterGW struct {
	RefreshTokenFn func(ctx context.Context, refreshToken string) (*gateway.OAuthTokenResponse, error)
	PostTweetFn    func(ctx context.Context, accessToken string, text string) (string, error)
}

func (m *mockScheduledTwitterGW) BuildAuthorizationURL(_ context.Context) (*gateway.OAuthStartResult, error) {
	return nil, nil
}
func (m *mockScheduledTwitterGW) ExchangeCode(_ context.Context, _, _ string) (*gateway.OAuthTokenResponse, error) {
	return nil, nil
}
func (m *mockScheduledTwitterGW) RefreshToken(ctx context.Context, refreshToken string) (*gateway.OAuthTokenResponse, error) {
	if m.RefreshTokenFn != nil {
		return m.RefreshTokenFn(ctx, refreshToken)
	}
	return &gateway.OAuthTokenResponse{
		AccessToken:  "refreshed-token",
		RefreshToken: "refreshed-refresh",
		ExpiresIn:    7200,
	}, nil
}
func (m *mockScheduledTwitterGW) GetUserInfo(_ context.Context, _ string) (*gateway.TwitterUserInfo, error) {
	return nil, nil
}
func (m *mockScheduledTwitterGW) SearchRecentTweets(_ context.Context, _ string, _ gateway.SearchTweetsInput) ([]gateway.Tweet, error) {
	return nil, nil
}
func (m *mockScheduledTwitterGW) GetTweetCounts(_ context.Context, _ string, _ string, _ time.Time) ([]gateway.TweetCountDataPoint, error) {
	return nil, nil
}
func (m *mockScheduledTwitterGW) PostTweet(ctx context.Context, accessToken string, text string) (string, error) {
	if m.PostTweetFn != nil {
		return m.PostTweetFn(ctx, accessToken, text)
	}
	return "https://x.com/user/status/123", nil
}
func (m *mockScheduledTwitterGW) DeleteTweet(_ context.Context, _ string, _ string) error {
	return nil
}
func (m *mockScheduledTwitterGW) VerifyCredentials(_ context.Context, _ string) error { return nil }
func (m *mockScheduledTwitterGW) SendDirectMessage(_ context.Context, _, _, _ string) error {
	return nil
}
func (m *mockScheduledTwitterGW) RevokeToken(_ context.Context, _, _ string) error { return nil }

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestScheduledPublish_Execute(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		topicName := "AI News"
		postRepo := &mockPostRepo{
			ListScheduledFn: func(_ context.Context) ([]*entity.Post, error) {
				return []*entity.Post{
					{ID: "post-1", UserID: "user-1", Content: "Hello world", TopicName: &topicName, Status: entity.PostScheduled},
				}, nil
			},
		}
		tweetConnRepo := &mockTweetConnRepo{
			FindByUserIDFn: func(_ context.Context, _ string) (*entity.TwitterConnection, error) {
				now := time.Now()
				return &entity.TwitterConnection{
					UserID:         "user-1",
					AccessToken:    "valid-token",
					RefreshToken:   "refresh-token",
					TokenExpiresAt: now.Add(1 * time.Hour), // Not expired
					Status:         entity.TwitterConnected,
					ConnectedAt:    &now,
				}, nil
			},
		}
		activityRepo := &mockActivityRepo{}
		twitterGW := &mockScheduledTwitterGW{
			PostTweetFn: func(_ context.Context, _ string, _ string) (string, error) {
				return "https://x.com/user/status/456", nil
			},
		}

		uc := NewScheduledPublishUsecase(postRepo, tweetConnRepo, activityRepo, twitterGW)
		if err := uc.Execute(context.Background()); err != nil {
			t.Fatalf("Execute: %v", err)
		}

		// Post が Published ステータスに更新されていること
		postRepo.mu.Lock()
		updated := postRepo.UpdatedPosts
		postRepo.mu.Unlock()
		if len(updated) != 1 {
			t.Fatalf("expected 1 updated post, got %d", len(updated))
		}
		if updated[0].Status != entity.PostPublished {
			t.Errorf("status: want %d, got %d", entity.PostPublished, updated[0].Status)
		}
		if updated[0].TweetURL == nil || *updated[0].TweetURL != "https://x.com/user/status/456" {
			t.Errorf("tweet_url: want https://x.com/user/status/456, got %v", updated[0].TweetURL)
		}
		if updated[0].PublishedAt == nil {
			t.Error("published_at should be set")
		}

		// Activity が記録されていること
		activityRepo.mu.Lock()
		activities := activityRepo.Activities
		activityRepo.mu.Unlock()
		if len(activities) != 1 {
			t.Fatalf("expected 1 activity, got %d", len(activities))
		}
	})

	t.Run("no_scheduled_posts", func(t *testing.T) {
		postRepo := &mockPostRepo{
			ListScheduledFn: func(_ context.Context) ([]*entity.Post, error) {
				return nil, nil
			},
		}
		uc := NewScheduledPublishUsecase(postRepo, &mockTweetConnRepo{}, &mockActivityRepo{}, &mockScheduledTwitterGW{})
		if err := uc.Execute(context.Background()); err != nil {
			t.Fatalf("Execute: %v", err)
		}
	})

	t.Run("no_connection", func(t *testing.T) {
		postRepo := &mockPostRepo{
			ListScheduledFn: func(_ context.Context) ([]*entity.Post, error) {
				return []*entity.Post{
					{ID: "post-1", UserID: "user-1", Content: "Hello", Status: entity.PostScheduled},
				}, nil
			},
		}
		tweetConnRepo := &mockTweetConnRepo{
			FindByUserIDFn: func(_ context.Context, _ string) (*entity.TwitterConnection, error) {
				return nil, apperror.NotFound("twitter connection not found")
			},
		}

		uc := NewScheduledPublishUsecase(postRepo, tweetConnRepo, &mockActivityRepo{}, &mockScheduledTwitterGW{})
		err := uc.Execute(context.Background())

		// 全投稿失敗 → エラー返却
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		// Post が Failed ステータスに更新されていること
		postRepo.mu.Lock()
		updated := postRepo.UpdatedPosts
		postRepo.mu.Unlock()
		if len(updated) != 1 {
			t.Fatalf("expected 1 updated post, got %d", len(updated))
		}
		if updated[0].Status != entity.PostFailed {
			t.Errorf("status: want %d, got %d", entity.PostFailed, updated[0].Status)
		}
		if updated[0].ErrorMessage == nil {
			t.Error("error_message should be set")
		}
	})

	t.Run("token_refresh_success", func(t *testing.T) {
		postRepo := &mockPostRepo{
			ListScheduledFn: func(_ context.Context) ([]*entity.Post, error) {
				return []*entity.Post{
					{ID: "post-1", UserID: "user-1", Content: "Hello", Status: entity.PostScheduled},
				}, nil
			},
		}
		tweetConnRepo := &mockTweetConnRepo{
			FindByUserIDFn: func(_ context.Context, _ string) (*entity.TwitterConnection, error) {
				now := time.Now()
				return &entity.TwitterConnection{
					UserID:         "user-1",
					AccessToken:    "expired-token",
					RefreshToken:   "refresh-token",
					TokenExpiresAt: now.Add(-1 * time.Hour), // Expired
					Status:         entity.TwitterConnected,
					ConnectedAt:    &now,
				}, nil
			},
		}
		var usedToken string
		twitterGW := &mockScheduledTwitterGW{
			RefreshTokenFn: func(_ context.Context, _ string) (*gateway.OAuthTokenResponse, error) {
				return &gateway.OAuthTokenResponse{
					AccessToken:  "new-access-token",
					RefreshToken: "new-refresh-token",
					ExpiresIn:    7200,
				}, nil
			},
			PostTweetFn: func(_ context.Context, accessToken string, _ string) (string, error) {
				usedToken = accessToken
				return "https://x.com/user/status/789", nil
			},
		}

		uc := NewScheduledPublishUsecase(postRepo, tweetConnRepo, &mockActivityRepo{}, twitterGW)
		if err := uc.Execute(context.Background()); err != nil {
			t.Fatalf("Execute: %v", err)
		}

		// リフレッシュされたトークンで投稿していること
		if usedToken != "new-access-token" {
			t.Errorf("expected refreshed token, got %q", usedToken)
		}

		// Post が Published ステータスに更新されていること
		postRepo.mu.Lock()
		updated := postRepo.UpdatedPosts
		postRepo.mu.Unlock()
		if len(updated) != 1 {
			t.Fatalf("expected 1 updated post, got %d", len(updated))
		}
		if updated[0].Status != entity.PostPublished {
			t.Errorf("status: want %d, got %d", entity.PostPublished, updated[0].Status)
		}
	})

	t.Run("token_refresh_failure", func(t *testing.T) {
		postRepo := &mockPostRepo{
			ListScheduledFn: func(_ context.Context) ([]*entity.Post, error) {
				return []*entity.Post{
					{ID: "post-1", UserID: "user-1", Content: "Hello", Status: entity.PostScheduled},
				}, nil
			},
		}
		tweetConnRepo := &mockTweetConnRepo{
			FindByUserIDFn: func(_ context.Context, _ string) (*entity.TwitterConnection, error) {
				now := time.Now()
				return &entity.TwitterConnection{
					UserID:         "user-1",
					AccessToken:    "expired-token",
					RefreshToken:   "refresh-token",
					TokenExpiresAt: now.Add(-1 * time.Hour), // Expired
					Status:         entity.TwitterConnected,
					ConnectedAt:    &now,
				}, nil
			},
		}
		twitterGW := &mockScheduledTwitterGW{
			RefreshTokenFn: func(_ context.Context, _ string) (*gateway.OAuthTokenResponse, error) {
				return nil, fmt.Errorf("invalid_grant: refresh token revoked")
			},
		}

		uc := NewScheduledPublishUsecase(postRepo, tweetConnRepo, &mockActivityRepo{}, twitterGW)
		err := uc.Execute(context.Background())

		// 全投稿失敗 → エラー返却
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		// Post が Failed ステータスに更新されていること
		postRepo.mu.Lock()
		updated := postRepo.UpdatedPosts
		postRepo.mu.Unlock()
		if len(updated) != 1 {
			t.Fatalf("expected 1 updated post, got %d", len(updated))
		}
		if updated[0].Status != entity.PostFailed {
			t.Errorf("status: want %d, got %d", entity.PostFailed, updated[0].Status)
		}

		// TwitterConnection のステータスが Error に更新されていること
		tweetConnRepo.mu.Lock()
		statuses := tweetConnRepo.UpdatedStatuses
		tweetConnRepo.mu.Unlock()
		if len(statuses) != 1 {
			t.Fatalf("expected 1 status update, got %d", len(statuses))
		}
		if statuses[0] != entity.TwitterError {
			t.Errorf("connection status: want %d, got %d", entity.TwitterError, statuses[0])
		}
	})

	t.Run("x_api_failure", func(t *testing.T) {
		postRepo := &mockPostRepo{
			ListScheduledFn: func(_ context.Context) ([]*entity.Post, error) {
				return []*entity.Post{
					{ID: "post-1", UserID: "user-1", Content: "Hello", Status: entity.PostScheduled},
				}, nil
			},
		}
		tweetConnRepo := &mockTweetConnRepo{
			FindByUserIDFn: func(_ context.Context, _ string) (*entity.TwitterConnection, error) {
				now := time.Now()
				return &entity.TwitterConnection{
					UserID:         "user-1",
					AccessToken:    "valid-token",
					RefreshToken:   "refresh-token",
					TokenExpiresAt: now.Add(1 * time.Hour),
					Status:         entity.TwitterConnected,
					ConnectedAt:    &now,
				}, nil
			},
		}
		twitterGW := &mockScheduledTwitterGW{
			PostTweetFn: func(_ context.Context, _ string, _ string) (string, error) {
				return "", fmt.Errorf("x api: 403 forbidden")
			},
		}

		uc := NewScheduledPublishUsecase(postRepo, tweetConnRepo, &mockActivityRepo{}, twitterGW)
		err := uc.Execute(context.Background())

		// 全投稿失敗 → エラー返却
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		// Post が Failed ステータスに更新されていること
		postRepo.mu.Lock()
		updated := postRepo.UpdatedPosts
		postRepo.mu.Unlock()
		if len(updated) != 1 {
			t.Fatalf("expected 1 updated post, got %d", len(updated))
		}
		if updated[0].Status != entity.PostFailed {
			t.Errorf("status: want %d, got %d", entity.PostFailed, updated[0].Status)
		}
	})
}
