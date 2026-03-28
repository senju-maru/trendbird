package entity

import "time"

// TopicStatus represents the current trend status of a topic.
type TopicStatus int

const (
	TopicSpike  TopicStatus = 1
	TopicRising TopicStatus = 2
	TopicStable TopicStatus = 3
)

// Topic represents a monitored trend topic.
// Genre は genres マスターテーブルの slug で、GenreID は FK。
// GenreSlug は JOIN/Preload で取得され、API レスポンスの genre フィールドに設定される。
type Topic struct {
	ID                  string
	Name                string
	Keywords            []string
	GenreID             string
	GenreSlug           string // genres テーブルから JOIN で取得
	Status              TopicStatus
	ChangePercent       float64
	ZScore              *float64
	CurrentVolume       int32
	BaselineVolume      int32
	Context             *string
	ContextSummary      *string
	SpikeStartedAt      *time.Time
	NotificationEnabled bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
}
