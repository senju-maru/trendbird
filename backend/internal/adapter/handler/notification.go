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

// NotificationHandler implements the NotificationService Connect RPC handler.
type NotificationHandler struct {
	uc *usecase.NotificationUsecase
}

// NewNotificationHandler creates a new NotificationHandler.
func NewNotificationHandler(uc *usecase.NotificationUsecase) *NotificationHandler {
	return &NotificationHandler{uc: uc}
}

// ListNotifications returns the user's notifications.
func (h *NotificationHandler) ListNotifications(
	ctx context.Context,
	req *connect.Request[trendbirdv1.ListNotificationsRequest],
) (*connect.Response[trendbirdv1.ListNotificationsResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	notifications, _, err := h.uc.ListNotifications(ctx, userID, 50, 0)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.ListNotificationsResponse{
		Notifications: converter.NotificationSliceToProto(notifications),
	}), nil
}

// MarkAsRead marks a single notification as read.
func (h *NotificationHandler) MarkAsRead(
	ctx context.Context,
	req *connect.Request[trendbirdv1.MarkAsReadRequest],
) (*connect.Response[trendbirdv1.MarkAsReadResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	if err := h.uc.MarkAsRead(ctx, userID, req.Msg.GetId()); err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.MarkAsReadResponse{}), nil
}

// MarkAllAsRead marks all notifications as read.
func (h *NotificationHandler) MarkAllAsRead(
	ctx context.Context,
	req *connect.Request[trendbirdv1.MarkAllAsReadRequest],
) (*connect.Response[trendbirdv1.MarkAllAsReadResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	if err := h.uc.MarkAllAsRead(ctx, userID); err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.MarkAllAsReadResponse{}), nil
}

var _ trendbirdv1connect.NotificationServiceHandler = (*NotificationHandler)(nil)
