package repository

import (
	"context"

	"github.com/trendbird/backend/internal/domain/entity"
)

// AIGenerationLogRepository defines the persistence operations for AIGenerationLog entities.
type AIGenerationLogRepository interface {
	Create(ctx context.Context, log *entity.AIGenerationLog) error
	CountByUserIDCurrentMonth(ctx context.Context, userID string) (int32, error)
}
