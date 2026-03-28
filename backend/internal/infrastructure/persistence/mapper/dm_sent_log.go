package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// DMSentLogToEntity converts a GORM DMSentLog model to a domain entity.
func DMSentLogToEntity(m *model.DMSentLog) *entity.DMSentLog {
	return &entity.DMSentLog{
		ID:                 m.ID,
		UserID:             m.UserID,
		RuleID:             m.RuleID,
		RecipientTwitterID: m.RecipientTwitterID,
		ReplyTweetID:       m.ReplyTweetID,
		TriggerKeyword:     m.TriggerKeyword,
		DMText:             m.DMText,
		SentAt:             m.SentAt,
	}
}

// DMSentLogToModel converts a domain DMSentLog entity to a GORM model.
func DMSentLogToModel(e *entity.DMSentLog) *model.DMSentLog {
	return &model.DMSentLog{
		ID:                 e.ID,
		UserID:             e.UserID,
		RuleID:             e.RuleID,
		RecipientTwitterID: e.RecipientTwitterID,
		ReplyTweetID:       e.ReplyTweetID,
		TriggerKeyword:     e.TriggerKeyword,
		DMText:             e.DMText,
		SentAt:             e.SentAt,
	}
}
