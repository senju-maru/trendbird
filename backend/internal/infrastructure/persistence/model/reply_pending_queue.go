package model

import "time"

// ReplyPendingQueue is the GORM model for the reply_pending_queue table.
type ReplyPendingQueue struct {
	ID               string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID           string `gorm:"type:uuid;not null;index:idx_reply_pending_queue_user_id"`
	RuleID           string `gorm:"type:uuid;not null"`
	OriginalTweetID  string `gorm:"type:varchar(64);not null;uniqueIndex:uq_reply_pending,priority:1"`
	OriginalAuthorID string `gorm:"type:varchar(64);not null;uniqueIndex:uq_reply_pending,priority:2"`
	TriggerKeyword   string `gorm:"type:varchar(255);not null"`
	Status           int    `gorm:"not null;default:1"`
	CreatedAt        time.Time

	User          User          `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	AutoReplyRule AutoReplyRule `gorm:"foreignKey:RuleID;constraint:OnDelete:CASCADE"`
}

func (ReplyPendingQueue) TableName() string {
	return "reply_pending_queue"
}
