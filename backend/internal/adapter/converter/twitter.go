package converter

import (
	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/internal/domain/entity"
)

// TwitterConnectionToProto は entity.TwitterConnection を Proto TwitterConnectionInfo に変換する。
// AccessToken, RefreshToken 等の機密フィールドは Proto に含めない。
func TwitterConnectionToProto(e *entity.TwitterConnection) *trendbirdv1.TwitterConnectionInfo {
	return &trendbirdv1.TwitterConnectionInfo{
		Status:       trendbirdv1.TwitterConnectionStatus(e.Status),
		ConnectedAt:  timeToOptionalString(e.ConnectedAt),
		LastTestedAt: timeToOptionalString(e.LastTestedAt),
		ErrorMessage: e.ErrorMessage,
	}
}
