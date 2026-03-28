package usecase

import (
	"context"
	"testing"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
)

// ---------------------------------------------------------------------------
// Mock implementations
// ---------------------------------------------------------------------------

type mockAutoDMRuleRepo struct {
	FindByIDFn              func(ctx context.Context, id string) (*entity.AutoDMRule, error)
	ListByUserIDFn          func(ctx context.Context, userID string) ([]*entity.AutoDMRule, error)
	CountByUserIDFn         func(ctx context.Context, userID string) (int, error)
	CreateFn                func(ctx context.Context, rule *entity.AutoDMRule) error
	UpdateFn                func(ctx context.Context, rule *entity.AutoDMRule) error
	DeleteByIDFn            func(ctx context.Context, id string) error
	ListEnabledFn           func(ctx context.Context) ([]*entity.AutoDMRule, error)
	UpdateLastCheckedFn     func(ctx context.Context, ruleID string, replyID string) error
}

func (m *mockAutoDMRuleRepo) FindByID(ctx context.Context, id string) (*entity.AutoDMRule, error) {
	if m.FindByIDFn != nil {
		return m.FindByIDFn(ctx, id)
	}
	return nil, apperror.NotFound("not found")
}

func (m *mockAutoDMRuleRepo) ListByUserID(ctx context.Context, userID string) ([]*entity.AutoDMRule, error) {
	if m.ListByUserIDFn != nil {
		return m.ListByUserIDFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockAutoDMRuleRepo) CountByUserID(ctx context.Context, userID string) (int, error) {
	if m.CountByUserIDFn != nil {
		return m.CountByUserIDFn(ctx, userID)
	}
	return 0, nil
}

func (m *mockAutoDMRuleRepo) Create(ctx context.Context, rule *entity.AutoDMRule) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, rule)
	}
	rule.ID = "generated-id"
	return nil
}

func (m *mockAutoDMRuleRepo) Update(ctx context.Context, rule *entity.AutoDMRule) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, rule)
	}
	return nil
}

func (m *mockAutoDMRuleRepo) DeleteByID(ctx context.Context, id string) error {
	if m.DeleteByIDFn != nil {
		return m.DeleteByIDFn(ctx, id)
	}
	return nil
}

func (m *mockAutoDMRuleRepo) ListEnabled(ctx context.Context) ([]*entity.AutoDMRule, error) {
	if m.ListEnabledFn != nil {
		return m.ListEnabledFn(ctx)
	}
	return nil, nil
}

func (m *mockAutoDMRuleRepo) UpdateLastCheckedReplyID(ctx context.Context, ruleID string, replyID string) error {
	if m.UpdateLastCheckedFn != nil {
		return m.UpdateLastCheckedFn(ctx, ruleID, replyID)
	}
	return nil
}

type mockDMSentLogRepo struct {
	ListByUserIDFn func(ctx context.Context, userID string, limit int) ([]*entity.DMSentLog, error)
}

func (m *mockDMSentLogRepo) Create(ctx context.Context, log *entity.DMSentLog) error { return nil }
func (m *mockDMSentLogRepo) ExistsByReplyTweetID(ctx context.Context, replyTweetID string, recipientTwitterID string) (bool, error) {
	return false, nil
}
func (m *mockDMSentLogRepo) ListByUserID(ctx context.Context, userID string, limit int) ([]*entity.DMSentLog, error) {
	if m.ListByUserIDFn != nil {
		return m.ListByUserIDFn(ctx, userID, limit)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestMatchKeyword(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		keywords []string
		want     string
	}{
		{
			name:     "exact match",
			text:     "I want to know more about TrendBird",
			keywords: []string{"TrendBird"},
			want:     "TrendBird",
		},
		{
			name:     "case insensitive",
			text:     "tell me about TRENDBIRD please",
			keywords: []string{"trendbird"},
			want:     "trendbird",
		},
		{
			name:     "partial match - first keyword wins",
			text:     "interested in pricing",
			keywords: []string{"pricing", "price"},
			want:     "pricing",
		},
		{
			name:     "substring match",
			text:     "check our price list",
			keywords: []string{"price"},
			want:     "price",
		},
		{
			name:     "no match",
			text:     "hello world",
			keywords: []string{"trendbird", "pricing"},
			want:     "",
		},
		{
			name:     "empty keywords",
			text:     "anything here",
			keywords: []string{},
			want:     "",
		},
		{
			name:     "japanese keywords",
			text:     "料金プランについて教えて",
			keywords: []string{"料金"},
			want:     "料金",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchKeyword(tt.text, tt.keywords)
			if got != tt.want {
				t.Errorf("matchKeyword(%q, %v) = %q, want %q", tt.text, tt.keywords, got, tt.want)
			}
		})
	}
}

func TestListRules(t *testing.T) {
	ctx := context.Background()

	t.Run("pro user succeeds", func(t *testing.T) {
		uc := NewAutoDMUsecase(
			&mockAutoDMRuleRepo{
				ListByUserIDFn: func(_ context.Context, _ string) ([]*entity.AutoDMRule, error) {
					return []*entity.AutoDMRule{}, nil
				},
			},
			&mockDMSentLogRepo{},
		)
		rules, err := uc.ListRules(ctx, "u1")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rules == nil {
			t.Fatal("expected non-nil rules slice")
		}
	})

}

func TestCreateRule(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		uc := NewAutoDMUsecase(
			&mockAutoDMRuleRepo{
				CountByUserIDFn: func(_ context.Context, _ string) (int, error) { return 0, nil },
			},
			&mockDMSentLogRepo{},
		)
		rule, err := uc.CreateRule(ctx, "u1", []string{"pricing"}, "Thanks!")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rule.ID != "generated-id" {
			t.Errorf("expected generated-id, got %q", rule.ID)
		}
		if !rule.Enabled {
			t.Error("expected rule to be enabled by default")
		}
	})

}

func TestUpdateRule_Ownership(t *testing.T) {
	ctx := context.Background()

	t.Run("owner can update", func(t *testing.T) {
		uc := NewAutoDMUsecase(
			&mockAutoDMRuleRepo{
				FindByIDFn: func(_ context.Context, id string) (*entity.AutoDMRule, error) {
					return &entity.AutoDMRule{ID: id, UserID: "u1", TriggerKeywords: []string{}, TemplateMessage: ""}, nil
				},
			},
			&mockDMSentLogRepo{},
		)
		rule, err := uc.UpdateRule(ctx, "u1", "rule-1", true, []string{"pricing"}, "Thanks!")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if rule == nil {
			t.Fatal("expected rule, got nil")
		}
	})

	t.Run("non-owner denied", func(t *testing.T) {
		uc := NewAutoDMUsecase(
			&mockAutoDMRuleRepo{
				FindByIDFn: func(_ context.Context, id string) (*entity.AutoDMRule, error) {
					return &entity.AutoDMRule{ID: id, UserID: "other-user"}, nil
				},
			},
			&mockDMSentLogRepo{},
		)
		_, err := uc.UpdateRule(ctx, "u1", "rule-1", true, []string{"pricing"}, "Thanks!")
		if err == nil {
			t.Fatal("expected permission denied error")
		}
		if !apperror.IsCode(err, apperror.CodePermissionDenied) {
			t.Errorf("expected CodePermissionDenied, got %v", err)
		}
	})
}

func TestDeleteRule_Ownership(t *testing.T) {
	ctx := context.Background()

	t.Run("owner can delete", func(t *testing.T) {
		var deletedID string
		uc := NewAutoDMUsecase(
			&mockAutoDMRuleRepo{
				FindByIDFn: func(_ context.Context, id string) (*entity.AutoDMRule, error) {
					return &entity.AutoDMRule{ID: id, UserID: "u1"}, nil
				},
				DeleteByIDFn: func(_ context.Context, id string) error {
					deletedID = id
					return nil
				},
			},
			&mockDMSentLogRepo{},
		)
		err := uc.DeleteRule(ctx, "u1", "rule-1")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if deletedID != "rule-1" {
			t.Errorf("expected rule-1 to be deleted, got %q", deletedID)
		}
	})

	t.Run("non-owner denied", func(t *testing.T) {
		uc := NewAutoDMUsecase(
			&mockAutoDMRuleRepo{
				FindByIDFn: func(_ context.Context, id string) (*entity.AutoDMRule, error) {
					return &entity.AutoDMRule{ID: id, UserID: "other-user"}, nil
				},
			},
			&mockDMSentLogRepo{},
		)
		err := uc.DeleteRule(ctx, "u1", "rule-1")
		if err == nil {
			t.Fatal("expected permission denied error")
		}
		if !apperror.IsCode(err, apperror.CodePermissionDenied) {
			t.Errorf("expected CodePermissionDenied, got %v", err)
		}
	})
}

func TestMatchRules(t *testing.T) {
	rules := []*entity.AutoDMRule{
		{ID: "r1", TriggerKeywords: []string{"pricing", "price"}},
		{ID: "r2", TriggerKeywords: []string{"demo", "trial"}},
	}

	t.Run("first rule matches", func(t *testing.T) {
		ruleID, kw := matchRules("tell me about pricing", rules)
		if ruleID != "r1" || kw != "pricing" {
			t.Errorf("expected r1/pricing, got %s/%s", ruleID, kw)
		}
	})

	t.Run("second rule matches", func(t *testing.T) {
		ruleID, kw := matchRules("can I get a demo?", rules)
		if ruleID != "r2" || kw != "demo" {
			t.Errorf("expected r2/demo, got %s/%s", ruleID, kw)
		}
	})

	t.Run("no match", func(t *testing.T) {
		ruleID, kw := matchRules("hello world", rules)
		if ruleID != "" || kw != "" {
			t.Errorf("expected empty, got %s/%s", ruleID, kw)
		}
	})

	t.Run("first rule wins on overlap", func(t *testing.T) {
		// Both rules could match if keywords overlap
		overlapping := []*entity.AutoDMRule{
			{ID: "r1", TriggerKeywords: []string{"test"}},
			{ID: "r2", TriggerKeywords: []string{"test"}},
		}
		ruleID, _ := matchRules("this is a test", overlapping)
		if ruleID != "r1" {
			t.Errorf("expected r1 (first rule wins), got %s", ruleID)
		}
	})
}

func TestMinSinceID(t *testing.T) {
	id1 := "100"
	id2 := "200"
	id3 := "50"

	t.Run("returns nil if any rule has no since_id", func(t *testing.T) {
		rules := []*entity.AutoDMRule{
			{LastCheckedReplyID: &id1},
			{LastCheckedReplyID: nil},
		}
		result := minSinceID(rules)
		if result != nil {
			t.Errorf("expected nil, got %v", *result)
		}
	})

	t.Run("returns min across rules", func(t *testing.T) {
		rules := []*entity.AutoDMRule{
			{LastCheckedReplyID: &id1},
			{LastCheckedReplyID: &id2},
			{LastCheckedReplyID: &id3},
		}
		result := minSinceID(rules)
		if result == nil || *result != "100" {
			// "100" < "200" and "100" > "50" in string comparison, so min is "50"... wait
			// Actually string comparison: "100" < "200" < "50" (lexicographic)
			// So min is "100"
			t.Errorf("expected 100, got %v", result)
		}
	})

	t.Run("single rule", func(t *testing.T) {
		rules := []*entity.AutoDMRule{
			{LastCheckedReplyID: &id2},
		}
		result := minSinceID(rules)
		if result == nil || *result != "200" {
			t.Errorf("expected 200, got %v", result)
		}
	})

	t.Run("all nil", func(t *testing.T) {
		rules := []*entity.AutoDMRule{
			{LastCheckedReplyID: nil},
		}
		result := minSinceID(rules)
		if result != nil {
			t.Errorf("expected nil, got %v", *result)
		}
	})
}

func TestGetSentLogs(t *testing.T) {
	ctx := context.Background()

	logs := []*entity.DMSentLog{
		{ID: "log-1", UserID: "u1", TriggerKeyword: "pricing", DMText: "Hello!"},
		{ID: "log-2", UserID: "u1", TriggerKeyword: "demo", DMText: "Thanks!"},
	}

	uc := NewAutoDMUsecase(
		&mockAutoDMRuleRepo{},
		&mockDMSentLogRepo{
			ListByUserIDFn: func(_ context.Context, _ string, limit int) ([]*entity.DMSentLog, error) {
				if limit < len(logs) {
					return logs[:limit], nil
				}
				return logs, nil
			},
		},
	)

	result, err := uc.GetSentLogs(ctx, "u1", 20)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 logs, got %d", len(result))
	}
}
