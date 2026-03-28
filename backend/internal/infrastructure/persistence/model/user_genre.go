package model

import "time"

// UserGenre is the GORM model for the user_genres table.
type UserGenre struct {
	ID        string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    string    `gorm:"type:uuid;not null;index:idx_user_genres_user_id"`
	GenreID   string    `gorm:"type:uuid;not null"`
	CreatedAt time.Time `gorm:"not null;default:now()"`

	User  User  `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Genre Genre `gorm:"foreignKey:GenreID;constraint:OnDelete:RESTRICT"`
}

func (UserGenre) TableName() string {
	return "user_genres"
}
