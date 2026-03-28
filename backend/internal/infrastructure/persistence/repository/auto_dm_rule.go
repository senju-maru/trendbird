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

var _ domainrepo.AutoDMRuleRepository = (*autoDMRuleRepository)(nil)

type autoDMRuleRepository struct {
	db *gorm.DB
}

func NewAutoDMRuleRepository(db *gorm.DB) *autoDMRuleRepository {
	return &autoDMRuleRepository{db: db}
}

func (r *autoDMRuleRepository) getDB(ctx context.Context) *gorm.DB {
	return persistence.GetDB(ctx, r.db).WithContext(ctx)
}

func (r *autoDMRuleRepository) FindByID(ctx context.Context, id string) (*entity.AutoDMRule, error) {
	var m model.AutoDMRule
	if err := r.getDB(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NotFound("auto dm rule not found")
		}
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to find auto dm rule", err)
	}
	return mapper.AutoDMRuleToEntity(&m), nil
}

func (r *autoDMRuleRepository) ListByUserID(ctx context.Context, userID string) ([]*entity.AutoDMRule, error) {
	var models []model.AutoDMRule
	if err := r.getDB(ctx).Where("user_id = ?", userID).Order("created_at ASC").Find(&models).Error; err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list auto dm rules", err)
	}
	entities := make([]*entity.AutoDMRule, len(models))
	for i := range models {
		entities[i] = mapper.AutoDMRuleToEntity(&models[i])
	}
	return entities, nil
}

func (r *autoDMRuleRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	var count int64
	if err := r.getDB(ctx).Model(&model.AutoDMRule{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return 0, apperror.Wrap(apperror.CodeInternal, "failed to count auto dm rules", err)
	}
	return int(count), nil
}

func (r *autoDMRuleRepository) Create(ctx context.Context, rule *entity.AutoDMRule) error {
	m := mapper.AutoDMRuleToModel(rule)
	if err := r.getDB(ctx).Create(m).Error; err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to create auto dm rule", err)
	}
	rule.ID = m.ID
	rule.CreatedAt = m.CreatedAt
	rule.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *autoDMRuleRepository) Update(ctx context.Context, rule *entity.AutoDMRule) error {
	result := r.getDB(ctx).Model(&model.AutoDMRule{}).Where("id = ?", rule.ID).Updates(map[string]any{
		"enabled":          rule.Enabled,
		"trigger_keywords": model.StringArray(rule.TriggerKeywords),
		"template_message": rule.TemplateMessage,
	})
	if result.Error != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to update auto dm rule", result.Error)
	}
	return nil
}

func (r *autoDMRuleRepository) DeleteByID(ctx context.Context, id string) error {
	result := r.getDB(ctx).Where("id = ?", id).Delete(&model.AutoDMRule{})
	if result.Error != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to delete auto dm rule", result.Error)
	}
	return nil
}

func (r *autoDMRuleRepository) ListEnabled(ctx context.Context) ([]*entity.AutoDMRule, error) {
	var models []model.AutoDMRule
	if err := r.getDB(ctx).Where("enabled = true").Find(&models).Error; err != nil {
		return nil, apperror.Wrap(apperror.CodeInternal, "failed to list enabled auto dm rules", err)
	}
	entities := make([]*entity.AutoDMRule, len(models))
	for i := range models {
		entities[i] = mapper.AutoDMRuleToEntity(&models[i])
	}
	return entities, nil
}

func (r *autoDMRuleRepository) UpdateLastCheckedReplyID(ctx context.Context, ruleID string, replyID string) error {
	result := r.getDB(ctx).Exec(
		"UPDATE auto_dm_rules SET last_checked_reply_id = ?, updated_at = NOW() WHERE id = ?",
		replyID, ruleID,
	)
	if result.Error != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to update last checked reply id", result.Error)
	}
	return nil
}
