package model

import "time"

// DMSentLog is the GORM model for the dm_sent_logs table.
type DMSentLog struct {
	ID                 string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID             string `gorm:"type:uuid;not null;index:idx_dm_sent_logs_user_id"`
	RuleID             string `gorm:"type:uuid;not null"`
	RecipientTwitterID string `gorm:"type:varchar(64);not null"`
	ReplyTweetID       string `gorm:"type:varchar(64);not null;uniqueIndex:uq_dm_sent_logs_reply,priority:1"`
	TriggerKeyword     string `gorm:"type:varchar(255);not null"`
	DMText             string `gorm:"type:text;not null"`
	SentAt             time.Time

	User        User       `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	AutoDMRule  AutoDMRule `gorm:"foreignKey:RuleID;constraint:OnDelete:CASCADE"`
}

func (DMSentLog) TableName() string {
	return "dm_sent_logs"
}
