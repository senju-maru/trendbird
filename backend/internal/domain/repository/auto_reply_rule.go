package repository

import (
	"context"

	"github.com/trendbird/backend/internal/domain/entity"
)

// AutoReplyRuleRepository defines the persistence operations for AutoReplyRule entities.
type AutoReplyRuleRepository interface {
	// FindByID returns the auto reply rule by its ID. Returns NotFound if not exists.
	FindByID(ctx context.Context, id string) (*entity.AutoReplyRule, error)
	// ListByUserID returns all auto reply rules for the given user.
	ListByUserID(ctx context.Context, userID string) ([]*entity.AutoReplyRule, error)
	// Create creates a new auto reply rule.
	Create(ctx context.Context, rule *entity.AutoReplyRule) error
	// Update updates an existing auto reply rule.
	Update(ctx context.Context, rule *entity.AutoReplyRule) error
	// DeleteByID deletes the auto reply rule by its ID.
	DeleteByID(ctx context.Context, id string) error
	// ListEnabled returns all enabled auto reply rules.
	ListEnabled(ctx context.Context) ([]*entity.AutoReplyRule, error)
	// UpdateLastCheckedReplyID updates the last_checked_reply_id for the given rule.
	UpdateLastCheckedReplyID(ctx context.Context, ruleID string, replyID string) error
}
