package e2etest

import (
	"context"
	"fmt"
	"testing"

	"connectrpc.com/connect"

	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/gen/trendbird/v1/trendbirdv1connect"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// ---------------------------------------------------------------------------
// TestAutoDMService_ListAutoDMRules
// ---------------------------------------------------------------------------

func TestAutoDMService_ListAutoDMRules(t *testing.T) {
	t.Run("success_pro", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		r1 := seedAutoDMRule(t, env.db, user.ID,
			withRuleTriggerKeywords([]string{"golang", "rust"}),
			withRuleTemplateMessage("Thanks!"),
		)
		r2 := seedAutoDMRule(t, env.db, user.ID,
			withRuleTriggerKeywords([]string{"react"}),
			withRuleTemplateMessage("Hi there!"),
		)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAutoDMServiceClient)
		resp, err := client.ListAutoDMRules(context.Background(), connect.NewRequest(&trendbirdv1.ListAutoDMRulesRequest{}))
		if err != nil {
			t.Fatalf("ListAutoDMRules: %v", err)
		}

		rules := resp.Msg.GetRules()
		if got := len(rules); got != 2 {
			t.Fatalf("expected 2 rules, got %d", got)
		}

		// created_at ASC order
		if rules[0].GetId() != r1.ID {
			t.Errorf("first rule ID: want %s, got %s", r1.ID, rules[0].GetId())
		}
		if rules[1].GetId() != r2.ID {
			t.Errorf("second rule ID: want %s, got %s", r2.ID, rules[1].GetId())
		}
	})

	t.Run("empty", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAutoDMServiceClient)
		resp, err := client.ListAutoDMRules(context.Background(), connect.NewRequest(&trendbirdv1.ListAutoDMRulesRequest{}))
		if err != nil {
			t.Fatalf("ListAutoDMRules: %v", err)
		}

		if got := len(resp.Msg.GetRules()); got != 0 {
			t.Errorf("expected 0 rules, got %d", got)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.autoDMClient.ListAutoDMRules(context.Background(), connect.NewRequest(&trendbirdv1.ListAutoDMRulesRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

// ---------------------------------------------------------------------------
// TestAutoDMService_CreateAutoDMRule
// ---------------------------------------------------------------------------

func TestAutoDMService_CreateAutoDMRule(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAutoDMServiceClient)
		resp, err := client.CreateAutoDMRule(context.Background(), connect.NewRequest(&trendbirdv1.CreateAutoDMRuleRequest{
			TriggerKeywords: []string{"golang", "rust"},
			TemplateMessage: "Thanks for your reply!",
		}))
		if err != nil {
			t.Fatalf("CreateAutoDMRule: %v", err)
		}

		rule := resp.Msg.GetRule()
		if rule.GetId() == "" {
			t.Error("expected rule ID to be set")
		}
		if !rule.GetEnabled() {
			t.Error("expected rule to be enabled by default")
		}
		keywords := rule.GetTriggerKeywords()
		if len(keywords) != 2 || keywords[0] != "golang" || keywords[1] != "rust" {
			t.Errorf("trigger_keywords: want [golang rust], got %v", keywords)
		}
		if rule.GetTemplateMessage() != "Thanks for your reply!" {
			t.Errorf("template_message: want %q, got %q", "Thanks for your reply!", rule.GetTemplateMessage())
		}

		// DB 検証
		var count int64
		env.db.Model(&model.AutoDMRule{}).Where("user_id = ?", user.ID).Count(&count)
		if count != 1 {
			t.Errorf("auto_dm_rules count: want 1, got %d", count)
		}
	})

	t.Run("many_rules_allowed", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		// Seed 5 rules — OSS has no practical limit (999)
		for i := 0; i < 5; i++ {
			seedAutoDMRule(t, env.db, user.ID)
		}

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAutoDMServiceClient)
		_, err := client.CreateAutoDMRule(context.Background(), connect.NewRequest(&trendbirdv1.CreateAutoDMRuleRequest{
			TriggerKeywords: []string{"extra"},
			TemplateMessage: "6th rule should succeed in OSS",
		}))
		if err != nil {
			t.Fatalf("expected 6th rule to succeed, got: %v", err)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.autoDMClient.CreateAutoDMRule(context.Background(), connect.NewRequest(&trendbirdv1.CreateAutoDMRuleRequest{
			TriggerKeywords: []string{"test"},
			TemplateMessage: "test",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

// ---------------------------------------------------------------------------
// TestAutoDMService_UpdateAutoDMRule
// ---------------------------------------------------------------------------

func TestAutoDMService_UpdateAutoDMRule(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		rule := seedAutoDMRule(t, env.db, user.ID,
			withRuleTriggerKeywords([]string{"original"}),
			withRuleTemplateMessage("original message"),
		)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAutoDMServiceClient)
		resp, err := client.UpdateAutoDMRule(context.Background(), connect.NewRequest(&trendbirdv1.UpdateAutoDMRuleRequest{
			Id:              rule.ID,
			Enabled:         false,
			TriggerKeywords: []string{"updated", "keywords"},
			TemplateMessage: "updated message",
		}))
		if err != nil {
			t.Fatalf("UpdateAutoDMRule: %v", err)
		}

		updated := resp.Msg.GetRule()
		if updated.GetEnabled() {
			t.Error("expected enabled=false after update")
		}
		keywords := updated.GetTriggerKeywords()
		if len(keywords) != 2 || keywords[0] != "updated" || keywords[1] != "keywords" {
			t.Errorf("trigger_keywords: want [updated keywords], got %v", keywords)
		}
		if updated.GetTemplateMessage() != "updated message" {
			t.Errorf("template_message: want %q, got %q", "updated message", updated.GetTemplateMessage())
		}

		// DB 検証
		var dbRule model.AutoDMRule
		if err := env.db.First(&dbRule, "id = ?", rule.ID).Error; err != nil {
			t.Fatalf("fetch rule: %v", err)
		}
		if dbRule.Enabled {
			t.Error("DB enabled: want false, got true")
		}
	})

	t.Run("permission_denied", func(t *testing.T) {
		env := setupTest(t)
		owner := seedUser(t, env.db)
		other := seedUser(t, env.db)
		rule := seedAutoDMRule(t, env.db, owner.ID)

		// other tries to update owner's rule
		client := connectClient(t, env, other.ID, trendbirdv1connect.NewAutoDMServiceClient)
		_, err := client.UpdateAutoDMRule(context.Background(), connect.NewRequest(&trendbirdv1.UpdateAutoDMRuleRequest{
			Id:              rule.ID,
			Enabled:         false,
			TriggerKeywords: []string{"hacked"},
			TemplateMessage: "hacked",
		}))
		assertConnectCode(t, err, connect.CodePermissionDenied)
	})

	t.Run("not_found", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAutoDMServiceClient)
		_, err := client.UpdateAutoDMRule(context.Background(), connect.NewRequest(&trendbirdv1.UpdateAutoDMRuleRequest{
			Id:              "00000000-0000-0000-0000-000000000000",
			Enabled:         false,
			TriggerKeywords: []string{"test"},
			TemplateMessage: "test",
		}))
		assertConnectCode(t, err, connect.CodeNotFound)
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.autoDMClient.UpdateAutoDMRule(context.Background(), connect.NewRequest(&trendbirdv1.UpdateAutoDMRuleRequest{
			Id: "some-id",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

// ---------------------------------------------------------------------------
// TestAutoDMService_DeleteAutoDMRule
// ---------------------------------------------------------------------------

func TestAutoDMService_DeleteAutoDMRule(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		rule := seedAutoDMRule(t, env.db, user.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAutoDMServiceClient)
		_, err := client.DeleteAutoDMRule(context.Background(), connect.NewRequest(&trendbirdv1.DeleteAutoDMRuleRequest{
			Id: rule.ID,
		}))
		if err != nil {
			t.Fatalf("DeleteAutoDMRule: %v", err)
		}

		// DB 検証: ルールが削除されていること
		var count int64
		env.db.Model(&model.AutoDMRule{}).Where("id = ?", rule.ID).Count(&count)
		if count != 0 {
			t.Errorf("auto_dm_rules count: want 0, got %d", count)
		}
	})

	t.Run("cascade_sent_logs", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		rule := seedAutoDMRule(t, env.db, user.ID)

		// Seed sent logs linked to this rule
		seedDMSentLog(t, env.db, user.ID, rule.ID)
		seedDMSentLog(t, env.db, user.ID, rule.ID)

		// Verify sent logs exist
		var logCount int64
		env.db.Model(&model.DMSentLog{}).Where("rule_id = ?", rule.ID).Count(&logCount)
		if logCount != 2 {
			t.Fatalf("pre-condition: expected 2 sent logs, got %d", logCount)
		}

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAutoDMServiceClient)
		_, err := client.DeleteAutoDMRule(context.Background(), connect.NewRequest(&trendbirdv1.DeleteAutoDMRuleRequest{
			Id: rule.ID,
		}))
		if err != nil {
			t.Fatalf("DeleteAutoDMRule: %v", err)
		}

		// Verify sent logs are also deleted via CASCADE
		env.db.Model(&model.DMSentLog{}).Where("rule_id = ?", rule.ID).Count(&logCount)
		if logCount != 0 {
			t.Errorf("dm_sent_logs should be cascade-deleted, got count=%d", logCount)
		}
	})

	t.Run("not_found", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAutoDMServiceClient)
		_, err := client.DeleteAutoDMRule(context.Background(), connect.NewRequest(&trendbirdv1.DeleteAutoDMRuleRequest{
			Id: "00000000-0000-0000-0000-000000000000",
		}))
		assertConnectCode(t, err, connect.CodeNotFound)
	})

	t.Run("permission_denied", func(t *testing.T) {
		env := setupTest(t)
		owner := seedUser(t, env.db)
		other := seedUser(t, env.db)
		rule := seedAutoDMRule(t, env.db, owner.ID)

		// other tries to delete owner's rule
		client := connectClient(t, env, other.ID, trendbirdv1connect.NewAutoDMServiceClient)
		_, err := client.DeleteAutoDMRule(context.Background(), connect.NewRequest(&trendbirdv1.DeleteAutoDMRuleRequest{
			Id: rule.ID,
		}))
		assertConnectCode(t, err, connect.CodePermissionDenied)
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.autoDMClient.DeleteAutoDMRule(context.Background(), connect.NewRequest(&trendbirdv1.DeleteAutoDMRuleRequest{
			Id: "some-id",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

// ---------------------------------------------------------------------------
// TestAutoDMService_GetDMSentLogs
// ---------------------------------------------------------------------------

func TestAutoDMService_GetDMSentLogs(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		rule := seedAutoDMRule(t, env.db, user.ID)

		// Seed 3 sent logs
		seedDMSentLog(t, env.db, user.ID, rule.ID, withDMLogTriggerKeyword("golang"))
		seedDMSentLog(t, env.db, user.ID, rule.ID, withDMLogTriggerKeyword("rust"))
		seedDMSentLog(t, env.db, user.ID, rule.ID, withDMLogTriggerKeyword("react"))

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAutoDMServiceClient)
		resp, err := client.GetDMSentLogs(context.Background(), connect.NewRequest(&trendbirdv1.GetDMSentLogsRequest{}))
		if err != nil {
			t.Fatalf("GetDMSentLogs: %v", err)
		}

		logs := resp.Msg.GetLogs()
		if got := len(logs); got != 3 {
			t.Fatalf("expected 3 logs, got %d", got)
		}
	})

	t.Run("empty", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAutoDMServiceClient)
		resp, err := client.GetDMSentLogs(context.Background(), connect.NewRequest(&trendbirdv1.GetDMSentLogsRequest{}))
		if err != nil {
			t.Fatalf("GetDMSentLogs: %v", err)
		}

		if got := len(resp.Msg.GetLogs()); got != 0 {
			t.Errorf("expected 0 logs, got %d", got)
		}
	})

	t.Run("only_own_logs", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db) 
		other := seedUser(t, env.db)
		userRule := seedAutoDMRule(t, env.db, user.ID)
		otherRule := seedAutoDMRule(t, env.db, other.ID)

		seedDMSentLog(t, env.db, user.ID, userRule.ID)
		seedDMSentLog(t, env.db, other.ID, otherRule.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAutoDMServiceClient)
		resp, err := client.GetDMSentLogs(context.Background(), connect.NewRequest(&trendbirdv1.GetDMSentLogsRequest{}))
		if err != nil {
			t.Fatalf("GetDMSentLogs: %v", err)
		}

		if got := len(resp.Msg.GetLogs()); got != 1 {
			t.Errorf("expected 1 log (own only), got %d", got)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.autoDMClient.GetDMSentLogs(context.Background(), connect.NewRequest(&trendbirdv1.GetDMSentLogsRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

// ---------------------------------------------------------------------------
// Ensure fmt is used (for unused import prevention)
// ---------------------------------------------------------------------------

var _ = fmt.Sprintf
