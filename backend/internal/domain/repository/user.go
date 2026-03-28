package repository

import (
	"context"

	"github.com/trendbird/backend/internal/domain/entity"
)

// UserRepository defines the persistence operations for User entities.
type UserRepository interface {
	FindByID(ctx context.Context, id string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByTwitterID(ctx context.Context, twitterID string) (*entity.User, error)
	UpsertByTwitterID(ctx context.Context, input entity.UpsertUserInput) (*entity.User, error)
	UpdateEmail(ctx context.Context, id string, email string) error
	CompleteTutorial(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	// ListByIDs returns users matching the given IDs.
	ListByIDs(ctx context.Context, ids []string) ([]*entity.User, error)
}
