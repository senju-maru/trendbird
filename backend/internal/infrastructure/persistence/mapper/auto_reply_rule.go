package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// AutoReplyRuleToEntity converts a GORM AutoReplyRule model to a domain entity.
func AutoReplyRuleToEntity(m *model.AutoReplyRule) *entity.AutoReplyRule {
	keywords := make([]string, len(m.TriggerKeywords))
	copy(keywords, m.TriggerKeywords)
	return &entity.AutoReplyRule{
		ID:                 m.ID,
		UserID:             m.UserID,
		Enabled:            m.Enabled,
		TargetTweetID:      m.TargetTweetID,
		TargetTweetText:    m.TargetTweetText,
		TriggerKeywords:    keywords,
		ReplyTemplate:      m.ReplyTemplate,
		LastCheckedReplyID: m.LastCheckedReplyID,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
}

// AutoReplyRuleToModel converts a domain AutoReplyRule entity to a GORM model.
func AutoReplyRuleToModel(e *entity.AutoReplyRule) *model.AutoReplyRule {
	keywords := make(model.StringArray, len(e.TriggerKeywords))
	copy(keywords, e.TriggerKeywords)
	return &model.AutoReplyRule{
		ID:                 e.ID,
		UserID:             e.UserID,
		Enabled:            e.Enabled,
		TargetTweetID:      e.TargetTweetID,
		TargetTweetText:    e.TargetTweetText,
		TriggerKeywords:    keywords,
		ReplyTemplate:      e.ReplyTemplate,
		LastCheckedReplyID: e.LastCheckedReplyID,
		CreatedAt:          e.CreatedAt,
		UpdatedAt:          e.UpdatedAt,
	}
}
