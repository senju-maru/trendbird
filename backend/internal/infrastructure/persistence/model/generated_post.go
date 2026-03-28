package model

import "time"

// GeneratedPost is the GORM model for the generated_posts table.
// This table is immutable (INSERT only) so it has no UpdatedAt field.
type GeneratedPost struct {
	ID              string  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID          string  `gorm:"type:uuid;not null;index:idx_generated_posts_user_id"`
	TopicID         *string `gorm:"type:uuid"`
	GenerationLogID *string `gorm:"type:uuid;index:idx_generated_posts_generation_log_id"`
	Style           int32   `gorm:"not null"`
	Content         string  `gorm:"type:text;not null"`
	CreatedAt       time.Time

	User          User           `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Topic         *Topic         `gorm:"foreignKey:TopicID;constraint:OnDelete:SET NULL"`
	GenerationLog *AIGenerationLog `gorm:"foreignKey:GenerationLogID;constraint:OnDelete:SET NULL"`
}

func (GeneratedPost) TableName() string {
	return "generated_posts"
}
