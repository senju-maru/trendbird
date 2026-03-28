package handler

import (
	"context"

	"connectrpc.com/connect"

	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/gen/trendbird/v1/trendbirdv1connect"
	"github.com/trendbird/backend/internal/adapter/converter"
	"github.com/trendbird/backend/internal/adapter/middleware"
	"github.com/trendbird/backend/internal/adapter/presenter"
	"github.com/trendbird/backend/internal/usecase"
)

// SettingsHandler implements the SettingsService Connect RPC handler.
type SettingsHandler struct {
	uc *usecase.SettingsUsecase
}

// NewSettingsHandler creates a new SettingsHandler.
func NewSettingsHandler(uc *usecase.SettingsUsecase) *SettingsHandler {
	return &SettingsHandler{uc: uc}
}

// GetProfile returns the authenticated user's profile.
func (h *SettingsHandler) GetProfile(
	ctx context.Context,
	req *connect.Request[trendbirdv1.GetProfileRequest],
) (*connect.Response[trendbirdv1.GetProfileResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	user, err := h.uc.GetProfile(ctx, userID)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.GetProfileResponse{
		User: converter.UserToProto(user),
	}), nil
}

// UpdateProfile updates the authenticated user's profile.
func (h *SettingsHandler) UpdateProfile(
	ctx context.Context,
	req *connect.Request[trendbirdv1.UpdateProfileRequest],
) (*connect.Response[trendbirdv1.UpdateProfileResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	user, err := h.uc.UpdateProfile(ctx, userID, req.Msg.Email)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.UpdateProfileResponse{
		User: converter.UserToProto(user),
	}), nil
}

// GetNotificationSettings returns the user's notification settings.
func (h *SettingsHandler) GetNotificationSettings(
	ctx context.Context,
	req *connect.Request[trendbirdv1.GetNotificationSettingsRequest],
) (*connect.Response[trendbirdv1.GetNotificationSettingsResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	settings, err := h.uc.GetNotificationSettings(ctx, userID)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.GetNotificationSettingsResponse{
		Settings: converter.NotificationSettingToProto(settings),
	}), nil
}

// UpdateNotifications updates the user's notification settings.
func (h *SettingsHandler) UpdateNotifications(
	ctx context.Context,
	req *connect.Request[trendbirdv1.UpdateNotificationsRequest],
) (*connect.Response[trendbirdv1.UpdateNotificationsResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	if err := h.uc.UpdateNotifications(ctx, userID, req.Msg.SpikeEnabled, req.Msg.RisingEnabled); err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.UpdateNotificationsResponse{
		Updated: true,
	}), nil
}

var _ trendbirdv1connect.SettingsServiceHandler = (*SettingsHandler)(nil)
