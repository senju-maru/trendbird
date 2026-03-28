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

// AutoDMUsecase handles auto DM rule management (settings UI).
type AutoDMUsecase struct {
	ruleRepo repository.AutoDMRuleRepository
	logRepo  repository.DMSentLogRepository
}

func NewAutoDMUsecase(
	ruleRepo repository.AutoDMRuleRepository,
	logRepo repository.DMSentLogRepository,
) *AutoDMUsecase {
	return &AutoDMUsecase{
		ruleRepo: ruleRepo,
		logRepo:  logRepo,
	}
}

// ListRules returns all auto DM rules for the user.
func (u *AutoDMUsecase) ListRules(ctx context.Context, userID string) ([]*entity.AutoDMRule, error) {
	return u.ruleRepo.ListByUserID(ctx, userID)
}

// CreateRule creates a new auto DM rule for the user.
func (u *AutoDMUsecase) CreateRule(ctx context.Context, userID string, keywords []string, template string) (*entity.AutoDMRule, error) {
	rule := &entity.AutoDMRule{
		UserID:          userID,
		Enabled:         true,
		TriggerKeywords: keywords,
		TemplateMessage: template,
	}
	if err := u.ruleRepo.Create(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

// UpdateRule updates an existing auto DM rule (with ownership check).
func (u *AutoDMUsecase) UpdateRule(ctx context.Context, userID string, ruleID string, enabled bool, keywords []string, template string) (*entity.AutoDMRule, error) {
	existing, err := u.ruleRepo.FindByID(ctx, ruleID)
	if err != nil {
		return nil, err
	}
	if existing.UserID != userID {
		return nil, apperror.PermissionDenied("not your rule")
	}
	existing.Enabled = enabled
	existing.TriggerKeywords = keywords
	existing.TemplateMessage = template
	if err := u.ruleRepo.Update(ctx, existing); err != nil {
		return nil, err
	}
	return u.ruleRepo.FindByID(ctx, ruleID)
}

// DeleteRule deletes an auto DM rule (with ownership check).
func (u *AutoDMUsecase) DeleteRule(ctx context.Context, userID string, ruleID string) error {
	existing, err := u.ruleRepo.FindByID(ctx, ruleID)
	if err != nil {
		return err
	}
	if existing.UserID != userID {
		return apperror.PermissionDenied("not your rule")
	}
	return u.ruleRepo.DeleteByID(ctx, ruleID)
}

// GetSentLogs returns recent DM sent logs for the user.
func (u *AutoDMUsecase) GetSentLogs(ctx context.Context, userID string, limit int) ([]*entity.DMSentLog, error) {
	if limit <= 0 {
		limit = 20
	}
	return u.logRepo.ListByUserID(ctx, userID, limit)
}

// AutoDMBatchUsecase handles the batch job that polls for replies and sends DMs.
type AutoDMBatchUsecase struct {
	ruleRepo    repository.AutoDMRuleRepository
	logRepo     repository.DMSentLogRepository
	pendingRepo repository.DMPendingQueueRepository
	connRepo    repository.TwitterConnectionRepository
	userRepo    repository.UserRepository
	twitterGW   gateway.TwitterGateway
}

func NewAutoDMBatchUsecase(
	ruleRepo repository.AutoDMRuleRepository,
	logRepo repository.DMSentLogRepository,
	pendingRepo repository.DMPendingQueueRepository,
	connRepo repository.TwitterConnectionRepository,
	userRepo repository.UserRepository,
	twitterGW gateway.TwitterGateway,
) *AutoDMBatchUsecase {
	return &AutoDMBatchUsecase{
		ruleRepo:    ruleRepo,
		logRepo:     logRepo,
		pendingRepo: pendingRepo,
		connRepo:    connRepo,
		userRepo:    userRepo,
		twitterGW:   twitterGW,
	}
}

// maxDMPerBatchPerUser limits DM sends per user per batch run (rate limit: 15/15min).
const maxDMPerBatchPerUser = 15

// Execute runs the auto DM batch: poll replies → match keywords → enqueue → send DMs.
func (u *AutoDMBatchUsecase) Execute(ctx context.Context) error {
	rules, err := u.ruleRepo.ListEnabled(ctx)
	if err != nil {
		return fmt.Errorf("list enabled rules: %w", err)
	}
	if len(rules) == 0 {
		slog.Info("no enabled auto dm rules found")
		return nil
	}
	slog.Info("auto dm batch started", "enabled_rules", len(rules))

	// Group rules by user for optimized X API calls
	userRules := make(map[string][]*entity.AutoDMRule)
	for _, rule := range rules {
		userRules[rule.UserID] = append(userRules[rule.UserID], rule)
	}

	var totalErrors int

	// Step 1: Poll replies and enqueue matching ones (per user)
	for userID, rules := range userRules {
		if err := u.processUserRules(ctx, userID, rules); err != nil {
			slog.Error("failed to process auto dm rules for user",
				"user_id", userID, "rule_count", len(rules), "error", err)
			totalErrors++
		}
	}

	// Step 2: Send pending DMs
	if err := u.sendPendingDMs(ctx); err != nil {
		slog.Error("failed to send pending dms", "error", err)
		totalErrors++
	}

	if totalErrors > 0 {
		slog.Warn("auto dm batch completed with errors", "error_count", totalErrors)
	} else {
		slog.Info("auto dm batch completed successfully")
	}
	return nil
}

// processUserRules processes all rules for a single user with one X API call.
func (u *AutoDMBatchUsecase) processUserRules(ctx context.Context, userID string, rules []*entity.AutoDMRule) error {
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
	sinceID := minSinceID(rules)

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

	// Track max tweet ID for since_id update
	var maxTweetID string
	for _, tweet := range tweets {
		if tweet.ID > maxTweetID {
			maxTweetID = tweet.ID
		}
	}

	// Match keywords from all rules and enqueue
	var enqueued int
	for _, tweet := range tweets {
		if tweet.AuthorID == user.TwitterID {
			continue
		}

		ruleID, keyword := matchRules(tweet.Text, rules)
		if keyword == "" {
			continue
		}

		exists, err := u.logRepo.ExistsByReplyTweetID(ctx, tweet.ID, tweet.AuthorID)
		if err != nil {
			slog.Error("failed to check dm sent log", "tweet_id", tweet.ID, "error", err)
			continue
		}
		if exists {
			continue
		}

		item := &entity.DMPendingQueue{
			UserID:             userID,
			RuleID:             ruleID,
			RecipientTwitterID: tweet.AuthorID,
			ReplyTweetID:       tweet.ID,
			TriggerKeyword:     keyword,
			Status:             entity.DMPendingStatusPending,
		}
		if err := u.pendingRepo.Create(ctx, item); err != nil {
			slog.Error("failed to enqueue dm", "tweet_id", tweet.ID, "error", err)
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

	slog.Info("enqueued dms", "user_id", userID, "enqueued", enqueued, "total_replies", len(tweets))
	return nil
}

func (u *AutoDMBatchUsecase) sendPendingDMs(ctx context.Context) error {
	grouped, err := u.pendingRepo.ListPendingGroupedByUser(ctx)
	if err != nil {
		return fmt.Errorf("list pending dms: %w", err)
	}

	for userID, items := range grouped {
		conn, err := u.connRepo.FindByUserID(ctx, userID)
		if err != nil {
			slog.Error("failed to find twitter connection for dm send", "user_id", userID, "error", err)
			continue
		}
		if conn.Status != entity.TwitterConnected {
			continue
		}

		// Build a map of rule ID → rule for template lookup
		userRules, err := u.ruleRepo.ListByUserID(ctx, userID)
		if err != nil {
			slog.Error("failed to list rules for dm send", "user_id", userID, "error", err)
			continue
		}
		ruleMap := make(map[string]*entity.AutoDMRule, len(userRules))
		for _, r := range userRules {
			ruleMap[r.ID] = r
		}

		sent := 0
		for _, item := range items {
			if sent >= maxDMPerBatchPerUser {
				slog.Info("rate limit reached, remaining items stay pending",
					"user_id", userID, "remaining", len(items)-sent)
				break
			}

			rule, ok := ruleMap[item.RuleID]
			if !ok {
				slog.Warn("rule not found for pending dm, marking failed", "rule_id", item.RuleID)
				if err := u.pendingRepo.UpdateStatus(ctx, item.ID, entity.DMPendingStatusFailed); err != nil {
					slog.Error("failed to update pending status to failed", "id", item.ID, "error", err)
				}
				continue
			}

			dmText := rule.TemplateMessage
			if err := u.twitterGW.SendDirectMessage(ctx, conn.AccessToken, item.RecipientTwitterID, dmText); err != nil {
				slog.Error("failed to send dm",
					"user_id", userID, "recipient", item.RecipientTwitterID, "error", err)
				if err := u.pendingRepo.UpdateStatus(ctx, item.ID, entity.DMPendingStatusFailed); err != nil {
					slog.Error("failed to update pending status to failed", "id", item.ID, "error", err)
				}
				if apperror.IsCode(err, apperror.CodeResourceExhausted) {
					break
				}
				continue
			}

			logEntry := &entity.DMSentLog{
				UserID:             userID,
				RuleID:             item.RuleID,
				RecipientTwitterID: item.RecipientTwitterID,
				ReplyTweetID:       item.ReplyTweetID,
				TriggerKeyword:     item.TriggerKeyword,
				DMText:             dmText,
				SentAt:             time.Now(),
			}
			if err := u.logRepo.Create(ctx, logEntry); err != nil {
				slog.Error("failed to create dm sent log", "id", item.ID, "error", err)
			}

			if err := u.pendingRepo.DeleteByID(ctx, item.ID); err != nil {
				slog.Error("failed to delete pending dm", "id", item.ID, "error", err)
			}

			sent++
		}

		slog.Info("sent dms for user", "user_id", userID, "sent", sent, "total_pending", len(items))
	}

	return nil
}

// matchRules checks tweet text against all rules' keywords. Returns (ruleID, keyword) of first match.
func matchRules(text string, rules []*entity.AutoDMRule) (string, string) {
	lower := strings.ToLower(text)
	for _, rule := range rules {
		for _, kw := range rule.TriggerKeywords {
			if strings.Contains(lower, strings.ToLower(kw)) {
				return rule.ID, kw
			}
		}
	}
	return "", ""
}

// matchKeyword checks if the tweet text contains any of the trigger keywords (case-insensitive).
func matchKeyword(text string, keywords []string) string {
	lower := strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return kw
		}
	}
	return ""
}

// minSinceID returns the minimum last_checked_reply_id across all rules.
// Returns nil if any rule has no last_checked_reply_id (to search from startTime).
func minSinceID(rules []*entity.AutoDMRule) *string {
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
