package model

import "time"

// UserNotification is the GORM model for the user_notifications table.
type UserNotification struct {
	ID             string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID         string    `gorm:"type:uuid;not null"`
	NotificationID string    `gorm:"type:uuid;not null"`
	IsRead         bool      `gorm:"not null;default:false"`
	CreatedAt      time.Time `gorm:"not null;default:now()"`

	User         User         `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Notification Notification `gorm:"foreignKey:NotificationID;constraint:OnDelete:CASCADE"`
}

func (UserNotification) TableName() string {
	return "user_notifications"
}
