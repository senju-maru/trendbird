package model

import "time"

// AIGenerationLog is the GORM model for the ai_generation_logs table.
// This table is immutable (INSERT only) so it has no UpdatedAt field.
type AIGenerationLog struct {
	ID        string  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string  `gorm:"type:uuid;not null;index:idx_ai_generation_logs_user_created"`
	TopicID   *string `gorm:"type:uuid"`
	Style     int32   `gorm:"not null"`
	Count     int32   `gorm:"not null;default:1"`
	CreatedAt time.Time `gorm:"index:idx_ai_generation_logs_user_created"`

	User  User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Topic *Topic `gorm:"foreignKey:TopicID;constraint:OnDelete:SET NULL"`
}

func (AIGenerationLog) TableName() string {
	return "ai_generation_logs"
}
