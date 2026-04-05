package model

import "time"

// XAnalyticsDaily is the GORM model for the x_analytics_daily table.
type XAnalyticsDaily struct {
	ID            string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID        string    `gorm:"type:uuid;not null"`
	Date          time.Time `gorm:"type:date;not null"`
	Impressions   int32     `gorm:"not null;default:0"`
	Likes         int32     `gorm:"not null;default:0"`
	Engagements   int32     `gorm:"not null;default:0"`
	Bookmarks     int32     `gorm:"not null;default:0"`
	Shares        int32     `gorm:"not null;default:0"`
	NewFollows    int32     `gorm:"not null;default:0"`
	Unfollows     int32     `gorm:"not null;default:0"`
	Replies       int32     `gorm:"not null;default:0"`
	Reposts       int32     `gorm:"not null;default:0"`
	ProfileVisits int32     `gorm:"not null;default:0"`
	PostsCreated  int32     `gorm:"not null;default:0"`
	VideoViews    int32     `gorm:"not null;default:0"`
	MediaViews    int32     `gorm:"not null;default:0"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (XAnalyticsDaily) TableName() string {
	return "x_analytics_daily"
}

// XAnalyticsPost is the GORM model for the x_analytics_posts table.
type XAnalyticsPost struct {
	ID              string    `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID          string    `gorm:"type:uuid;not null"`
	PostID          string    `gorm:"type:varchar(64);not null"`
	PostedAt        time.Time `gorm:"not null"`
	PostText        string    `gorm:"type:text;not null;default:''"`
	PostURL         string    `gorm:"type:text;not null;default:''"`
	Impressions     int32     `gorm:"not null;default:0"`
	Likes           int32     `gorm:"not null;default:0"`
	Engagements     int32     `gorm:"not null;default:0"`
	Bookmarks       int32     `gorm:"not null;default:0"`
	Shares          int32     `gorm:"not null;default:0"`
	NewFollows      int32     `gorm:"not null;default:0"`
	Replies         int32     `gorm:"not null;default:0"`
	Reposts         int32     `gorm:"not null;default:0"`
	ProfileVisits   int32     `gorm:"not null;default:0"`
	DetailClicks    int32     `gorm:"not null;default:0"`
	URLClicks       int32     `gorm:"not null;default:0"`
	HashtagClicks   int32     `gorm:"not null;default:0"`
	PermalinkClicks int32     `gorm:"not null;default:0"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (XAnalyticsPost) TableName() string {
	return "x_analytics_posts"
}
