package model

import "time"

// UserTopic is the GORM model for the user_topics table.
type UserTopic struct {
	ID                  string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID              string `gorm:"type:uuid;not null;uniqueIndex:uq_user_topics_user_topic,priority:1;index:idx_user_topics_user_id"`
	TopicID             string `gorm:"type:uuid;not null;uniqueIndex:uq_user_topics_user_topic,priority:2;index:idx_user_topics_topic_id"`
	NotificationEnabled bool   `gorm:"not null;default:true"`
	IsCreator           bool   `gorm:"not null;default:false"`
	CreatedAt           time.Time

	User  User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Topic Topic `gorm:"foreignKey:TopicID;constraint:OnDelete:CASCADE"`
}

func (UserTopic) TableName() string {
	return "user_topics"
}
