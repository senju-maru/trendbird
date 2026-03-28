package e2etest

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/gateway"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
	"github.com/trendbird/backend/internal/infrastructure/persistence/repository"
	"github.com/trendbird/backend/internal/usecase"
)

func TestReplyDMBatch_FullFlow(t *testing.T) {
	env := setupTest(t)

	user := seedUser(t, env.db, withTwitterID("tw-user-dm"), withTwitterHandle("dmuser"))
	seedTwitterConnection(t, env.db, user.ID)
	rule := seedAutoDMRule(t, env.db, user.ID,
		withRuleTriggerKeywords([]string{"interested"}),
		withRuleTemplateMessage("Thanks for your interest!"),
	)

	// Mock: SearchRecentTweets returns a reply matching the keyword
	env.mockTwitter.SearchRecentTweetsFn = func(_ context.Context, _ string, _ gateway.SearchTweetsInput) ([]gateway.Tweet, error) {
		return []gateway.Tweet{
			{
				ID:           "reply-tweet-001",
				Text:         "I'm interested in this!",
				AuthorID:     "replier-001",
				AuthorName:   "Replier",
				AuthorHandle: "replier1",
				CreatedAt:    time.Now(),
			},
		}, nil
	}

	// Track SendDirectMessage calls
	var dmSentCount atomic.Int64
	env.mockTwitter.SendDirectMessageFn = func(_ context.Context, _ string, _ string, _ string) error {
		dmSentCount.Add(1)
		return nil
	}

	ruleRepo := repository.NewAutoDMRuleRepository(env.db)
	logRepo := repository.NewDMSentLogRepository(env.db)
	pendingRepo := repository.NewDMPendingQueueRepository(env.db)
	connRepo := repository.NewTwitterConnectionRepository(env.db)
	userRepo := repository.NewUserRepository(env.db)

	uc := usecase.NewAutoDMBatchUsecase(ruleRepo, logRepo, pendingRepo, connRepo, userRepo, env.mockTwitter)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify DM was sent
	if dmSentCount.Load() != 1 {
		t.Errorf("expected 1 DM sent, got %d", dmSentCount.Load())
	}

	// Verify DMSentLog created
	var logCount int64
	env.db.Model(&model.DMSentLog{}).Where("user_id = ?", user.ID).Count(&logCount)
	if logCount != 1 {
		t.Errorf("expected 1 dm_sent_log, got %d", logCount)
	}

	// Verify pending queue is empty (item deleted after send)
	var pendingCount int64
	env.db.Model(&model.DMPendingQueue{}).Where("user_id = ?", user.ID).Count(&pendingCount)
	if pendingCount != 0 {
		t.Errorf("expected 0 pending items, got %d", pendingCount)
	}

	// Verify LastCheckedReplyID updated
	var dbRule model.AutoDMRule
	env.db.First(&dbRule, "id = ?", rule.ID)
	if dbRule.LastCheckedReplyID == nil {
		t.Error("expected LastCheckedReplyID to be set")
	}
}

func TestReplyDMBatch_NoEnabledRules(t *testing.T) {
	env := setupTest(t)

	user := seedUser(t, env.db)
	seedAutoDMRule(t, env.db, user.ID, withRuleEnabled(false))

	ruleRepo := repository.NewAutoDMRuleRepository(env.db)
	logRepo := repository.NewDMSentLogRepository(env.db)
	pendingRepo := repository.NewDMPendingQueueRepository(env.db)
	connRepo := repository.NewTwitterConnectionRepository(env.db)
	userRepo := repository.NewUserRepository(env.db)

	uc := usecase.NewAutoDMBatchUsecase(ruleRepo, logRepo, pendingRepo, connRepo, userRepo, env.mockTwitter)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}

func TestReplyDMBatch_NoConnection(t *testing.T) {
	env := setupTest(t)

	user := seedUser(t, env.db, withTwitterHandle("noconn"))
	// No seedTwitterConnection — should skip this user
	seedAutoDMRule(t, env.db, user.ID, withRuleTriggerKeywords([]string{"test"}))

	ruleRepo := repository.NewAutoDMRuleRepository(env.db)
	logRepo := repository.NewDMSentLogRepository(env.db)
	pendingRepo := repository.NewDMPendingQueueRepository(env.db)
	connRepo := repository.NewTwitterConnectionRepository(env.db)
	userRepo := repository.NewUserRepository(env.db)

	uc := usecase.NewAutoDMBatchUsecase(ruleRepo, logRepo, pendingRepo, connRepo, userRepo, env.mockTwitter)

	// Should complete without error (user skipped)
	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}

func TestReplyDMBatch_DuplicatePrevention(t *testing.T) {
	env := setupTest(t)

	user := seedUser(t, env.db, withTwitterID("tw-dup"), withTwitterHandle("dupuser"))
	seedTwitterConnection(t, env.db, user.ID)
	rule := seedAutoDMRule(t, env.db, user.ID,
		withRuleTriggerKeywords([]string{"interested"}),
		withRuleTemplateMessage("Thanks!"),
	)

	// Pre-seed a DMSentLog for the same reply_tweet_id + recipient
	seedDMSentLog(t, env.db, user.ID, rule.ID,
		withDMLogRecipientID("replier-dup"),
	)
	// Get the seeded log to find the reply_tweet_id
	var existingLog model.DMSentLog
	env.db.Where("user_id = ?", user.ID).First(&existingLog)

	// Mock: return a reply with the same author + tweet that was already DM'd
	env.mockTwitter.SearchRecentTweetsFn = func(_ context.Context, _ string, _ gateway.SearchTweetsInput) ([]gateway.Tweet, error) {
		return []gateway.Tweet{
			{
				ID:           existingLog.ReplyTweetID,
				Text:         "I'm interested!",
				AuthorID:     existingLog.RecipientTwitterID,
				AuthorName:   "Dup Replier",
				AuthorHandle: "dupreplier",
				CreatedAt:    time.Now(),
			},
		}, nil
	}

	var dmSentCount atomic.Int64
	env.mockTwitter.SendDirectMessageFn = func(_ context.Context, _ string, _ string, _ string) error {
		dmSentCount.Add(1)
		return nil
	}

	ruleRepo := repository.NewAutoDMRuleRepository(env.db)
	logRepo := repository.NewDMSentLogRepository(env.db)
	pendingRepo := repository.NewDMPendingQueueRepository(env.db)
	connRepo := repository.NewTwitterConnectionRepository(env.db)
	userRepo := repository.NewUserRepository(env.db)

	uc := usecase.NewAutoDMBatchUsecase(ruleRepo, logRepo, pendingRepo, connRepo, userRepo, env.mockTwitter)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// No new DMs should be sent (duplicate)
	if dmSentCount.Load() != 0 {
		t.Errorf("expected 0 DMs (duplicate prevention), got %d", dmSentCount.Load())
	}
}

func TestReplyDMBatch_SelfReplySkipped(t *testing.T) {
	env := setupTest(t)

	user := seedUser(t, env.db, withTwitterID("tw-self-001"), withTwitterHandle("selfuser"))
	seedTwitterConnection(t, env.db, user.ID)
	seedAutoDMRule(t, env.db, user.ID,
		withRuleTriggerKeywords([]string{"interested"}),
		withRuleTemplateMessage("Thanks!"),
	)

	// Mock: return a reply from the user themselves
	env.mockTwitter.SearchRecentTweetsFn = func(_ context.Context, _ string, _ gateway.SearchTweetsInput) ([]gateway.Tweet, error) {
		return []gateway.Tweet{
			{
				ID:           "self-reply-001",
				Text:         "I'm interested in my own post",
				AuthorID:     "tw-self-001", // Same as user.TwitterID
				AuthorName:   "Self User",
				AuthorHandle: "selfuser",
				CreatedAt:    time.Now(),
			},
		}, nil
	}

	var dmSentCount atomic.Int64
	env.mockTwitter.SendDirectMessageFn = func(_ context.Context, _ string, _ string, _ string) error {
		dmSentCount.Add(1)
		return nil
	}

	ruleRepo := repository.NewAutoDMRuleRepository(env.db)
	logRepo := repository.NewDMSentLogRepository(env.db)
	pendingRepo := repository.NewDMPendingQueueRepository(env.db)
	connRepo := repository.NewTwitterConnectionRepository(env.db)
	userRepo := repository.NewUserRepository(env.db)

	uc := usecase.NewAutoDMBatchUsecase(ruleRepo, logRepo, pendingRepo, connRepo, userRepo, env.mockTwitter)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// No DMs — self replies are skipped
	if dmSentCount.Load() != 0 {
		t.Errorf("expected 0 DMs (self reply skipped), got %d", dmSentCount.Load())
	}
}

func TestReplyDMBatch_SendFailure(t *testing.T) {
	env := setupTest(t)

	user := seedUser(t, env.db, withTwitterID("tw-fail"), withTwitterHandle("failuser"))
	seedTwitterConnection(t, env.db, user.ID)
	rule := seedAutoDMRule(t, env.db, user.ID,
		withRuleTriggerKeywords([]string{"interested"}),
		withRuleTemplateMessage("Thanks!"),
	)

	// Pre-seed a pending DM item
	pending := seedDMPendingQueue(t, env.db, user.ID, rule.ID,
		withDMPendingRecipientID("fail-recipient"),
		withDMPendingTriggerKeyword("interested"),
	)

	// Mock: no new replies (skip phase 1)
	env.mockTwitter.SearchRecentTweetsFn = func(_ context.Context, _ string, _ gateway.SearchTweetsInput) ([]gateway.Tweet, error) {
		return []gateway.Tweet{}, nil
	}

	// Mock: SendDirectMessage fails
	env.mockTwitter.SendDirectMessageFn = func(_ context.Context, _ string, _ string, _ string) error {
		return fmt.Errorf("DM send failed")
	}

	ruleRepo := repository.NewAutoDMRuleRepository(env.db)
	logRepo := repository.NewDMSentLogRepository(env.db)
	pendingRepo := repository.NewDMPendingQueueRepository(env.db)
	connRepo := repository.NewTwitterConnectionRepository(env.db)
	userRepo := repository.NewUserRepository(env.db)

	uc := usecase.NewAutoDMBatchUsecase(ruleRepo, logRepo, pendingRepo, connRepo, userRepo, env.mockTwitter)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify pending item status changed to Failed (3)
	var dbPending model.DMPendingQueue
	env.db.First(&dbPending, "id = ?", pending.ID)
	if dbPending.Status != 3 { // DMPendingStatusFailed
		t.Errorf("pending status: want 3 (Failed), got %d", dbPending.Status)
	}
}

func TestReplyDMBatch_MaxDMPerBatchPerUser(t *testing.T) {
	env := setupTest(t)

	user := seedUser(t, env.db, withTwitterID("tw-max"), withTwitterHandle("maxuser"))
	seedTwitterConnection(t, env.db, user.ID)
	rule := seedAutoDMRule(t, env.db, user.ID,
		withRuleTriggerKeywords([]string{"interested"}),
		withRuleTemplateMessage("Thanks!"),
	)

	// Pre-seed 16 pending DM items (max is 15 per batch per user)
	for i := 0; i < 16; i++ {
		seedDMPendingQueue(t, env.db, user.ID, rule.ID,
			withDMPendingRecipientID(fmt.Sprintf("max-recipient-%d", i)),
			withDMPendingReplyTweetID(fmt.Sprintf("max-reply-%d", i)),
			withDMPendingTriggerKeyword("interested"),
		)
	}

	// Mock: no new replies (skip phase 1)
	env.mockTwitter.SearchRecentTweetsFn = func(_ context.Context, _ string, _ gateway.SearchTweetsInput) ([]gateway.Tweet, error) {
		return []gateway.Tweet{}, nil
	}

	var dmSentCount atomic.Int64
	env.mockTwitter.SendDirectMessageFn = func(_ context.Context, _ string, _ string, _ string) error {
		dmSentCount.Add(1)
		return nil
	}

	ruleRepo := repository.NewAutoDMRuleRepository(env.db)
	logRepo := repository.NewDMSentLogRepository(env.db)
	pendingRepo := repository.NewDMPendingQueueRepository(env.db)
	connRepo := repository.NewTwitterConnectionRepository(env.db)
	userRepo := repository.NewUserRepository(env.db)

	uc := usecase.NewAutoDMBatchUsecase(ruleRepo, logRepo, pendingRepo, connRepo, userRepo, env.mockTwitter)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Only 15 DMs should be sent (rate limit: maxDMPerBatchPerUser = 15)
	sent := dmSentCount.Load()
	if sent != 15 {
		t.Errorf("expected 15 DMs sent (max per batch), got %d", sent)
	}

	// 1 item should still be pending
	var remainingPending int64
	env.db.Model(&model.DMPendingQueue{}).Where("user_id = ? AND status = 1", user.ID).Count(&remainingPending)
	if remainingPending != 1 {
		t.Errorf("expected 1 remaining pending item, got %d", remainingPending)
	}
}

// TestReplyDMBatch_RateLimitStopsSending verifies that a rate limit error during DM sending
// stops processing for the user.
func TestReplyDMBatch_RateLimitStopsSending(t *testing.T) {
	env := setupTest(t)

	user := seedUser(t, env.db, withTwitterID("tw-rl"), withTwitterHandle("rluser"))
	seedTwitterConnection(t, env.db, user.ID)
	rule := seedAutoDMRule(t, env.db, user.ID,
		withRuleTriggerKeywords([]string{"interested"}),
		withRuleTemplateMessage("Thanks!"),
	)

	// Pre-seed 3 pending DM items
	for i := 0; i < 3; i++ {
		seedDMPendingQueue(t, env.db, user.ID, rule.ID,
			withDMPendingRecipientID(fmt.Sprintf("rl-recipient-%d", i)),
			withDMPendingReplyTweetID(fmt.Sprintf("rl-reply-%d", i)),
		)
	}

	// Mock: no new replies
	env.mockTwitter.SearchRecentTweetsFn = func(_ context.Context, _ string, _ gateway.SearchTweetsInput) ([]gateway.Tweet, error) {
		return []gateway.Tweet{}, nil
	}

	// Mock: first DM succeeds, second hits rate limit
	var dmCallCount atomic.Int64
	env.mockTwitter.SendDirectMessageFn = func(_ context.Context, _ string, _ string, _ string) error {
		n := dmCallCount.Add(1)
		if n >= 2 {
			return apperror.ResourceExhausted("rate limit")
		}
		return nil
	}

	ruleRepo := repository.NewAutoDMRuleRepository(env.db)
	logRepo := repository.NewDMSentLogRepository(env.db)
	pendingRepo := repository.NewDMPendingQueueRepository(env.db)
	connRepo := repository.NewTwitterConnectionRepository(env.db)
	userRepo := repository.NewUserRepository(env.db)

	uc := usecase.NewAutoDMBatchUsecase(ruleRepo, logRepo, pendingRepo, connRepo, userRepo, env.mockTwitter)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Only 1 DM should have been successfully sent (second failed with rate limit, third not attempted)
	var logCount int64
	env.db.Model(&model.DMSentLog{}).Where("user_id = ?", user.ID).Count(&logCount)
	if logCount != 1 {
		t.Errorf("expected 1 dm_sent_log (before rate limit), got %d", logCount)
	}
}
