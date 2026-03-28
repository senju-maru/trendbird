package model

import "time"

// Post is the GORM model for the posts table.
type Post struct {
	ID           string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID       string     `gorm:"type:uuid;not null;index:idx_posts_user_id_status"`
	Content      string     `gorm:"type:text;not null"`
	TopicID      *string    `gorm:"type:uuid"`
	TopicName    *string    `gorm:"type:varchar(100)"`
	Status       int32      `gorm:"not null;default:1;index:idx_posts_user_id_status"`
	ScheduledAt  *time.Time `gorm:""`
	PublishedAt  *time.Time `gorm:""`
	FailedAt     *time.Time `gorm:""`
	ErrorMessage *string    `gorm:"type:varchar(1000)"`
	TweetURL     *string    `gorm:"type:text"`
	Likes        int32      `gorm:"not null;default:0"`
	Retweets     int32      `gorm:"not null;default:0"`
	Replies      int32      `gorm:"not null;default:0"`
	Views        int32      `gorm:"not null;default:0"`
	CreatedAt    time.Time
	UpdatedAt    time.Time

	User  User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Topic *Topic `gorm:"foreignKey:TopicID;constraint:OnDelete:SET NULL"`
}

func (Post) TableName() string {
	return "posts"
}
