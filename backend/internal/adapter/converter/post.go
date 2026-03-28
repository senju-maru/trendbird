package converter

import (
	"time"

	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/internal/domain/entity"
)

// styleLabels は PostStyle ごとの日本語ラベル。
var styleLabels = map[entity.PostStyle]string{
	entity.PostStyleCasual:   "カジュアル",
	entity.PostStyleBreaking: "速報",
	entity.PostStyleAnalysis: "分析",
}

// styleIcons は PostStyle ごとのアイコン名。
var styleIcons = map[entity.PostStyle]string{
	entity.PostStyleCasual:   "smile",
	entity.PostStyleBreaking: "zap",
	entity.PostStyleAnalysis: "bar-chart",
}

// GeneratedPostToProto は entity.GeneratedPost を Proto GeneratedPost に変換する。
func GeneratedPostToProto(e *entity.GeneratedPost) *trendbirdv1.GeneratedPost {
	topicID := ""
	if e.TopicID != nil {
		topicID = *e.TopicID
	}

	return &trendbirdv1.GeneratedPost{
		Id:         e.ID,
		Style:      trendbirdv1.PostStyle(e.Style),
		StyleLabel: styleLabels[e.Style],
		StyleIcon:  styleIcons[e.Style],
		Content:    e.Content,
		TopicId:    topicID,
	}
}

// GeneratedPostSliceToProto は GeneratedPost スライスを Proto スライスに変換する。
func GeneratedPostSliceToProto(es []*entity.GeneratedPost) []*trendbirdv1.GeneratedPost {
	result := make([]*trendbirdv1.GeneratedPost, len(es))
	for i, e := range es {
		result[i] = GeneratedPostToProto(e)
	}
	return result
}

// PostToScheduledPostProto は entity.Post を Proto ScheduledPost に変換する。
func PostToScheduledPostProto(e *entity.Post) *trendbirdv1.ScheduledPost {
	return &trendbirdv1.ScheduledPost{
		Id:             e.ID,
		Content:        e.Content,
		TopicId:        e.TopicID,
		TopicName:      e.TopicName,
		Status:         trendbirdv1.PostStatus(e.Status),
		ScheduledAt:    timeToOptionalString(e.ScheduledAt),
		PublishedAt:    timeToOptionalString(e.PublishedAt),
		FailedAt:       timeToOptionalString(e.FailedAt),
		ErrorMessage:   e.ErrorMessage,
		CreatedAt:      e.CreatedAt.Format(time.RFC3339),
		UpdatedAt:      e.UpdatedAt.Format(time.RFC3339),
		CharacterCount: int32(len([]rune(e.Content))),
	}
}

// PostToPostHistoryProto は entity.Post を Proto PostHistory に変換する。
func PostToPostHistoryProto(e *entity.Post) *trendbirdv1.PostHistory {
	publishedAt := ""
	if e.PublishedAt != nil {
		publishedAt = e.PublishedAt.Format(time.RFC3339)
	}

	return &trendbirdv1.PostHistory{
		Id:          e.ID,
		Content:     e.Content,
		TopicId:     e.TopicID,
		TopicName:   e.TopicName,
		PublishedAt: publishedAt,
		Likes:       e.Likes,
		Retweets:    e.Retweets,
		Replies:     e.Replies,
		Views:       e.Views,
		TweetUrl:    e.TweetURL,
	}
}

// PostSliceToScheduledPostProto は Post スライスを ScheduledPost Proto スライスに変換する。
func PostSliceToScheduledPostProto(es []*entity.Post) []*trendbirdv1.ScheduledPost {
	result := make([]*trendbirdv1.ScheduledPost, len(es))
	for i, e := range es {
		result[i] = PostToScheduledPostProto(e)
	}
	return result
}

// PostSliceToPostHistoryProto は Post スライスを PostHistory Proto スライスに変換する。
func PostSliceToPostHistoryProto(es []*entity.Post) []*trendbirdv1.PostHistory {
	result := make([]*trendbirdv1.PostHistory, len(es))
	for i, e := range es {
		result[i] = PostToPostHistoryProto(e)
	}
	return result
}
