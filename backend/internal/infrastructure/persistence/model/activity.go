package model

import "time"

// Activity is the GORM model for the activities table.
// This table is immutable (INSERT only) so it has no UpdatedAt field.
type Activity struct {
	ID          string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      string    `gorm:"type:uuid;not null;index:idx_activities_user_timestamp"`
	Type        int32     `gorm:"not null"`
	TopicName   string    `gorm:"type:varchar(100);not null;default:''"`
	Description string    `gorm:"type:varchar(500);not null"`
	Timestamp   time.Time `gorm:"not null;index:idx_activities_user_timestamp,sort:desc"`
	CreatedAt   time.Time

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (Activity) TableName() string {
	return "activities"
}
