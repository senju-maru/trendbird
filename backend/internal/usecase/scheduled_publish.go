package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/domain/gateway"
	"github.com/trendbird/backend/internal/domain/repository"
)

// ScheduledPublishUsecase handles batch publishing of scheduled posts.
type ScheduledPublishUsecase struct {
	postRepo      repository.PostRepository
	tweetConnRepo repository.TwitterConnectionRepository
	activityRepo  repository.ActivityRepository
	twitterGW     gateway.TwitterGateway
}

// NewScheduledPublishUsecase creates a new ScheduledPublishUsecase.
func NewScheduledPublishUsecase(
	postRepo repository.PostRepository,
	tweetConnRepo repository.TwitterConnectionRepository,
	activityRepo repository.ActivityRepository,
	twitterGW gateway.TwitterGateway,
) *ScheduledPublishUsecase {
	return &ScheduledPublishUsecase{
		postRepo:      postRepo,
		tweetConnRepo: tweetConnRepo,
		activityRepo:  activityRepo,
		twitterGW:     twitterGW,
	}
}

// Execute runs the scheduled publish batch job.
func (u *ScheduledPublishUsecase) Execute(ctx context.Context) error {
	posts, err := u.postRepo.ListScheduled(ctx)
	if err != nil {
		return fmt.Errorf("list scheduled posts: %w", err)
	}
	if len(posts) == 0 {
		slog.Info("no scheduled posts found")
		return nil
	}
	slog.Info("starting scheduled publish", "post_count", len(posts))

	var successCount, errorCount int

	for _, post := range posts {
		if err := u.publishOne(ctx, post); err != nil {
			if apperror.IsCode(err, apperror.CodeResourceExhausted) {
				slog.Warn("rate limit hit, stopping scheduled publish loop",
					"published", successCount, "remaining", len(posts)-successCount-errorCount)
				break
			}
			slog.Error("failed to publish scheduled post",
				"post_id", post.ID, "user_id", post.UserID, "error", err)
			errorCount++
			continue
		}
		successCount++
	}

	slog.Info("scheduled publish completed",
		"success", successCount, "errors", errorCount)

	if errorCount > 0 && successCount == 0 {
		return fmt.Errorf("all scheduled posts failed (%d errors)", errorCount)
	}
	return nil
}

// publishOne publishes a single scheduled post to X.
func (u *ScheduledPublishUsecase) publishOne(ctx context.Context, post *entity.Post) error {
	now := time.Now()

	// Get X connection for the user
	conn, err := u.tweetConnRepo.FindByUserID(ctx, post.UserID)
	if err != nil {
		if apperror.IsCode(err, apperror.CodeNotFound) {
			u.markFailed(ctx, post, now, "X連携が見つかりません")
			return fmt.Errorf("no twitter connection for user %s", post.UserID)
		}
		return err
	}

	accessToken := conn.AccessToken

	// Refresh token if expired
	if conn.TokenExpiresAt.Before(now) {
		tokenResp, err := u.twitterGW.RefreshToken(ctx, conn.RefreshToken)
		if err != nil {
			errMsg := "failed to refresh token: " + err.Error()
			_ = u.tweetConnRepo.UpdateStatus(ctx, post.UserID, entity.TwitterError, &errMsg)
			u.markFailed(ctx, post, now, errMsg)
			return fmt.Errorf("refresh token for user %s: %w", post.UserID, err)
		}

		updated := &entity.TwitterConnection{
			UserID:         post.UserID,
			AccessToken:    tokenResp.AccessToken,
			RefreshToken:   tokenResp.RefreshToken,
			TokenExpiresAt: now.Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
			Status:         conn.Status,
			ConnectedAt:    conn.ConnectedAt,
		}
		if err := u.tweetConnRepo.Upsert(ctx, updated); err != nil {
			return fmt.Errorf("update refreshed token for user %s: %w", post.UserID, err)
		}
		accessToken = tokenResp.AccessToken
	}

	// Post tweet
	tweetURL, err := u.twitterGW.PostTweet(ctx, accessToken, post.Content)
	if err != nil {
		// Propagate rate limit errors to break the loop
		if apperror.IsCode(err, apperror.CodeResourceExhausted) {
			return err
		}
		u.markFailed(ctx, post, now, err.Error())
		return fmt.Errorf("post tweet for post %s: %w", post.ID, err)
	}

	// Mark as published
	post.Status = entity.PostPublished
	post.PublishedAt = &now
	post.TweetURL = &tweetURL

	if err := u.postRepo.Update(ctx, post); err != nil {
		return fmt.Errorf("update post %s to published: %w", post.ID, err)
	}

	// Best-effort: record activity
	topicName := ""
	if post.TopicName != nil {
		topicName = *post.TopicName
	}
	RecordActivity(ctx, u.activityRepo, post.UserID, entity.ActivityPosted, topicName, "予約投稿をXに公開しました")

	slog.Info("published scheduled post", "post_id", post.ID, "user_id", post.UserID)
	return nil
}

// markFailed marks a post as failed with an error message.
func (u *ScheduledPublishUsecase) markFailed(ctx context.Context, post *entity.Post, now time.Time, errMsg string) {
	post.Status = entity.PostFailed
	post.FailedAt = &now
	post.ErrorMessage = &errMsg
	if updateErr := u.postRepo.Update(ctx, post); updateErr != nil {
		slog.Warn("failed to update post status to failed", "post_id", post.ID, "error", updateErr)
	}
}
