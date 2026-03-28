package repository

import (
	"context"

	"github.com/trendbird/backend/internal/domain/entity"
)

// NotificationSettingRepository defines the persistence operations for NotificationSetting entities.
type NotificationSettingRepository interface {
	FindByUserID(ctx context.Context, userID string) (*entity.NotificationSetting, error)
	Upsert(ctx context.Context, setting *entity.NotificationSetting) error
}
