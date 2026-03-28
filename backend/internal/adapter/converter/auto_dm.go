package converter

import (
	"time"

	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/internal/domain/entity"
)

// AutoDMRuleToProto converts a domain AutoDMRule entity to a proto message.
func AutoDMRuleToProto(e *entity.AutoDMRule) *trendbirdv1.AutoDMRule {
	return &trendbirdv1.AutoDMRule{
		Id:              e.ID,
		Enabled:         e.Enabled,
		TriggerKeywords: e.TriggerKeywords,
		TemplateMessage: e.TemplateMessage,
		CreatedAt:       e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       e.UpdatedAt.Format(time.RFC3339),
	}
}

// AutoDMRuleSliceToProto converts a slice of domain AutoDMRule entities to proto messages.
func AutoDMRuleSliceToProto(es []*entity.AutoDMRule) []*trendbirdv1.AutoDMRule {
	protos := make([]*trendbirdv1.AutoDMRule, len(es))
	for i, e := range es {
		protos[i] = AutoDMRuleToProto(e)
	}
	return protos
}

// DMSentLogToProto converts a domain DMSentLog entity to a proto message.
func DMSentLogToProto(e *entity.DMSentLog) *trendbirdv1.DMSentLog {
	return &trendbirdv1.DMSentLog{
		Id:                 e.ID,
		RecipientTwitterId: e.RecipientTwitterID,
		ReplyTweetId:       e.ReplyTweetID,
		TriggerKeyword:     e.TriggerKeyword,
		DmText:             e.DMText,
		SentAt:             e.SentAt.Format(time.RFC3339),
	}
}

// DMSentLogSliceToProto converts a slice of domain DMSentLog entities to proto messages.
func DMSentLogSliceToProto(es []*entity.DMSentLog) []*trendbirdv1.DMSentLog {
	protos := make([]*trendbirdv1.DMSentLog, len(es))
	for i, e := range es {
		protos[i] = DMSentLogToProto(e)
	}
	return protos
}
