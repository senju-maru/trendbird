package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// AutoDMRuleToEntity converts a GORM AutoDMRule model to a domain entity.
func AutoDMRuleToEntity(m *model.AutoDMRule) *entity.AutoDMRule {
	keywords := make([]string, len(m.TriggerKeywords))
	copy(keywords, m.TriggerKeywords)
	return &entity.AutoDMRule{
		ID:                 m.ID,
		UserID:             m.UserID,
		Enabled:            m.Enabled,
		TriggerKeywords:    keywords,
		TemplateMessage:    m.TemplateMessage,
		LastCheckedReplyID: m.LastCheckedReplyID,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
}

// AutoDMRuleToModel converts a domain AutoDMRule entity to a GORM model.
func AutoDMRuleToModel(e *entity.AutoDMRule) *model.AutoDMRule {
	keywords := make(model.StringArray, len(e.TriggerKeywords))
	copy(keywords, e.TriggerKeywords)
	return &model.AutoDMRule{
		ID:                 e.ID,
		UserID:             e.UserID,
		Enabled:            e.Enabled,
		TriggerKeywords:    keywords,
		TemplateMessage:    e.TemplateMessage,
		LastCheckedReplyID: e.LastCheckedReplyID,
		CreatedAt:          e.CreatedAt,
		UpdatedAt:          e.UpdatedAt,
	}
}
