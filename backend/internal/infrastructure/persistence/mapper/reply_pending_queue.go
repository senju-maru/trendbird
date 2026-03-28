package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// ReplyPendingQueueToEntity converts a GORM ReplyPendingQueue model to a domain entity.
func ReplyPendingQueueToEntity(m *model.ReplyPendingQueue) *entity.ReplyPendingQueue {
	return &entity.ReplyPendingQueue{
		ID:               m.ID,
		UserID:           m.UserID,
		RuleID:           m.RuleID,
		OriginalTweetID:  m.OriginalTweetID,
		OriginalAuthorID: m.OriginalAuthorID,
		TriggerKeyword:   m.TriggerKeyword,
		Status:           entity.ReplyPendingStatus(m.Status),
		CreatedAt:        m.CreatedAt,
	}
}

// ReplyPendingQueueToModel converts a domain ReplyPendingQueue entity to a GORM model.
func ReplyPendingQueueToModel(e *entity.ReplyPendingQueue) *model.ReplyPendingQueue {
	return &model.ReplyPendingQueue{
		ID:               e.ID,
		UserID:           e.UserID,
		RuleID:           e.RuleID,
		OriginalTweetID:  e.OriginalTweetID,
		OriginalAuthorID: e.OriginalAuthorID,
		TriggerKeyword:   e.TriggerKeyword,
		Status:           int(e.Status),
		CreatedAt:        e.CreatedAt,
	}
}
