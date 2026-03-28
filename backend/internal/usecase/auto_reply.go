package usecase

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/domain/gateway"
	"github.com/trendbird/backend/internal/domain/repository"
)

// AutoReplyUsecase handles auto reply rule management (settings UI).
type AutoReplyUsecase struct {
	ruleRepo repository.AutoReplyRuleRepository
	logRepo  repository.ReplySentLogRepository
}

func NewAutoReplyUsecase(
	ruleRepo repository.AutoReplyRuleRepository,
	logRepo repository.ReplySentLogRepository,
) *AutoReplyUsecase {
	return &AutoReplyUsecase{
		ruleRepo: ruleRepo,
		logRepo:  logRepo,
	}
}

// ListRules returns all auto reply rules for the user.
func (u *AutoReplyUsecase) ListRules(ctx context.Context, userID string) ([]*entity.AutoReplyRule, error) {
	return u.ruleRepo.ListByUserID(ctx, userID)
}

// CreateRule creates a new auto reply rule for the user.
func (u *AutoReplyUsecase) CreateRule(ctx context.Context, userID string, targetTweetID string, targetTweetText string, keywords []string, replyTemplate string) (*entity.AutoReplyRule, error) {
	rule := &entity.AutoReplyRule{
		UserID:          userID,
		Enabled:         true,
		TargetTweetID:   targetTweetID,
		TargetTweetText: targetTweetText,
		TriggerKeywords: keywords,
		ReplyTemplate:   replyTemplate,
	}
	if err := u.ruleRepo.Create(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

// UpdateRule updates an existing auto reply rule (with ownership check).
func (u *AutoReplyUsecase) UpdateRule(ctx context.Context, userID string, ruleID string, enabled bool, keywords []string, replyTemplate string) (*entity.AutoReplyRule, error) {
	existing, err := u.ruleRepo.FindByID(ctx, ruleID)
	if err != nil {
		return nil, err
	}
	if existing.UserID != userID {
		return nil, apperror.PermissionDenied("not your rule")
	}
	existing.Enabled = enabled
	existing.TriggerKeywords = keywords
	existing.ReplyTemplate = replyTemplate
	if err := u.ruleRepo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return u.ruleRepo.FindByID(ctx, ruleID)
}

// DeleteRule deletes an auto reply rule (with ownership check).
func (u *AutoReplyUsecase) DeleteRule(ctx context.Context, userID string, ruleID string) error {
	existing, err := u.ruleRepo.FindByID(ctx, ruleID)
	if err != nil {
		return err
	}
	if existing.UserID != userID {
		return apperror.PermissionDenied("not your rule")
	}
	return u.ruleRepo.DeleteByID(ctx, ruleID)
}

// GetSentLogs returns recent reply sent logs for the user.
func (u *AutoReplyUsecase) GetSentLogs(ctx context.Context, userID string, limit int) ([]*entity.ReplySentLog, error) {
	if limit <= 0 {
		limit = 20
	}
	return u.logRepo.ListByUserID(ctx, userID, limit)
}

// AutoReplyBatchUsecase handles the batch job that polls for replies and sends auto-replies.
type AutoReplyBatchUsecase struct {
	ruleRepo    repository.AutoReplyRuleRepository
	logRepo     repository.ReplySentLogRepository
	pendingRepo repository.ReplyPendingQueueRepository
	connRepo    repository.TwitterConnectionRepository
	userRepo    repository.UserRepository
	twitterGW   gateway.TwitterGateway
}

func NewAutoReplyBatchUsecase(
	ruleRepo repository.AutoReplyRuleRepository,
	logRepo repository.ReplySentLogRepository,
	pendingRepo repository.ReplyPendingQueueRepository,
	connRepo repository.TwitterConnectionRepository,
	userRepo repository.UserRepository,
	twitterGW gateway.TwitterGateway,
) *AutoReplyBatchUsecase {
	return &AutoReplyBatchUsecase{
		ruleRepo:    ruleRepo,
		logRepo:     logRepo,
		pendingRepo: pendingRepo,
		connRepo:    connRepo,
		userRepo:    userRepo,
		twitterGW:   twitterGW,
	}
}

// maxReplyPerBatchPerUser limits reply sends per user per batch run (rate limit: 100/15min for tweets).
const maxReplyPerBatchPerUser = 50

// Execute runs the auto reply batch: poll replies → match keywords + conversation_id → enqueue → send replies.
func (u *AutoReplyBatchUsecase) Execute(ctx context.Context) error {
	rules, err := u.ruleRepo.ListEnabled(ctx)
	if err != nil {
		return fmt.Errorf("list enabled rules: %w", err)
	}
	if len(rules) == 0 {
		slog.Info("no enabled auto reply rules found")
		return nil
	}
	slog.Info("auto reply batch started", "enabled_rules", len(rules))

	// Group rules by user for optimized X API calls
	userRules := make(map[string][]*entity.AutoReplyRule)
	for _, rule := range rules {
		userRules[rule.UserID] = append(userRules[rule.UserID], rule)
	}

	var totalErrors int

	// Step 1: Poll replies and enqueue matching ones (per user)
	for userID, rules := range userRules {
		if err := u.processUserRules(ctx, userID, rules); err != nil {
			slog.Error("failed to process auto reply rules for user",
				"user_id", userID, "rule_count", len(rules), "error", err)
			totalErrors++
		}
	}

	// Step 2: Send pending replies
	if err := u.sendPendingReplies(ctx); err != nil {
		slog.Error("failed to send pending replies", "error", err)
		totalErrors++
	}

	if totalErrors > 0 {
		slog.Warn("auto reply batch completed with errors", "error_count", totalErrors)
	} else {
		slog.Info("auto reply batch completed successfully")
	}
	return nil
}

// processUserRules processes all rules for a single user with one X API call.
func (u *AutoReplyBatchUsecase) processUserRules(ctx context.Context, userID string, rules []*entity.AutoReplyRule) error {
	user, err := u.userRepo.FindByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}

	conn, err := u.connRepo.FindByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find twitter connection: %w", err)
	}
	if conn.Status != entity.TwitterConnected {
		slog.Warn("twitter not connected, skipping", "user_id", userID)
		return nil
	}

	// Use the minimum since_id across all rules for the search
	sinceID := autoReplyMinSinceID(rules)

	query := fmt.Sprintf("to:%s", user.TwitterHandle)
	startTime := time.Now().Add(-6 * time.Hour)
	input := gateway.SearchTweetsInput{
		Query:      query,
		MaxResults: 100,
		StartTime:  &startTime,
		SinceID:    sinceID,
	}

	tweets, err := u.twitterGW.SearchRecentTweets(ctx, conn.AccessToken, input)
	if err != nil {
		return fmt.Errorf("search recent tweets: %w", err)
	}
	slog.Info("fetched replies", "user_id", userID, "reply_count", len(tweets), "rule_count", len(rules))

	if len(tweets) == 0 {
		return nil
	}

	// Build target tweet ID set for fast lookup
	targetTweetIDs := make(map[string]*entity.AutoReplyRule, len(rules))
	for _, rule := range rules {
		targetTweetIDs[rule.TargetTweetID] = rule
	}

	// Track max tweet ID for since_id update
	var maxTweetID string
	for _, tweet := range tweets {
		if tweet.ID > maxTweetID {
			maxTweetID = tweet.ID
		}
	}

	// Match conversation_id + keywords and enqueue
	var enqueued int
	for _, tweet := range tweets {
		// Skip self-replies
		if tweet.AuthorID == user.TwitterID {
			continue
		}

		// Match conversation_id to target_tweet_id
		rule, ok := targetTweetIDs[tweet.ConversationID]
		if !ok {
			continue
		}

		// Match keywords
		keyword := matchAutoReplyKeyword(tweet.Text, rule.TriggerKeywords)
		if keyword == "" {
			continue
		}

		// Check if already sent
		exists, err := u.logRepo.ExistsByOriginalTweetID(ctx, tweet.ID, tweet.AuthorID)
		if err != nil {
			slog.Error("failed to check reply sent log", "tweet_id", tweet.ID, "error", err)
			continue
		}
		if exists {
			continue
		}

		item := &entity.ReplyPendingQueue{
			UserID:           userID,
			RuleID:           rule.ID,
			OriginalTweetID:  tweet.ID,
			OriginalAuthorID: tweet.AuthorID,
			TriggerKeyword:   keyword,
			Status:           entity.ReplyPendingStatusPending,
		}
		if err := u.pendingRepo.Create(ctx, item); err != nil {
			slog.Error("failed to enqueue reply", "tweet_id", tweet.ID, "error", err)
			continue
		}
		enqueued++
	}

	// Update since_id for all rules
	if maxTweetID != "" {
		for _, rule := range rules {
			if err := u.ruleRepo.UpdateLastCheckedReplyID(ctx, rule.ID, maxTweetID); err != nil {
				slog.Error("failed to update last checked reply id", "rule_id", rule.ID, "error", err)
			}
		}
	}

	slog.Info("enqueued replies", "user_id", userID, "enqueued", enqueued, "total_replies", len(tweets))
	return nil
}

func (u *AutoReplyBatchUsecase) sendPendingReplies(ctx context.Context) error {
	grouped, err := u.pendingRepo.ListPendingGroupedByUser(ctx)
	if err != nil {
		return fmt.Errorf("list pending replies: %w", err)
	}

	for userID, items := range grouped {
		conn, err := u.connRepo.FindByUserID(ctx, userID)
		if err != nil {
			slog.Error("failed to find twitter connection for reply send", "user_id", userID, "error", err)
			continue
		}
		if conn.Status != entity.TwitterConnected {
			continue
		}

		// Build a map of rule ID → rule for template lookup
		userRules, err := u.ruleRepo.ListByUserID(ctx, userID)
		if err != nil {
			slog.Error("failed to list rules for reply send", "user_id", userID, "error", err)
			continue
		}
		ruleMap := make(map[string]*entity.AutoReplyRule, len(userRules))
		for _, r := range userRules {
			ruleMap[r.ID] = r
		}

		sent := 0
		for _, item := range items {
			if sent >= maxReplyPerBatchPerUser {
				slog.Info("rate limit reached, remaining items stay pending",
					"user_id", userID, "remaining", len(items)-sent)
				break
			}

			rule, ok := ruleMap[item.RuleID]
			if !ok {
				slog.Warn("rule not found for pending reply, marking failed", "rule_id", item.RuleID)
				if err := u.pendingRepo.UpdateStatus(ctx, item.ID, entity.ReplyPendingStatusFailed); err != nil {
					slog.Error("failed to update pending status to failed", "id", item.ID, "error", err)
				}
				continue
			}

			replyText := rule.ReplyTemplate
			replyTweetID, err := u.twitterGW.PostReply(ctx, conn.AccessToken, replyText, item.OriginalTweetID)
			if err != nil {
				slog.Error("failed to send reply",
					"user_id", userID, "original_tweet_id", item.OriginalTweetID, "error", err)
				if err := u.pendingRepo.UpdateStatus(ctx, item.ID, entity.ReplyPendingStatusFailed); err != nil {
					slog.Error("failed to update pending status to failed", "id", item.ID, "error", err)
				}
				if apperror.IsCode(err, apperror.CodeResourceExhausted) {
					break
				}
				continue
			}

			logEntry := &entity.ReplySentLog{
				UserID:           userID,
				RuleID:           item.RuleID,
				OriginalTweetID:  item.OriginalTweetID,
				OriginalAuthorID: item.OriginalAuthorID,
				ReplyTweetID:     replyTweetID,
				TriggerKeyword:   item.TriggerKeyword,
				ReplyText:        replyText,
				SentAt:           time.Now(),
			}
			if err := u.logRepo.Create(ctx, logEntry); err != nil {
				slog.Error("failed to create reply sent log", "id", item.ID, "error", err)
			}

			if err := u.pendingRepo.DeleteByID(ctx, item.ID); err != nil {
				slog.Error("failed to delete pending reply", "id", item.ID, "error", err)
			}

			sent++
		}

		slog.Info("sent replies for user", "user_id", userID, "sent", sent, "total_pending", len(items))
	}

	return nil
}

// matchAutoReplyKeyword checks if the tweet text contains any of the trigger keywords (case-insensitive).
func matchAutoReplyKeyword(text string, keywords []string) string {
	lower := strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return kw
		}
	}
	return ""
}

// autoReplyMinSinceID returns the minimum last_checked_reply_id across all rules.
// Returns nil if any rule has no last_checked_reply_id (to search from startTime).
func autoReplyMinSinceID(rules []*entity.AutoReplyRule) *string {
	var min *string
	for _, rule := range rules {
		if rule.LastCheckedReplyID == nil {
			return nil
		}
		if min == nil || *rule.LastCheckedReplyID < *min {
			min = rule.LastCheckedReplyID
		}
	}
	return min
}
