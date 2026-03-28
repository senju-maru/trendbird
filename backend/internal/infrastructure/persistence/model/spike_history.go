package model

import "time"

// SpikeHistory is the GORM model for the spike_histories table.
// This table is immutable (INSERT only) so it has no UpdatedAt field,
// except for notified_at which is updated by batch jobs.
type SpikeHistory struct {
	ID              string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TopicID         string     `gorm:"type:uuid;not null;index:idx_spike_histories_topic_id"`
	Timestamp       time.Time  `gorm:"not null"`
	PeakZScore      float64    `gorm:"not null"`
	Status          int32      `gorm:"not null"`
	Summary         string     `gorm:"type:varchar(500);not null"`
	DurationMinutes int32      `gorm:"not null"`
	NotifiedAt *time.Time `gorm:"type:timestamptz"`
	CreatedAt       time.Time

	Topic Topic `gorm:"foreignKey:TopicID;constraint:OnDelete:CASCADE"`
}

func (SpikeHistory) TableName() string {
	return "spike_histories"
}
