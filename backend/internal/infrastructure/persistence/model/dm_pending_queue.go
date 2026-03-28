package model

import "time"

// DMPendingQueue is the GORM model for the dm_pending_queue table.
type DMPendingQueue struct {
	ID                 string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID             string `gorm:"type:uuid;not null;index:idx_dm_pending_queue_user_id"`
	RuleID             string `gorm:"type:uuid;not null"`
	RecipientTwitterID string `gorm:"type:varchar(64);not null"`
	ReplyTweetID       string `gorm:"type:varchar(64);not null;uniqueIndex:uq_dm_pending_reply,priority:1"`
	TriggerKeyword     string `gorm:"type:varchar(255);not null"`
	Status             int    `gorm:"not null;default:1"`
	CreatedAt          time.Time

	User       User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	AutoDMRule AutoDMRule `gorm:"foreignKey:RuleID;constraint:OnDelete:CASCADE"`
}

func (DMPendingQueue) TableName() string {
	return "dm_pending_queue"
}
