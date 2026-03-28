package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// DMPendingQueueToEntity converts a GORM DMPendingQueue model to a domain entity.
func DMPendingQueueToEntity(m *model.DMPendingQueue) *entity.DMPendingQueue {
	return &entity.DMPendingQueue{
		ID:                 m.ID,
		UserID:             m.UserID,
		RuleID:             m.RuleID,
		RecipientTwitterID: m.RecipientTwitterID,
		ReplyTweetID:       m.ReplyTweetID,
		TriggerKeyword:     m.TriggerKeyword,
		Status:             entity.DMPendingStatus(m.Status),
		CreatedAt:          m.CreatedAt,
	}
}

// DMPendingQueueToModel converts a domain DMPendingQueue entity to a GORM model.
func DMPendingQueueToModel(e *entity.DMPendingQueue) *model.DMPendingQueue {
	return &model.DMPendingQueue{
		ID:                 e.ID,
		UserID:             e.UserID,
		RuleID:             e.RuleID,
		RecipientTwitterID: e.RecipientTwitterID,
		ReplyTweetID:       e.ReplyTweetID,
		TriggerKeyword:     e.TriggerKeyword,
		Status:             int(e.Status),
		CreatedAt:          e.CreatedAt,
	}
}
