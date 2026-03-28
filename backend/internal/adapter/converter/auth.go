package converter

import (
	"time"

	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/internal/domain/entity"
)

// UserToProto は entity.User を Proto User に変換する。
func UserToProto(e *entity.User) *trendbirdv1.User {
	var email *string
	if e.Email != "" {
		email = &e.Email
	}
	var image *string
	if e.Image != "" {
		image = &e.Image
	}

	return &trendbirdv1.User{
		Id:            e.ID,
		Name:          e.Name,
		Email:         email,
		Image:         image,
		TwitterHandle: e.TwitterHandle,
		CreatedAt:     e.CreatedAt.Format(time.RFC3339),
	}
}
