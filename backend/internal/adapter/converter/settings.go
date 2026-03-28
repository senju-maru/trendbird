package converter

import (
	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/internal/domain/entity"
)

// NotificationSettingToProto は entity.NotificationSetting を Proto NotificationSettings に変換する。
func NotificationSettingToProto(e *entity.NotificationSetting) *trendbirdv1.NotificationSettings {
	return &trendbirdv1.NotificationSettings{
		SpikeEnabled:  e.SpikeEnabled,
		RisingEnabled: e.RisingEnabled,
	}
}
