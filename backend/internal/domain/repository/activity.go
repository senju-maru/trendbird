package repository

import (
	"context"

	"github.com/trendbird/backend/internal/domain/entity"
)

// ActivityRepository defines the persistence operations for Activity entities.
type ActivityRepository interface {
	Create(ctx context.Context, activity *entity.Activity) error
	ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*entity.Activity, error)
}
