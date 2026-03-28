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

// TwitterHandler implements the TwitterService Connect RPC handler.
type TwitterHandler struct {
	uc *usecase.TwitterUsecase
}

// NewTwitterHandler creates a new TwitterHandler.
func NewTwitterHandler(uc *usecase.TwitterUsecase) *TwitterHandler {
	return &TwitterHandler{uc: uc}
}

// GetConnectionInfo returns the user's Twitter connection status.
func (h *TwitterHandler) GetConnectionInfo(
	ctx context.Context,
	req *connect.Request[trendbirdv1.GetConnectionInfoRequest],
) (*connect.Response[trendbirdv1.GetConnectionInfoResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	conn, err := h.uc.GetConnectionInfo(ctx, userID)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.GetConnectionInfoResponse{
		Info: converter.TwitterConnectionToProto(conn),
	}), nil
}

// TestConnection tests the Twitter connection and returns the updated status.
func (h *TwitterHandler) TestConnection(
	ctx context.Context,
	req *connect.Request[trendbirdv1.TestConnectionRequest],
) (*connect.Response[trendbirdv1.TestConnectionResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	if err := h.uc.TestConnection(ctx, userID); err != nil {
		return nil, presenter.ToConnectError(err)
	}

	conn, err := h.uc.GetConnectionInfo(ctx, userID)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.TestConnectionResponse{
		Info: converter.TwitterConnectionToProto(conn),
	}), nil
}

// DisconnectTwitter disconnects the user's Twitter account.
func (h *TwitterHandler) DisconnectTwitter(
	ctx context.Context,
	req *connect.Request[trendbirdv1.DisconnectTwitterRequest],
) (*connect.Response[trendbirdv1.DisconnectTwitterResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	if err := h.uc.DisconnectTwitter(ctx, userID); err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.DisconnectTwitterResponse{}), nil
}

var _ trendbirdv1connect.TwitterServiceHandler = (*TwitterHandler)(nil)
