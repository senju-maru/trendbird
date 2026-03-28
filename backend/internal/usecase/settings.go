package usecase

import (
	"context"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/domain/repository"
)

// SettingsUsecase handles user profile and notification settings operations.
type SettingsUsecase struct {
	userRepo        repository.UserRepository
	notiSettingRepo repository.NotificationSettingRepository
}

// NewSettingsUsecase creates a new SettingsUsecase.
func NewSettingsUsecase(
	userRepo repository.UserRepository,
	notiSettingRepo repository.NotificationSettingRepository,
) *SettingsUsecase {
	return &SettingsUsecase{
		userRepo:        userRepo,
		notiSettingRepo: notiSettingRepo,
	}
}

// GetProfile returns the user profile.
func (u *SettingsUsecase) GetProfile(ctx context.Context, userID string) (*entity.User, error) {
	return u.userRepo.FindByID(ctx, userID)
}

// UpdateProfile updates the user's email address.
func (u *SettingsUsecase) UpdateProfile(ctx context.Context, userID string, email *string) (*entity.User, error) {
	if email != nil {
		if *email == "" {
			return nil, apperror.InvalidArgument("email must not be empty")
		}
		if err := u.userRepo.UpdateEmail(ctx, userID, *email); err != nil {
			return nil, err
		}
	}
	return u.userRepo.FindByID(ctx, userID)
}

// GetNotificationSettings returns the notification settings for the user.
func (u *SettingsUsecase) GetNotificationSettings(ctx context.Context, userID string) (*entity.NotificationSetting, error) {
	return u.notiSettingRepo.FindByUserID(ctx, userID)
}

// UpdateNotifications updates individual notification setting fields.
// Only non-nil fields are updated.
func (u *SettingsUsecase) UpdateNotifications(ctx context.Context, userID string, spike, rising *bool) error {
	setting, err := u.notiSettingRepo.FindByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if spike != nil {
		setting.SpikeEnabled = *spike
	}
	if rising != nil {
		setting.RisingEnabled = *rising
	}

	return u.notiSettingRepo.Upsert(ctx, setting)
}
