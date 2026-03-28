package repository

import (
	"context"
	"errors"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	domainrepo "github.com/trendbird/backend/internal/domain/repository"
	"github.com/trendbird/backend/internal/infrastructure/persistence"
	"github.com/trendbird/backend/internal/infrastructure/persistence/mapper"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
	"gorm.io/gorm"
)

var _ domainrepo.AutoReplyRuleRepository = (*autoReplyRuleRepository)(nil)

type autoReplyRuleRepository struct {
	db *gorm.DB
}

func NewAutoReplyRuleRepository(db *gorm.DB) *autoReplyRuleRepository {
	return &autoReplyRuleRepository{db: db}
}

func (r *autoReplyRuleRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

func (r *autoReplyRuleRepository) FindByID(ctx context.Context, id string) (*entity.AutoReplyRule, error) {
	var m model.AutoReplyRule
	if err := r.getDB(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("auto reply rule not found")
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to find auto reply rule", err)
	}
	return mapper.AutoReplyRuleToEntity(&m), nil
}

func (r *autoReplyRuleRepository) ListByUserID(ctx context.Context, userID string) ([]*entity.AutoReplyRule, error) {
	var models []model.AutoReplyRule
	if err := r.getDB(ctx).Where("user_id = ?", userID).Order("created_at ASC").Find(&models).Error; err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list auto reply rules", err)
	}
	entities := make([]*entity.AutoReplyRule, len(models))
	for i := range models {
		entities[i] = mapper.AutoReplyRuleToEntity(&models[i])
	}
	return entities, nil
}

func (r *autoReplyRuleRepository) Create(ctx context.Context, rule *entity.AutoReplyRule) error {
	m := mapper.AutoReplyRuleToModel(rule)
	if err := r.getDB(ctx).Create(m).Error; err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to create auto reply rule", err)
	}
	rule.ID = m.ID
	rule.CreatedAt = m.CreatedAt
	rule.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *autoReplyRuleRepository) Update(ctx context.Context, rule *entity.AutoReplyRule) error {
	result := r.getDB(ctx).Model(&model.AutoReplyRule{}).Where("id = ?", rule.ID).Updates(map[string]any{
		"enabled":          rule.Enabled,
		"trigger_keywords": model.StringArray(rule.TriggerKeywords),
		"reply_template":   rule.ReplyTemplate,
	})
	if result.Error != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to update auto reply rule", result.Error)
	}
	return nil
}

func (r *autoReplyRuleRepository) DeleteByID(ctx context.Context, id string) error {
	result := r.getDB(ctx).Where("id = ?", id).Delete(&model.AutoReplyRule{})
	if result.Error != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to delete auto reply rule", result.Error)
	}
	return nil
}

func (r *autoReplyRuleRepository) ListEnabled(ctx context.Context) ([]*entity.AutoReplyRule, error) {
	var models []model.AutoReplyRule
	if err := r.getDB(ctx).Where("enabled = true").Find(&models).Error; err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list enabled auto reply rules", err)
	}
	entities := make([]*entity.AutoReplyRule, len(models))
	for i := range models {
		entities[i] = mapper.AutoReplyRuleToEntity(&models[i])
	}
	return entities, nil
}

func (r *autoReplyRuleRepository) UpdateLastCheckedReplyID(ctx context.Context, ruleID string, replyID string) error {
	result := r.getDB(ctx).Exec(
		"UPDATE auto_reply_rules SET last_checked_reply_id = ?, updated_at = NOW() WHERE id = ?",
		replyID, ruleID,
	)
	if result.Error != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to update last checked reply id", result.Error)
	}
	return nil
}
