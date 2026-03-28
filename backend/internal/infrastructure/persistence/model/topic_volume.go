package model

import "time"

// TopicVolume is the GORM model for the topic_volumes table.
// This table is immutable (INSERT only) so it has no UpdatedAt field.
type TopicVolume struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TopicID   string    `gorm:"type:uuid;not null;index:idx_topic_volumes_topic_timestamp"`
	Timestamp time.Time `gorm:"not null;index:idx_topic_volumes_topic_timestamp"`
	Value     int32     `gorm:"not null;default:0"`
	CreatedAt time.Time

	Topic Topic `gorm:"foreignKey:TopicID;constraint:OnDelete:CASCADE"`
}

func (TopicVolume) TableName() string {
	return "topic_volumes"
}
