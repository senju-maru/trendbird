package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// ReplySentLogToEntity converts a GORM ReplySentLog model to a domain entity.
func ReplySentLogToEntity(m *model.ReplySentLog) *entity.ReplySentLog {
	return &entity.ReplySentLog{
		ID:               m.ID,
		UserID:           m.UserID,
		RuleID:           m.RuleID,
		OriginalTweetID:  m.OriginalTweetID,
		OriginalAuthorID: m.OriginalAuthorID,
		ReplyTweetID:     m.ReplyTweetID,
		TriggerKeyword:   m.TriggerKeyword,
		ReplyText:        m.ReplyText,
		SentAt:           m.SentAt,
	}
}

// ReplySentLogToModel converts a domain ReplySentLog entity to a GORM model.
func ReplySentLogToModel(e *entity.ReplySentLog) *model.ReplySentLog {
	return &model.ReplySentLog{
		ID:               e.ID,
		UserID:           e.UserID,
		RuleID:           e.RuleID,
		OriginalTweetID:  e.OriginalTweetID,
		OriginalAuthorID: e.OriginalAuthorID,
		ReplyTweetID:     e.ReplyTweetID,
		TriggerKeyword:   e.TriggerKeyword,
		ReplyText:        e.ReplyText,
		SentAt:           e.SentAt,
	}
}
