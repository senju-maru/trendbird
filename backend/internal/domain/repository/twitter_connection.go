package repository

import (
	"context"

	"github.com/trendbird/backend/internal/domain/entity"
)

// TwitterConnectionRepository defines the persistence operations for TwitterConnection entities.
type TwitterConnectionRepository interface {
	FindByUserID(ctx context.Context, userID string) (*entity.TwitterConnection, error)
	Upsert(ctx context.Context, conn *entity.TwitterConnection) error
	UpdateStatus(ctx context.Context, userID string, status entity.TwitterConnectionStatus, errorMessage *string) error
	UpdateLastTestedAt(ctx context.Context, userID string) error
	DeleteByUserID(ctx context.Context, userID string) error
}
