package entity

import "time"

// UserGenre represents a genre selected by a user.
// GenreSlug は genres テーブルとの JOIN で取得され、API レスポンスに使用される。
type UserGenre struct {
	ID        string
	UserID    string
	GenreID   string
	GenreSlug string // genres テーブルから JOIN で取得
	CreatedAt time.Time
}
