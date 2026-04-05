package handler

import (
	"context"

	"connectrpc.com/connect"

	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/gen/trendbird/v1/trendbirdv1connect"
	"github.com/trendbird/backend/internal/adapter/converter"
	"github.com/trendbird/backend/internal/adapter/middleware"
	"github.com/trendbird/backend/internal/adapter/presenter"
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/usecase"
)

// AnalyticsHandler implements the AnalyticsService Connect RPC handler.
type AnalyticsHandler struct {
	uc *usecase.AnalyticsUsecase
}

// NewAnalyticsHandler creates a new AnalyticsHandler.
func NewAnalyticsHandler(uc *usecase.AnalyticsUsecase) *AnalyticsHandler {
	return &AnalyticsHandler{uc: uc}
}

func (h *AnalyticsHandler) ImportDailyAnalytics(
	ctx context.Context,
	req *connect.Request[trendbirdv1.ImportDailyAnalyticsRequest],
) (*connect.Response[trendbirdv1.ImportDailyAnalyticsResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	records := make([]*entity.XAnalyticsDaily, 0, len(req.Msg.GetRecords()))
	for _, r := range req.Msg.GetRecords() {
		e, err := converter.DailyAnalyticsFromProto(r)
		if err != nil {
			return nil, presenter.ToConnectError(err)
		}
		records = append(records, e)
	}

	inserted, updated, err := h.uc.ImportDailyAnalytics(ctx, userID, records)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.ImportDailyAnalyticsResponse{
		ImportedCount: inserted,
		UpdatedCount:  updated,
	}), nil
}

func (h *AnalyticsHandler) ImportPostAnalytics(
	ctx context.Context,
	req *connect.Request[trendbirdv1.ImportPostAnalyticsRequest],
) (*connect.Response[trendbirdv1.ImportPostAnalyticsResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	records := make([]*entity.XAnalyticsPost, 0, len(req.Msg.GetRecords()))
	for _, r := range req.Msg.GetRecords() {
		e, err := converter.PostAnalyticsFromProto(r)
		if err != nil {
			return nil, presenter.ToConnectError(err)
		}
		records = append(records, e)
	}

	inserted, updated, err := h.uc.ImportPostAnalytics(ctx, userID, records)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.ImportPostAnalyticsResponse{
		ImportedCount: inserted,
		UpdatedCount:  updated,
	}), nil
}

func (h *AnalyticsHandler) GetAnalyticsSummary(
	ctx context.Context,
	req *connect.Request[trendbirdv1.GetAnalyticsSummaryRequest],
) (*connect.Response[trendbirdv1.GetAnalyticsSummaryResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	summary, err := h.uc.GetAnalyticsSummary(ctx, userID, req.Msg.StartDate, req.Msg.EndDate)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.GetAnalyticsSummaryResponse{
		Summary: converter.AnalyticsSummaryToProto(summary),
	}), nil
}

func (h *AnalyticsHandler) ListPostAnalytics(
	ctx context.Context,
	req *connect.Request[trendbirdv1.ListPostAnalyticsRequest],
) (*connect.Response[trendbirdv1.ListPostAnalyticsResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	posts, total, err := h.uc.ListPostAnalytics(ctx, userID, req.Msg.SortBy, req.Msg.Limit, req.Msg.StartDate, req.Msg.EndDate)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.ListPostAnalyticsResponse{
		Posts: converter.PostAnalyticsSliceToProto(posts),
		Total: int32(total),
	}), nil
}

func (h *AnalyticsHandler) GetGrowthInsights(
	ctx context.Context,
	req *connect.Request[trendbirdv1.GetGrowthInsightsRequest],
) (*connect.Response[trendbirdv1.GetGrowthInsightsResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	insights, summary, err := h.uc.GetGrowthInsights(ctx, userID, req.Msg.StartDate, req.Msg.EndDate)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.GetGrowthInsightsResponse{
		Insights: converter.GrowthInsightSliceToProto(insights),
		Summary:  converter.AnalyticsSummaryToProto(summary),
	}), nil
}

var _ trendbirdv1connect.AnalyticsServiceHandler = (*AnalyticsHandler)(nil)
