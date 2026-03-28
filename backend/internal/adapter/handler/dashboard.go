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

// DashboardHandler implements the DashboardService Connect RPC handler.
type DashboardHandler struct {
	uc *usecase.DashboardUsecase
}

// NewDashboardHandler creates a new DashboardHandler.
func NewDashboardHandler(uc *usecase.DashboardUsecase) *DashboardHandler {
	return &DashboardHandler{uc: uc}
}

// GetActivities returns the user's recent activities.
func (h *DashboardHandler) GetActivities(
	ctx context.Context,
	req *connect.Request[trendbirdv1.GetActivitiesRequest],
) (*connect.Response[trendbirdv1.GetActivitiesResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	activities, err := h.uc.GetActivities(ctx, userID, 20, 0)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.GetActivitiesResponse{
		Activities: converter.ActivitySliceToProto(activities),
	}), nil
}

// GetStats returns the user's dashboard statistics.
func (h *DashboardHandler) GetStats(
	ctx context.Context,
	req *connect.Request[trendbirdv1.GetStatsRequest],
) (*connect.Response[trendbirdv1.GetStatsResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	stats, err := h.uc.GetStats(ctx, userID)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.GetStatsResponse{
		Stats: converter.DashboardStatsToProto(stats.Detections, stats.Generations, stats.LastCheckedAt),
	}), nil
}

var _ trendbirdv1connect.DashboardServiceHandler = (*DashboardHandler)(nil)
