package model

import "time"

// TwitterConnection is the GORM model for the twitter_connections table.
type TwitterConnection struct {
	ID             string     `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID         string     `gorm:"type:uuid;uniqueIndex;not null"`
	AccessToken    string     `gorm:"type:text;not null"`
	RefreshToken   string     `gorm:"type:text;not null"`
	TokenExpiresAt time.Time  `gorm:"not null"`
	Status         int32      `gorm:"not null;default:3"`
	ConnectedAt    *time.Time `gorm:""`
	LastTestedAt   *time.Time `gorm:""`
	ErrorMessage   *string    `gorm:"type:varchar(1000)"`
	CreatedAt      time.Time
	UpdatedAt      time.Time

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (TwitterConnection) TableName() string {
	return "twitter_connections"
}
