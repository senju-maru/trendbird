package converter

import (
	"time"

	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/internal/domain/entity"
)

// AutoReplyRuleToProto converts a domain AutoReplyRule entity to a proto message.
func AutoReplyRuleToProto(e *entity.AutoReplyRule) *trendbirdv1.AutoReplyRule {
	return &trendbirdv1.AutoReplyRule{
		Id:              e.ID,
		Enabled:         e.Enabled,
		TargetTweetId:   e.TargetTweetID,
		TargetTweetText: e.TargetTweetText,
		TriggerKeywords: e.TriggerKeywords,
		ReplyTemplate:   e.ReplyTemplate,
		CreatedAt:       e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       e.UpdatedAt.Format(time.RFC3339),
	}
}

// AutoReplyRuleSliceToProto converts a slice of domain AutoReplyRule entities to proto messages.
func AutoReplyRuleSliceToProto(es []*entity.AutoReplyRule) []*trendbirdv1.AutoReplyRule {
	protos := make([]*trendbirdv1.AutoReplyRule, len(es))
	for i, e := range es {
		protos[i] = AutoReplyRuleToProto(e)
	}
	return protos
}

// ReplySentLogToProto converts a domain ReplySentLog entity to a proto message.
func ReplySentLogToProto(e *entity.ReplySentLog) *trendbirdv1.ReplySentLog {
	return &trendbirdv1.ReplySentLog{
		Id:               e.ID,
		OriginalTweetId:  e.OriginalTweetID,
		OriginalAuthorId: e.OriginalAuthorID,
		ReplyTweetId:     e.ReplyTweetID,
		TriggerKeyword:   e.TriggerKeyword,
		ReplyText:        e.ReplyText,
		SentAt:           e.SentAt.Format(time.RFC3339),
	}
}

// ReplySentLogSliceToProto converts a slice of domain ReplySentLog entities to proto messages.
func ReplySentLogSliceToProto(es []*entity.ReplySentLog) []*trendbirdv1.ReplySentLog {
	protos := make([]*trendbirdv1.ReplySentLog, len(es))
	for i, e := range es {
		protos[i] = ReplySentLogToProto(e)
	}
	return protos
}
