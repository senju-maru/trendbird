package converter

import (
	"time"

	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/internal/domain/entity"
)

// ActivityToProto は entity.Activity を Proto Activity に変換する。
func ActivityToProto(e *entity.Activity) *trendbirdv1.Activity {
	return &trendbirdv1.Activity{
		Id:          e.ID,
		Type:        trendbirdv1.ActivityType(e.Type),
		TopicName:   e.TopicName,
		Description: e.Description,
		Timestamp:   e.Timestamp.Format(time.RFC3339),
	}
}

// ActivitySliceToProto は Activity スライスを Proto スライスに変換する。
func ActivitySliceToProto(es []*entity.Activity) []*trendbirdv1.Activity {
	result := make([]*trendbirdv1.Activity, len(es))
	for i, e := range es {
		result[i] = ActivityToProto(e)
	}
	return result
}

// DashboardStatsToProto は検知数・生成数・最終チェック時刻から Proto DashboardStats を生成する。
func DashboardStatsToProto(detections, generations int32, lastCheckedAt *time.Time) *trendbirdv1.DashboardStats {
	stats := &trendbirdv1.DashboardStats{
		Detections:  detections,
		Generations: generations,
	}
	if lastCheckedAt != nil {
		s := lastCheckedAt.Format(time.RFC3339)
		stats.LastCheckedAt = &s
	}
	return stats
}
