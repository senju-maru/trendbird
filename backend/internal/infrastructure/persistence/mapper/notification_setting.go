package mapper

import (
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// NotificationSettingToEntity converts a GORM NotificationSetting model to a domain entity.
func NotificationSettingToEntity(m *model.NotificationSetting) *entity.NotificationSetting {
	return &entity.NotificationSetting{
		ID:            m.ID,
		UserID:        m.UserID,
		SpikeEnabled:  m.SpikeEnabled,
		RisingEnabled: m.RisingEnabled,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

// NotificationSettingToModel converts a domain NotificationSetting entity to a GORM model.
func NotificationSettingToModel(e *entity.NotificationSetting) *model.NotificationSetting {
	return &model.NotificationSetting{
		ID:            e.ID,
		UserID:        e.UserID,
		SpikeEnabled:  e.SpikeEnabled,
		RisingEnabled: e.RisingEnabled,
		CreatedAt:     e.CreatedAt,
		UpdatedAt:     e.UpdatedAt,
	}
}
