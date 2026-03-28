package model

import "time"

// Notification is the GORM model for the notifications table.
// user_id と is_read は user_notifications テーブルに移行済み。
type Notification struct {
	ID          string  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Type        int32   `gorm:"not null"`
	Title       string  `gorm:"type:varchar(200);not null"`
	Message     string  `gorm:"type:varchar(1000);not null"`
	TopicID     *string `gorm:"type:uuid"`
	TopicName   *string `gorm:"type:varchar(100)"`
	TopicStatus *int32  `gorm:""`
	ActionURL   *string `gorm:"type:text"`
	ActionLabel *string `gorm:"type:varchar(100)"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Topic *Topic `gorm:"foreignKey:TopicID;constraint:OnDelete:SET NULL"`
}

func (Notification) TableName() string {
	return "notifications"
}
