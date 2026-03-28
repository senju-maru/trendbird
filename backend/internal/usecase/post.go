package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/domain/gateway"
	"github.com/trendbird/backend/internal/domain/repository"
)

// PostStatsResult holds aggregated post statistics for a user.
type PostStatsResult struct {
	TotalPublished     int32
	TotalScheduled     int32
	TotalDrafts        int32
	TotalFailed        int32
	ThisMonthPublished int32
}

// PostUsecase handles AI post generation, draft management, publishing, and post history.
type PostUsecase struct {
	topicRepo         repository.TopicRepository
	postRepo          repository.PostRepository
	genPostRepo       repository.GeneratedPostRepository
	aiGenLogRepo      repository.AIGenerationLogRepository
	activityRepo      repository.ActivityRepository
	tweetConnRepo     repository.TwitterConnectionRepository
	topicResearchRepo repository.TopicResearchRepository
	aiGW              gateway.AIGateway
	twitterGW         gateway.TwitterGateway
	txManager         repository.TransactionManager
}

// NewPostUsecase creates a new PostUsecase.
func NewPostUsecase(
	topicRepo repository.TopicRepository,
	postRepo repository.PostRepository,
	genPostRepo repository.GeneratedPostRepository,
	aiGenLogRepo repository.AIGenerationLogRepository,
	activityRepo repository.ActivityRepository,
	tweetConnRepo repository.TwitterConnectionRepository,
	topicResearchRepo repository.TopicResearchRepository,
	aiGW gateway.AIGateway,
	twitterGW gateway.TwitterGateway,
	txManager repository.TransactionManager,
) *PostUsecase {
	return &PostUsecase{
		topicRepo:         topicRepo,
		postRepo:          postRepo,
		genPostRepo:       genPostRepo,
		aiGenLogRepo:      aiGenLogRepo,
		activityRepo:      activityRepo,
		tweetConnRepo:     tweetConnRepo,
		topicResearchRepo: topicResearchRepo,
		aiGW:              aiGW,
		twitterGW:         twitterGW,
		txManager:         txManager,
	}
}

// GeneratePosts generates AI post content for the given topic.
// Flow: topic fetch → AI generate → log → link
func (u *PostUsecase) GeneratePosts(ctx context.Context, userID, topicID string, style *entity.PostStyle) ([]*entity.GeneratedPost, error) {
	// Style customization check
	resolvedStyle := entity.PostStyleCasual
	if style != nil {
		resolvedStyle = *style
	}

	topic, err := u.topicRepo.FindByIDForUser(ctx, topicID, userID)
	if err != nil {
		return nil, err
	}

	// Build AI generation input
	topicContext := gateway.TopicContext{}
	if topic.ContextSummary != nil {
		topicContext.Summary = *topic.ContextSummary
	}
	topicContext.TrendKeywords = topic.Keywords

	// Fetch latest topic research results (best-effort: fallback to keywords only)
	if u.topicResearchRepo != nil {
		researches, err := u.topicResearchRepo.ListByTopicID(ctx, topicID, 5)
		if err != nil {
			slog.WarnContext(ctx, "failed to fetch topic research, continuing without", "topic_id", topicID, "error", err)
		} else {
			for _, r := range researches {
				topicContext.ResearchResults = append(topicContext.ResearchResults, r.Summary)
			}
		}
	}

	input := gateway.GeneratePostsInput{
		TopicName:    topic.Name,
		TopicContext: topicContext,
		Style:        resolvedStyle,
		Count:        3,
	}

	slog.InfoContext(ctx, "ai post generation started",
		"topic_id", topicID,
		"topic_name", topic.Name,
		"style", resolvedStyle,
	)
	output, err := u.aiGW.GeneratePosts(ctx, input)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to generate posts", err)
	}

	// Create AI generation log and generated posts atomically to prevent orphaned logs.
	genLog := &entity.AIGenerationLog{
		UserID:  userID,
		TopicID: &topicID,
		Style:   resolvedStyle,
		Count:   int32(len(output.Posts)),
	}
	generatedPosts := make([]*entity.GeneratedPost, len(output.Posts))
	for i, content := range output.Posts {
		generatedPosts[i] = &entity.GeneratedPost{
			UserID:  userID,
			TopicID: &topicID,
			Style:   resolvedStyle,
			Content: content,
		}
	}

	if err := u.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		if err := u.aiGenLogRepo.Create(txCtx, genLog); err != nil {
			return err
		}
		for _, gp := range generatedPosts {
			gp.GenerationLogID = &genLog.ID
		}
		return u.genPostRepo.BulkCreate(txCtx, generatedPosts)
	}); err != nil {
		return nil, err
	}

	// Best-effort: record activity
	RecordActivity(ctx, u.activityRepo, userID, entity.ActivityAIGenerated, topic.Name, "「"+topic.Name+"」のAI投稿文を生成しました")

	slog.InfoContext(ctx, "ai post generation completed",
		"topic_id", topicID,
		"topic_name", topic.Name,
		"generated_count", len(generatedPosts),
	)
	return generatedPosts, nil
}

// getValidAccessToken retrieves a valid X access token, refreshing if expired.
func (u *PostUsecase) getValidAccessToken(ctx context.Context, userID string) (string, error) {
	conn, err := u.tweetConnRepo.FindByUserID(ctx, userID)
	if err != nil {
		return "", err
	}

	if conn.TokenExpiresAt.After(time.Now()) {
		return conn.AccessToken, nil
	}

	tokenResp, err := u.twitterGW.RefreshToken(ctx, conn.RefreshToken)
	if err != nil {
		errMsg := "failed to refresh token: " + err.Error()
		_ = u.tweetConnRepo.UpdateStatus(ctx, userID, entity.TwitterError, &errMsg)
		return "", apperror.Wrap(apperror.CodeInternal, "failed to refresh X token", err)
	}

	now := time.Now()
	updated := &entity.TwitterConnection{
		UserID:         userID,
		AccessToken:    tokenResp.AccessToken,
		RefreshToken:   tokenResp.RefreshToken,
		TokenExpiresAt: now.Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
		Status:         conn.Status,
		ConnectedAt:    conn.ConnectedAt,
	}
	if err := u.tweetConnRepo.Upsert(ctx, updated); err != nil {
		return "", apperror.Wrap(apperror.CodeInternal, "failed to update refreshed token", err)
	}
	return tokenResp.AccessToken, nil
}

// ListDrafts returns the user's draft posts with statistics.
func (u *PostUsecase) ListDrafts(ctx context.Context, userID string, limit, offset int) ([]*entity.Post, *PostStatsResult, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	statuses := []entity.PostStatus{entity.PostDraft, entity.PostScheduled, entity.PostFailed}
	drafts, total, err := u.postRepo.ListByUserIDAndStatuses(ctx, userID, statuses, limit, offset)
	if err != nil {
		return nil, nil, 0, err
	}

	stats, err := u.GetPostStats(ctx, userID)
	if err != nil {
		return nil, nil, 0, err
	}

	return drafts, stats, total, nil
}

// CreateDraft creates a new draft post.
func (u *PostUsecase) CreateDraft(ctx context.Context, userID, content string, topicID *string) (*entity.Post, error) {
	if content == "" {
		return nil, apperror.InvalidArgument("content is required")
	}

	var topicName *string
	if topicID != nil {
		topic, err := u.topicRepo.FindByIDForUser(ctx, *topicID, userID)
		if err != nil {
			return nil, err
		}
		topicName = &topic.Name
	}

	post := &entity.Post{
		UserID:    userID,
		Content:   content,
		TopicID:   topicID,
		TopicName: topicName,
		Status:    entity.PostDraft,
	}

	if err := u.postRepo.Create(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

// UpdateDraft updates the content of a draft post.
func (u *PostUsecase) UpdateDraft(ctx context.Context, userID, postID, content string) (*entity.Post, error) {
	post, err := u.postRepo.FindByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if post.UserID != userID {
		return nil, apperror.PermissionDenied("post does not belong to user")
	}
	if post.Status != entity.PostDraft && post.Status != entity.PostScheduled {
		return nil, apperror.InvalidArgument("only drafts or scheduled posts can be edited")
	}

	post.Content = content
	if err := u.postRepo.Update(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

// DeleteDraft deletes a draft post.
func (u *PostUsecase) DeleteDraft(ctx context.Context, userID, postID string) error {
	post, err := u.postRepo.FindByID(ctx, postID)
	if err != nil {
		return err
	}
	if post.UserID != userID {
		return apperror.PermissionDenied("post does not belong to user")
	}
	if post.Status != entity.PostDraft && post.Status != entity.PostScheduled {
		return apperror.InvalidArgument("only drafts or scheduled posts can be deleted")
	}

	return u.postRepo.Delete(ctx, postID)
}

// SchedulePost schedules a draft post for future publishing.
func (u *PostUsecase) SchedulePost(ctx context.Context, userID, postID string, scheduledAt time.Time) (*entity.Post, error) {
	post, err := u.postRepo.FindByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if post.UserID != userID {
		return nil, apperror.PermissionDenied("post does not belong to user")
	}
	if post.Status != entity.PostDraft && post.Status != entity.PostScheduled {
		return nil, apperror.InvalidArgument("only drafts or scheduled posts can be scheduled")
	}
	if scheduledAt.Minute() != 0 || scheduledAt.Second() != 0 {
		return nil, apperror.InvalidArgument("scheduled time must be on the hour")
	}
	// Must be at least the next full hour
	nextHour := time.Now().Truncate(time.Hour).Add(time.Hour)
	if scheduledAt.Before(nextHour) {
		return nil, apperror.InvalidArgument("scheduled time must be in the future")
	}

	post.Status = entity.PostScheduled
	post.ScheduledAt = &scheduledAt

	if err := u.postRepo.Update(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

// PublishPost publishes a draft or scheduled post to X.
func (u *PostUsecase) PublishPost(ctx context.Context, userID, postID string) (*entity.Post, error) {
	post, err := u.postRepo.FindByID(ctx, postID)
	if err != nil {
		return nil, err
	}
	if post.UserID != userID {
		return nil, apperror.PermissionDenied("post does not belong to user")
	}
	if post.Status != entity.PostDraft && post.Status != entity.PostScheduled {
		return nil, apperror.InvalidArgument("only drafts or scheduled posts can be published")
	}

	conn, err := u.tweetConnRepo.FindByUserID(ctx, userID)
	if err != nil {
		if apperror.IsCode(err, apperror.CodeNotFound) {
			return nil, apperror.PermissionDenied("X連携が必要です。設定画面でXアカウントを連携してください")
		}
		return nil, err
	}

	accessToken := conn.AccessToken

	// Refresh token if expired
	if conn.TokenExpiresAt.Before(time.Now()) {
		tokenResp, err := u.twitterGW.RefreshToken(ctx, conn.RefreshToken)
		if err != nil {
			errMsg := "failed to refresh token: " + err.Error()
			_ = u.tweetConnRepo.UpdateStatus(ctx, userID, entity.TwitterError, &errMsg)
			return nil, apperror.Wrap(apperror.CodeInternal, "failed to refresh X token", err)
		}

		now := time.Now()
		updated := &entity.TwitterConnection{
			UserID:         userID,
			AccessToken:    tokenResp.AccessToken,
			RefreshToken:   tokenResp.RefreshToken,
			TokenExpiresAt: now.Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
			Status:         conn.Status,
			ConnectedAt:    conn.ConnectedAt,
		}
		if err := u.tweetConnRepo.Upsert(ctx, updated); err != nil {
			return nil, apperror.Wrap(apperror.CodeInternal, "failed to update refreshed token", err)
		}
		accessToken = tokenResp.AccessToken
	}

	// Attempt to post tweet
	now := time.Now()
	tweetURL, err := u.twitterGW.PostTweet(ctx, accessToken, post.Content)
	if err != nil {
		// Mark as failed
		post.Status = entity.PostFailed
		post.FailedAt = &now
		errMsg := err.Error()
		post.ErrorMessage = &errMsg
		if updateErr := u.postRepo.Update(ctx, post); updateErr != nil {
			slog.Warn("failed to update post status to failed", "postID", postID, "error", updateErr)
		}
		return post, apperror.Wrap(apperror.CodeInternal, "failed to post tweet", err)
	}

	// Mark as published
	post.Status = entity.PostPublished
	post.PublishedAt = &now
	post.TweetURL = &tweetURL

	if err := u.postRepo.Update(ctx, post); err != nil {
		return nil, err
	}

	// Best-effort: record activity
	topicName := ""
	if post.TopicName != nil {
		topicName = *post.TopicName
	}
	RecordActivity(ctx, u.activityRepo, userID, entity.ActivityPosted, topicName, "投稿をXに公開しました")

	return post, nil
}

// ListPostHistory returns the user's published posts with pagination.
func (u *PostUsecase) ListPostHistory(ctx context.Context, userID string, limit, offset int) ([]*entity.Post, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return u.postRepo.ListPublishedByUserID(ctx, userID, limit, offset)
}

// GetPostStats returns aggregated post statistics for the user.
func (u *PostUsecase) GetPostStats(ctx context.Context, userID string) (*PostStatsResult, error) {
	totalDrafts, err := u.postRepo.CountByUserIDAndStatus(ctx, userID, entity.PostDraft)
	if err != nil {
		return nil, err
	}

	totalScheduled, err := u.postRepo.CountByUserIDAndStatus(ctx, userID, entity.PostScheduled)
	if err != nil {
		return nil, err
	}

	totalPublished, err := u.postRepo.CountByUserIDAndStatus(ctx, userID, entity.PostPublished)
	if err != nil {
		return nil, err
	}

	totalFailed, err := u.postRepo.CountByUserIDAndStatus(ctx, userID, entity.PostFailed)
	if err != nil {
		return nil, err
	}

	thisMonthPublished, err := u.postRepo.CountPublishedByUserIDCurrentMonth(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &PostStatsResult{
		TotalPublished:     int32(totalPublished),
		TotalScheduled:     int32(totalScheduled),
		TotalDrafts:        int32(totalDrafts),
		TotalFailed:        int32(totalFailed),
		ThisMonthPublished: thisMonthPublished,
	}, nil
}

