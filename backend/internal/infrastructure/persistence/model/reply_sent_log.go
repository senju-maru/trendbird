package model

import "time"

// ReplySentLog is the GORM model for the reply_sent_logs table.
type ReplySentLog struct {
	ID               string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID           string `gorm:"type:uuid;not null;index:idx_reply_sent_logs_user_id"`
	RuleID           string `gorm:"type:uuid;not null"`
	OriginalTweetID  string `gorm:"type:varchar(64);not null;uniqueIndex:uq_reply_sent_logs,priority:1"`
	OriginalAuthorID string `gorm:"type:varchar(64);not null;uniqueIndex:uq_reply_sent_logs,priority:2"`
	ReplyTweetID     string `gorm:"type:varchar(64);not null"`
	TriggerKeyword   string `gorm:"type:varchar(255);not null"`
	ReplyText        string `gorm:"type:text;not null"`
	SentAt           time.Time

	User          User          `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	AutoReplyRule AutoReplyRule `gorm:"foreignKey:RuleID;constraint:OnDelete:CASCADE"`
}

func (ReplySentLog) TableName() string {
	return "reply_sent_logs"
}
