package entity

import "time"

// User represents a registered user.
type User struct {
	ID                string
	TwitterID         string
	Name              string
	Email             string
	Image             string
	TwitterHandle     string
	TutorialCompleted bool
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// IsTutorialPending returns true if the user has not yet completed the onboarding tutorial.
func (u *User) IsTutorialPending() bool {
	return !u.TutorialCompleted
}

// UpsertUserInput holds the fields required to create or update a user via X OAuth.
type UpsertUserInput struct {
	TwitterID     string
	Name          string
	Email         string
	Image         string
	TwitterHandle string
}
