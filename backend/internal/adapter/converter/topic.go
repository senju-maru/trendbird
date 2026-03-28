package converter

import (
	"time"

	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/internal/domain/entity"
)

// TopicToProtoInput は Topic の複合変換に必要なデータを集約する。
type TopicToProtoInput struct {
	Topic               *entity.Topic
	SparklineData       []*entity.TopicVolume
	WeeklySparklineData []*entity.TopicVolume
	SpikeHistory        []*entity.SpikeHistory
	PostingTip          *entity.PostingTip
}

// TopicToProto は TopicToProtoInput を Proto Topic に変換する（詳細表示用）。
func TopicToProto(input *TopicToProtoInput) *trendbirdv1.Topic {
	e := input.Topic

	t := &trendbirdv1.Topic{
		Id:                  e.ID,
		Name:                e.Name,
		Keywords:            e.Keywords,
		Genre:               e.GenreSlug,
		Status:              trendbirdv1.TopicStatus(e.Status),
		ChangePercent:       e.ChangePercent,
		ZScore:              e.ZScore,
		CurrentVolume:       e.CurrentVolume,
		BaselineVolume:      e.BaselineVolume,
		SparklineData:       topicVolumesToSparkline(input.SparklineData),
		Context:             e.Context,
		CreatedAt:           e.CreatedAt.Format(time.RFC3339),
		ContextSummary:      e.ContextSummary,
		SpikeStartedAt:      timeToOptionalString(e.SpikeStartedAt),
		WeeklySparklineData: topicVolumesToSparkline(input.WeeklySparklineData),
		SpikeHistory:        spikeHistoryToProto(input.SpikeHistory),
		PostingTips:         postingTipToProto(input.PostingTip),
		NotificationEnabled: e.NotificationEnabled,
	}

	return t
}

// TopicToProtoSimple は entity.Topic を Proto Topic に変換する（リスト表示用、関連データなし）。
func TopicToProtoSimple(e *entity.Topic) *trendbirdv1.Topic {
	return &trendbirdv1.Topic{
		Id:                  e.ID,
		Name:                e.Name,
		Keywords:            e.Keywords,
		Genre:               e.GenreSlug,
		Status:              trendbirdv1.TopicStatus(e.Status),
		ChangePercent:       e.ChangePercent,
		ZScore:              e.ZScore,
		CurrentVolume:       e.CurrentVolume,
		BaselineVolume:      e.BaselineVolume,
		Context:             e.Context,
		CreatedAt:           e.CreatedAt.Format(time.RFC3339),
		ContextSummary:      e.ContextSummary,
		SpikeStartedAt:      timeToOptionalString(e.SpikeStartedAt),
		NotificationEnabled: e.NotificationEnabled,
	}
}

// TopicSliceToProtoSimple は Topic スライスを Proto スライスに変換する（リスト表示用）。
func TopicSliceToProtoSimple(es []*entity.Topic) []*trendbirdv1.Topic {
	result := make([]*trendbirdv1.Topic, len(es))
	for i, e := range es {
		result[i] = TopicToProtoSimple(e)
	}
	return result
}

// TopicSuggestionToProto は TopicSuggestion を Proto TopicSuggestion に変換する。
func TopicSuggestionToProto(e *entity.TopicSuggestion) *trendbirdv1.TopicSuggestion {
	return &trendbirdv1.TopicSuggestion{
		Id:              e.ID,
		Name:            e.Name,
		Keywords:        e.Keywords,
		Genre:           e.GenreSlug,
		GenreLabel:      e.GenreLabel,
		SimilarityScore: e.SimilarityScore,
	}
}

// TopicSuggestionSliceToProto は TopicSuggestion スライスを Proto スライスに変換する。
func TopicSuggestionSliceToProto(es []*entity.TopicSuggestion) []*trendbirdv1.TopicSuggestion {
	result := make([]*trendbirdv1.TopicSuggestion, len(es))
	for i, e := range es {
		result[i] = TopicSuggestionToProto(e)
	}
	return result
}

func topicVolumesToSparkline(volumes []*entity.TopicVolume) []*trendbirdv1.SparklineDataPoint {
	if volumes == nil {
		return nil
	}
	result := make([]*trendbirdv1.SparklineDataPoint, len(volumes))
	for i, v := range volumes {
		result[i] = &trendbirdv1.SparklineDataPoint{
			Timestamp: v.Timestamp.Format(time.RFC3339),
			Value:     v.Value,
		}
	}
	return result
}

func spikeHistoryToProto(entries []*entity.SpikeHistory) []*trendbirdv1.SpikeHistoryEntry {
	if entries == nil {
		return nil
	}
	result := make([]*trendbirdv1.SpikeHistoryEntry, len(entries))
	for i, e := range entries {
		result[i] = &trendbirdv1.SpikeHistoryEntry{
			Id:              e.ID,
			Timestamp:       e.Timestamp.Format(time.RFC3339),
			PeakZScore:      e.PeakZScore,
			Status:          trendbirdv1.TopicStatus(e.Status),
			Summary:         e.Summary,
			DurationMinutes: e.DurationMinutes,
		}
	}
	return result
}

func postingTipToProto(tip *entity.PostingTip) *trendbirdv1.PostingTips {
	if tip == nil {
		return nil
	}
	return &trendbirdv1.PostingTips{
		PeakDays:          tip.PeakDays,
		PeakHoursStart:    tip.PeakHoursStart,
		PeakHoursEnd:      tip.PeakHoursEnd,
		NextSuggestedTime: tip.NextSuggestedTime.Format(time.RFC3339),
	}
}
