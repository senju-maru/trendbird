package e2etest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/gateway"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
	"github.com/trendbird/backend/internal/infrastructure/persistence/repository"
	"github.com/trendbird/backend/internal/usecase"
)

func TestScheduledPublishBatch_Success(t *testing.T) {
	env := setupTest(t)

	user := seedUser(t, env.db)
	topic := seedTopic(t, env.db)
	seedUserTopic(t, env.db, user.ID, topic.ID)
	seedTwitterConnection(t, env.db, user.ID)

	// 過去の scheduled_at を持つ投稿を seed
	pastTime := time.Now().Add(-1 * time.Hour)
	post := seedPost(t, env.db, user.ID,
		withPostStatus(2), // Scheduled
		withPostScheduledAt(pastTime),
		withPostContent("Scheduled tweet content"),
		withPostTopicName("Test Topic"),
	)

	// PostTweet のモックが URL を返すよう設定
	env.mockTwitter.PostTweetFn = func(_ context.Context, _ string, _ string) (string, error) {
		return "https://x.com/user/status/123456", nil
	}

	postRepo := repository.NewPostRepository(env.db)
	connRepo := repository.NewTwitterConnectionRepository(env.db)
	activityRepo := repository.NewActivityRepository(env.db)

	uc := usecase.NewScheduledPublishUsecase(postRepo, connRepo, activityRepo, env.mockTwitter)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// DB: status=3 (Published), tweet_url が設定されている
	var dbPost model.Post
	if err := env.db.First(&dbPost, "id = ?", post.ID).Error; err != nil {
		t.Fatalf("query post: %v", err)
	}
	if dbPost.Status != 3 { // Published
		t.Errorf("post.status: want 3 (Published), got %d", dbPost.Status)
	}
	if dbPost.TweetURL == nil || *dbPost.TweetURL != "https://x.com/user/status/123456" {
		t.Errorf("post.tweet_url: want https://x.com/user/status/123456, got %v", dbPost.TweetURL)
	}
	if dbPost.PublishedAt == nil {
		t.Error("post.published_at: want non-nil")
	}

	// Activity が記録されている
	var actCount int64
	env.db.Model(&model.Activity{}).Where("user_id = ? AND type = ?", user.ID, 4).Count(&actCount) // ActivityPosted=4
	if actCount == 0 {
		t.Error("expected activity record for published post")
	}
}

func TestScheduledPublishBatch_NoConnection(t *testing.T) {
	env := setupTest(t)

	user := seedUser(t, env.db)
	// twitter_connection を seed しない

	pastTime := time.Now().Add(-1 * time.Hour)
	post := seedPost(t, env.db, user.ID,
		withPostStatus(2), // Scheduled
		withPostScheduledAt(pastTime),
		withPostContent("No connection tweet"),
	)

	postRepo := repository.NewPostRepository(env.db)
	connRepo := repository.NewTwitterConnectionRepository(env.db)
	activityRepo := repository.NewActivityRepository(env.db)

	uc := usecase.NewScheduledPublishUsecase(postRepo, connRepo, activityRepo, env.mockTwitter)

	// 全投稿失敗 → エラー
	if err := uc.Execute(context.Background()); err == nil {
		t.Fatal("expected error when all posts fail")
	}

	// DB: status=4 (Failed), error_message が設定されている
	var dbPost model.Post
	if err := env.db.First(&dbPost, "id = ?", post.ID).Error; err != nil {
		t.Fatalf("query post: %v", err)
	}
	if dbPost.Status != 4 { // Failed
		t.Errorf("post.status: want 4 (Failed), got %d", dbPost.Status)
	}
	if dbPost.ErrorMessage == nil || *dbPost.ErrorMessage == "" {
		t.Error("post.error_message: want non-empty")
	}
}

func TestScheduledPublishBatch_TweetFailure(t *testing.T) {
	env := setupTest(t)

	user := seedUser(t, env.db)
	seedTwitterConnection(t, env.db, user.ID)

	pastTime := time.Now().Add(-1 * time.Hour)
	post := seedPost(t, env.db, user.ID,
		withPostStatus(2), // Scheduled
		withPostScheduledAt(pastTime),
		withPostContent("Will fail tweet"),
	)

	// PostTweet がエラーを返す
	env.mockTwitter.PostTweetFn = func(_ context.Context, _ string, _ string) (string, error) {
		return "", fmt.Errorf("X API internal error")
	}

	postRepo := repository.NewPostRepository(env.db)
	connRepo := repository.NewTwitterConnectionRepository(env.db)
	activityRepo := repository.NewActivityRepository(env.db)

	uc := usecase.NewScheduledPublishUsecase(postRepo, connRepo, activityRepo, env.mockTwitter)

	// 全投稿失敗 → エラー
	if err := uc.Execute(context.Background()); err == nil {
		t.Fatal("expected error when all posts fail")
	}

	// DB: status=4 (Failed), error_message に X API エラーが含まれる
	var dbPost model.Post
	if err := env.db.First(&dbPost, "id = ?", post.ID).Error; err != nil {
		t.Fatalf("query post: %v", err)
	}
	if dbPost.Status != 4 {
		t.Errorf("post.status: want 4 (Failed), got %d", dbPost.Status)
	}
	if dbPost.ErrorMessage == nil || *dbPost.ErrorMessage == "" {
		t.Error("post.error_message: want non-empty")
	}
}

func TestScheduledPublishBatch_TokenRefresh(t *testing.T) {
	env := setupTest(t)

	user := seedUser(t, env.db)
	// 期限切れトークンの twitter_connection を seed
	seedTwitterConnection(t, env.db, user.ID,
		withTokenExpiresAt(time.Now().Add(-1*time.Hour)), // 期限切れ
	)

	pastTime := time.Now().Add(-1 * time.Hour)
	seedPost(t, env.db, user.ID,
		withPostStatus(2),
		withPostScheduledAt(pastTime),
		withPostContent("Token refresh tweet"),
	)

	// RefreshToken → 成功
	env.mockTwitter.RefreshTokenFn = func(_ context.Context, _ string) (*gateway.OAuthTokenResponse, error) {
		return &gateway.OAuthTokenResponse{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			ExpiresIn:    7200,
		}, nil
	}
	env.mockTwitter.PostTweetFn = func(_ context.Context, accessToken string, _ string) (string, error) {
		// リフレッシュ後のトークンが使われていることを検証
		if accessToken != "new-access-token" {
			return "", fmt.Errorf("expected new-access-token, got %s", accessToken)
		}
		return "https://x.com/user/status/789", nil
	}

	postRepo := repository.NewPostRepository(env.db)
	connRepo := repository.NewTwitterConnectionRepository(env.db)
	activityRepo := repository.NewActivityRepository(env.db)

	uc := usecase.NewScheduledPublishUsecase(postRepo, connRepo, activityRepo, env.mockTwitter)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// DB: status=3 (Published)
	var count int64
	env.db.Model(&model.Post{}).Where("user_id = ? AND status = ?", user.ID, 3).Count(&count)
	if count != 1 {
		t.Errorf("published posts: want 1, got %d", count)
	}
}

func TestScheduledPublishBatch_RateLimit(t *testing.T) {
	env := setupTest(t)

	user := seedUser(t, env.db)
	seedTwitterConnection(t, env.db, user.ID)

	pastTime := time.Now().Add(-1 * time.Hour)
	// 2件の予約投稿を seed
	post1 := seedPost(t, env.db, user.ID, withPostStatus(2), withPostScheduledAt(pastTime), withPostContent("Tweet 1"))
	time.Sleep(10 * time.Millisecond) // 順序保証
	post2 := seedPost(t, env.db, user.ID, withPostStatus(2), withPostScheduledAt(pastTime), withPostContent("Tweet 2"))

	callCount := 0
	env.mockTwitter.PostTweetFn = func(_ context.Context, _ string, _ string) (string, error) {
		callCount++
		if callCount == 1 {
			return "https://x.com/user/status/1", nil
		}
		// 2件目でレートリミット
		return "", apperror.ResourceExhausted("rate limit exceeded")
	}

	postRepo := repository.NewPostRepository(env.db)
	connRepo := repository.NewTwitterConnectionRepository(env.db)
	activityRepo := repository.NewActivityRepository(env.db)

	uc := usecase.NewScheduledPublishUsecase(postRepo, connRepo, activityRepo, env.mockTwitter)

	// 部分成功 → エラーなし
	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// post1: Published
	var dbPost1 model.Post
	env.db.First(&dbPost1, "id = ?", post1.ID)
	if dbPost1.Status != 3 {
		t.Errorf("post1.status: want 3 (Published), got %d", dbPost1.Status)
	}

	// post2: まだ Scheduled（レートリミットでスキップされた）
	var dbPost2 model.Post
	env.db.First(&dbPost2, "id = ?", post2.ID)
	if dbPost2.Status != 2 {
		t.Errorf("post2.status: want 2 (Scheduled, skipped by rate limit), got %d", dbPost2.Status)
	}
}

func TestScheduledPublishBatch_Empty(t *testing.T) {
	env := setupTest(t)

	// 予約投稿なし
	postRepo := repository.NewPostRepository(env.db)
	connRepo := repository.NewTwitterConnectionRepository(env.db)
	activityRepo := repository.NewActivityRepository(env.db)

	uc := usecase.NewScheduledPublishUsecase(postRepo, connRepo, activityRepo, env.mockTwitter)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute with empty: %v", err)
	}
}

func TestScheduledPublishBatch_Idempotent(t *testing.T) {
	env := setupTest(t)

	user := seedUser(t, env.db)
	seedTwitterConnection(t, env.db, user.ID)

	pastTime := time.Now().Add(-1 * time.Hour)
	seedPost(t, env.db, user.ID,
		withPostStatus(2),
		withPostScheduledAt(pastTime),
		withPostContent("Idempotent tweet"),
	)

	env.mockTwitter.PostTweetFn = func(_ context.Context, _ string, _ string) (string, error) {
		return "https://x.com/user/status/idem", nil
	}

	postRepo := repository.NewPostRepository(env.db)
	connRepo := repository.NewTwitterConnectionRepository(env.db)
	activityRepo := repository.NewActivityRepository(env.db)

	uc := usecase.NewScheduledPublishUsecase(postRepo, connRepo, activityRepo, env.mockTwitter)

	// 1回目
	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("1st Execute: %v", err)
	}

	// 2回目: 既に Published なので PostTweet は呼ばれない
	beforeCalls := env.mockTwitter.Calls.Load()
	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("2nd Execute: %v", err)
	}
	afterCalls := env.mockTwitter.Calls.Load()

	if afterCalls != beforeCalls {
		t.Errorf("expected no additional Twitter API calls on 2nd run, got %d more", afterCalls-beforeCalls)
	}
}
