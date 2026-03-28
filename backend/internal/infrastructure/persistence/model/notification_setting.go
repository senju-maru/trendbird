package model

import "time"

// NotificationSetting is the GORM model for the notification_settings table.
type NotificationSetting struct {
	ID            string `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID        string `gorm:"type:uuid;uniqueIndex;not null"`
	SpikeEnabled  bool   `gorm:"not null;default:true"`
	RisingEnabled bool   `gorm:"not null;default:true"`
	CreatedAt     time.Time
	UpdatedAt     time.Time

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (NotificationSetting) TableName() string {
	return "notification_settings"
}
