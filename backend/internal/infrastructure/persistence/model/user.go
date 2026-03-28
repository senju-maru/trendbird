package model

import "time"

// User is the GORM model for the users table.
type User struct {
	ID                string  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TwitterID         string  `gorm:"type:varchar(64);uniqueIndex;not null"`
	Name              string  `gorm:"type:varchar(100);not null"`
	Email             string  `gorm:"type:varchar(255);not null;default:''"`
	Image             string  `gorm:"type:text;not null;default:''"`
	TwitterHandle     string `gorm:"type:varchar(15);uniqueIndex;not null"`
	TutorialCompleted bool   `gorm:"not null;default:false"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (User) TableName() string {
	return "users"
}
