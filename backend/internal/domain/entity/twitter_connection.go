package entity

import "time"

// TwitterConnectionStatus represents the state of a user's X connection.
type TwitterConnectionStatus int

const (
	TwitterDisconnected TwitterConnectionStatus = 1
	TwitterConnecting   TwitterConnectionStatus = 2
	TwitterConnected    TwitterConnectionStatus = 3
	TwitterError        TwitterConnectionStatus = 4
)

// TwitterConnection holds X OAuth tokens and connection state for a user.
type TwitterConnection struct {
	ID             string
	UserID         string
	AccessToken    string
	RefreshToken   string
	TokenExpiresAt time.Time
	Status         TwitterConnectionStatus
	ConnectedAt    *time.Time
	LastTestedAt   *time.Time
	ErrorMessage   *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
