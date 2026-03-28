package e2etest

import (
	"context"
	"fmt"
	"testing"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/gateway"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
	"github.com/trendbird/backend/internal/infrastructure/persistence/repository"
	"github.com/trendbird/backend/internal/usecase"
)

func TestAutoReplyBatch_FullFlow(t *testing.T) {
	env := setupTest(t)
	user := seedUser(t, env.db, withTwitterID("tw-reply-user-1"), withTwitterHandle("replyuser1"))
	seedTwitterConnection(t, env.db, user.ID)

	rule := seedAutoReplyRule(t, env.db, user.ID,
		withAutoReplyTargetTweetID("target-post-100"),
		withAutoReplyKeywords([]string{"hello"}),
		withAutoReplyTemplate("Thanks for your reply!"),
	)

	var postReplyCalls int
	env.mockTwitter.SearchRecentTweetsFn = func(_ context.Context, _ string, input gateway.SearchTweetsInput) ([]gateway.Tweet, error) {
		return []gateway.Tweet{
			{ID: "reply-1", Text: "hello world", AuthorID: "other-user-1", ConversationID: "target-post-100"},
		}, nil
	}
	env.mockTwitter.PostReplyFn = func(_ context.Context, _ string, text string, inReplyTo string) (string, error) {
		postReplyCalls++
		return fmt.Sprintf("sent-reply-%d", postReplyCalls), nil
	}

	uc := usecase.NewAutoReplyBatchUsecase(
		repository.NewAutoReplyRuleRepository(env.db),
		repository.NewReplySentLogRepository(env.db),
		repository.NewReplyPendingQueueRepository(env.db),
		repository.NewTwitterConnectionRepository(env.db),
		repository.NewUserRepository(env.db),
		env.mockTwitter,
	)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if postReplyCalls != 1 {
		t.Errorf("PostReply calls: want 1, got %d", postReplyCalls)
	}

	// Verify reply sent log was created
	logRepo := repository.NewReplySentLogRepository(env.db)
	logs, err := logRepo.ListByUserID(context.Background(), user.ID, 10)
	if err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("reply sent logs: want 1, got %d", len(logs))
	}
	if logs[0].OriginalTweetID != "reply-1" {
		t.Errorf("original_tweet_id: want reply-1, got %s", logs[0].OriginalTweetID)
	}
	if logs[0].ReplyText != "Thanks for your reply!" {
		t.Errorf("reply_text: want template, got %s", logs[0].ReplyText)
	}

	// Verify since_id was updated
	ruleRepo := repository.NewAutoReplyRuleRepository(env.db)
	updated, _ := ruleRepo.FindByID(context.Background(), rule.ID)
	if updated.LastCheckedReplyID == nil || *updated.LastCheckedReplyID != "reply-1" {
		t.Errorf("last_checked_reply_id: want reply-1, got %v", updated.LastCheckedReplyID)
	}
}

func TestAutoReplyBatch_DisabledRuleSkipped(t *testing.T) {
	env := setupTest(t)
	user := seedUser(t, env.db, withTwitterID("tw-reply-user-2"), withTwitterHandle("replyuser2"))
	seedTwitterConnection(t, env.db, user.ID)

	seedAutoReplyRule(t, env.db, user.ID,
		withAutoReplyEnabled(false),
		withAutoReplyTargetTweetID("target-post-200"),
		withAutoReplyKeywords([]string{"hello"}),
	)

	env.mockTwitter.SearchRecentTweetsFn = func(_ context.Context, _ string, _ gateway.SearchTweetsInput) ([]gateway.Tweet, error) {
		t.Error("SearchRecentTweets should not be called for disabled rules")
		return nil, nil
	}

	uc := usecase.NewAutoReplyBatchUsecase(
		repository.NewAutoReplyRuleRepository(env.db),
		repository.NewReplySentLogRepository(env.db),
		repository.NewReplyPendingQueueRepository(env.db),
		repository.NewTwitterConnectionRepository(env.db),
		repository.NewUserRepository(env.db),
		env.mockTwitter,
	)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute: %v", err)
	}
}

func TestAutoReplyBatch_SelfReplySkipped(t *testing.T) {
	env := setupTest(t)
	user := seedUser(t, env.db, withTwitterID("tw-reply-user-3"), withTwitterHandle("replyuser3"))
	seedTwitterConnection(t, env.db, user.ID)

	seedAutoReplyRule(t, env.db, user.ID,
		withAutoReplyTargetTweetID("target-post-300"),
		withAutoReplyKeywords([]string{"hello"}),
		withAutoReplyTemplate("reply template"),
	)

	env.mockTwitter.SearchRecentTweetsFn = func(_ context.Context, _ string, _ gateway.SearchTweetsInput) ([]gateway.Tweet, error) {
		return []gateway.Tweet{
			{ID: "self-reply-1", Text: "hello", AuthorID: "tw-reply-user-3", ConversationID: "target-post-300"},
		}, nil
	}

	var postReplyCalls int
	env.mockTwitter.PostReplyFn = func(_ context.Context, _ string, _ string, _ string) (string, error) {
		postReplyCalls++
		return "sent-reply", nil
	}

	uc := usecase.NewAutoReplyBatchUsecase(
		repository.NewAutoReplyRuleRepository(env.db),
		repository.NewReplySentLogRepository(env.db),
		repository.NewReplyPendingQueueRepository(env.db),
		repository.NewTwitterConnectionRepository(env.db),
		repository.NewUserRepository(env.db),
		env.mockTwitter,
	)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if postReplyCalls != 0 {
		t.Errorf("PostReply calls: want 0 (self-reply skipped), got %d", postReplyCalls)
	}
}

func TestAutoReplyBatch_DuplicatePrevention(t *testing.T) {
	env := setupTest(t)
	user := seedUser(t, env.db, withTwitterID("tw-reply-user-4"), withTwitterHandle("replyuser4"))
	seedTwitterConnection(t, env.db, user.ID)

	rule := seedAutoReplyRule(t, env.db, user.ID,
		withAutoReplyTargetTweetID("target-post-400"),
		withAutoReplyKeywords([]string{"hello"}),
		withAutoReplyTemplate("reply template"),
	)

	// Pre-seed a sent log for this reply
	env.db.Create(&model.ReplySentLog{
		UserID:           user.ID,
		RuleID:           rule.ID,
		OriginalTweetID:  "already-replied-tweet",
		OriginalAuthorID: "other-user-4",
		ReplyTweetID:     "existing-reply",
		TriggerKeyword:   "hello",
		ReplyText:        "existing reply",
	})

	env.mockTwitter.SearchRecentTweetsFn = func(_ context.Context, _ string, _ gateway.SearchTweetsInput) ([]gateway.Tweet, error) {
		return []gateway.Tweet{
			{ID: "already-replied-tweet", Text: "hello", AuthorID: "other-user-4", ConversationID: "target-post-400"},
		}, nil
	}

	var postReplyCalls int
	env.mockTwitter.PostReplyFn = func(_ context.Context, _ string, _ string, _ string) (string, error) {
		postReplyCalls++
		return "sent-reply", nil
	}

	uc := usecase.NewAutoReplyBatchUsecase(
		repository.NewAutoReplyRuleRepository(env.db),
		repository.NewReplySentLogRepository(env.db),
		repository.NewReplyPendingQueueRepository(env.db),
		repository.NewTwitterConnectionRepository(env.db),
		repository.NewUserRepository(env.db),
		env.mockTwitter,
	)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if postReplyCalls != 0 {
		t.Errorf("PostReply calls: want 0 (duplicate prevented), got %d", postReplyCalls)
	}
}

func TestAutoReplyBatch_ConversationIDMismatchSkipped(t *testing.T) {
	env := setupTest(t)
	user := seedUser(t, env.db, withTwitterID("tw-reply-user-5"), withTwitterHandle("replyuser5"))
	seedTwitterConnection(t, env.db, user.ID)

	seedAutoReplyRule(t, env.db, user.ID,
		withAutoReplyTargetTweetID("target-post-500"),
		withAutoReplyKeywords([]string{"hello"}),
		withAutoReplyTemplate("reply template"),
	)

	env.mockTwitter.SearchRecentTweetsFn = func(_ context.Context, _ string, _ gateway.SearchTweetsInput) ([]gateway.Tweet, error) {
		return []gateway.Tweet{
			{ID: "diff-conv-reply", Text: "hello", AuthorID: "other-user-5", ConversationID: "different-post"},
		}, nil
	}

	var postReplyCalls int
	env.mockTwitter.PostReplyFn = func(_ context.Context, _ string, _ string, _ string) (string, error) {
		postReplyCalls++
		return "sent-reply", nil
	}

	uc := usecase.NewAutoReplyBatchUsecase(
		repository.NewAutoReplyRuleRepository(env.db),
		repository.NewReplySentLogRepository(env.db),
		repository.NewReplyPendingQueueRepository(env.db),
		repository.NewTwitterConnectionRepository(env.db),
		repository.NewUserRepository(env.db),
		env.mockTwitter,
	)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if postReplyCalls != 0 {
		t.Errorf("PostReply calls: want 0 (conversation_id mismatch), got %d", postReplyCalls)
	}
}

func TestAutoReplyBatch_KeywordMismatchSkipped(t *testing.T) {
	env := setupTest(t)
	user := seedUser(t, env.db, withTwitterID("tw-reply-user-6"), withTwitterHandle("replyuser6"))
	seedTwitterConnection(t, env.db, user.ID)

	seedAutoReplyRule(t, env.db, user.ID,
		withAutoReplyTargetTweetID("target-post-600"),
		withAutoReplyKeywords([]string{"specific-keyword"}),
		withAutoReplyTemplate("reply template"),
	)

	env.mockTwitter.SearchRecentTweetsFn = func(_ context.Context, _ string, _ gateway.SearchTweetsInput) ([]gateway.Tweet, error) {
		return []gateway.Tweet{
			{ID: "no-keyword-reply", Text: "unrelated text", AuthorID: "other-user-6", ConversationID: "target-post-600"},
		}, nil
	}

	var postReplyCalls int
	env.mockTwitter.PostReplyFn = func(_ context.Context, _ string, _ string, _ string) (string, error) {
		postReplyCalls++
		return "sent-reply", nil
	}

	uc := usecase.NewAutoReplyBatchUsecase(
		repository.NewAutoReplyRuleRepository(env.db),
		repository.NewReplySentLogRepository(env.db),
		repository.NewReplyPendingQueueRepository(env.db),
		repository.NewTwitterConnectionRepository(env.db),
		repository.NewUserRepository(env.db),
		env.mockTwitter,
	)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if postReplyCalls != 0 {
		t.Errorf("PostReply calls: want 0 (keyword mismatch), got %d", postReplyCalls)
	}
}

func TestAutoReplyBatch_RateLimitStops(t *testing.T) {
	env := setupTest(t)
	user := seedUser(t, env.db, withTwitterID("tw-reply-user-7"), withTwitterHandle("replyuser7"))
	seedTwitterConnection(t, env.db, user.ID)

	rule := seedAutoReplyRule(t, env.db, user.ID,
		withAutoReplyTargetTweetID("target-post-700"),
		withAutoReplyKeywords([]string{"hello"}),
		withAutoReplyTemplate("reply template"),
	)

	// Pre-seed pending items
	for i := 0; i < 3; i++ {
		seedReplyPendingQueue(t, env.db, user.ID, rule.ID,
			withReplyPendingOriginalTweetID(fmt.Sprintf("pending-tweet-%d", i)),
			withReplyPendingOriginalAuthorID(fmt.Sprintf("author-pending-%d", i)),
		)
	}

	// Don't return new tweets from search
	env.mockTwitter.SearchRecentTweetsFn = func(_ context.Context, _ string, _ gateway.SearchTweetsInput) ([]gateway.Tweet, error) {
		return []gateway.Tweet{}, nil
	}

	var postReplyCalls int
	env.mockTwitter.PostReplyFn = func(_ context.Context, _ string, _ string, _ string) (string, error) {
		postReplyCalls++
		if postReplyCalls >= 2 {
			return "", apperror.Wrap(apperror.CodeResourceExhausted, "rate limit exceeded", fmt.Errorf("429"))
		}
		return fmt.Sprintf("sent-reply-%d", postReplyCalls), nil
	}

	uc := usecase.NewAutoReplyBatchUsecase(
		repository.NewAutoReplyRuleRepository(env.db),
		repository.NewReplySentLogRepository(env.db),
		repository.NewReplyPendingQueueRepository(env.db),
		repository.NewTwitterConnectionRepository(env.db),
		repository.NewUserRepository(env.db),
		env.mockTwitter,
	)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// Should have attempted 2 calls (1 success + 1 rate limit)
	if postReplyCalls != 2 {
		t.Errorf("PostReply calls: want 2, got %d", postReplyCalls)
	}

	// Verify 1 pending item remains
	pendingRepo := repository.NewReplyPendingQueueRepository(env.db)
	grouped, _ := pendingRepo.ListPendingGroupedByUser(context.Background())
	remaining := len(grouped[user.ID])
	if remaining != 1 {
		t.Errorf("remaining pending: want 1, got %d", remaining)
	}
}

func TestAutoReplyBatch_NoConnectionSkipped(t *testing.T) {
	env := setupTest(t)
	user := seedUser(t, env.db, withTwitterID("tw-reply-user-8"), withTwitterHandle("replyuser8"))
	seedTwitterConnection(t, env.db, user.ID, withTwitterConnStatus(1)) // Disconnected

	seedAutoReplyRule(t, env.db, user.ID,
		withAutoReplyTargetTweetID("target-post-800"),
		withAutoReplyKeywords([]string{"hello"}),
	)

	env.mockTwitter.SearchRecentTweetsFn = func(_ context.Context, _ string, _ gateway.SearchTweetsInput) ([]gateway.Tweet, error) {
		t.Error("SearchRecentTweets should not be called for disconnected users")
		return nil, nil
	}

	uc := usecase.NewAutoReplyBatchUsecase(
		repository.NewAutoReplyRuleRepository(env.db),
		repository.NewReplySentLogRepository(env.db),
		repository.NewReplyPendingQueueRepository(env.db),
		repository.NewTwitterConnectionRepository(env.db),
		repository.NewUserRepository(env.db),
		env.mockTwitter,
	)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute: %v", err)
	}
}
