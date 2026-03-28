package repository

import (
	"context"

	"github.com/trendbird/backend/internal/domain/entity"
)

// AutoDMRuleRepository defines the persistence operations for AutoDMRule entities.
type AutoDMRuleRepository interface {
	// FindByID returns the auto DM rule by its ID. Returns NotFound if not exists.
	FindByID(ctx context.Context, id string) (*entity.AutoDMRule, error)
	// ListByUserID returns all auto DM rules for the given user.
	ListByUserID(ctx context.Context, userID string) ([]*entity.AutoDMRule, error)
	// CountByUserID returns the number of auto DM rules for the given user.
	CountByUserID(ctx context.Context, userID string) (int, error)
	// Create creates a new auto DM rule.
	Create(ctx context.Context, rule *entity.AutoDMRule) error
	// Update updates an existing auto DM rule.
	Update(ctx context.Context, rule *entity.AutoDMRule) error
	// DeleteByID deletes the auto DM rule by its ID.
	DeleteByID(ctx context.Context, id string) error
	// ListEnabled returns all enabled auto DM rules.
	ListEnabled(ctx context.Context) ([]*entity.AutoDMRule, error)
	// UpdateLastCheckedReplyID updates the last_checked_reply_id for the given rule.
	UpdateLastCheckedReplyID(ctx context.Context, ruleID string, replyID string) error
}
