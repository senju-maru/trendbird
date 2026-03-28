package model

import "time"

// AutoReplyRule is the GORM model for the auto_reply_rules table.
type AutoReplyRule struct {
	ID                 string      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID             string      `gorm:"type:uuid;not null;index:idx_auto_reply_rules_user_id"`
	Enabled            bool        `gorm:"not null;default:false"`
	TargetTweetID      string      `gorm:"type:varchar(64);not null;uniqueIndex:uq_auto_reply_rules_user_tweet,priority:2"`
	TargetTweetText    string      `gorm:"type:text;not null;default:''"`
	TriggerKeywords    StringArray `gorm:"type:text[];not null;default:'{}'"`
	ReplyTemplate      string      `gorm:"type:text;not null;default:''"`
	LastCheckedReplyID *string     `gorm:"type:varchar(64)"`
	CreatedAt          time.Time
	UpdatedAt          time.Time

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (AutoReplyRule) TableName() string {
	return "auto_reply_rules"
}
