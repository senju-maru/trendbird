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

// AutoDMHandler implements the AutoDMService Connect RPC handler.
type AutoDMHandler struct {
	uc *usecase.AutoDMUsecase
}

// NewAutoDMHandler creates a new AutoDMHandler.
func NewAutoDMHandler(uc *usecase.AutoDMUsecase) *AutoDMHandler {
	return &AutoDMHandler{uc: uc}
}

func (h *AutoDMHandler) ListAutoDMRules(
	ctx context.Context,
	req *connect.Request[trendbirdv1.ListAutoDMRulesRequest],
) (*connect.Response[trendbirdv1.ListAutoDMRulesResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	rules, err := h.uc.ListRules(ctx, userID)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.ListAutoDMRulesResponse{
		Rules: converter.AutoDMRuleSliceToProto(rules),
	}), nil
}

func (h *AutoDMHandler) CreateAutoDMRule(
	ctx context.Context,
	req *connect.Request[trendbirdv1.CreateAutoDMRuleRequest],
) (*connect.Response[trendbirdv1.CreateAutoDMRuleResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	rule, err := h.uc.CreateRule(ctx, userID, req.Msg.TriggerKeywords, req.Msg.TemplateMessage)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.CreateAutoDMRuleResponse{
		Rule: converter.AutoDMRuleToProto(rule),
	}), nil
}

func (h *AutoDMHandler) UpdateAutoDMRule(
	ctx context.Context,
	req *connect.Request[trendbirdv1.UpdateAutoDMRuleRequest],
) (*connect.Response[trendbirdv1.UpdateAutoDMRuleResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	rule, err := h.uc.UpdateRule(ctx, userID, req.Msg.Id, req.Msg.Enabled, req.Msg.TriggerKeywords, req.Msg.TemplateMessage)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.UpdateAutoDMRuleResponse{
		Rule: converter.AutoDMRuleToProto(rule),
	}), nil
}

func (h *AutoDMHandler) DeleteAutoDMRule(
	ctx context.Context,
	req *connect.Request[trendbirdv1.DeleteAutoDMRuleRequest],
) (*connect.Response[trendbirdv1.DeleteAutoDMRuleResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	if err := h.uc.DeleteRule(ctx, userID, req.Msg.Id); err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.DeleteAutoDMRuleResponse{}), nil
}

func (h *AutoDMHandler) GetDMSentLogs(
	ctx context.Context,
	req *connect.Request[trendbirdv1.GetDMSentLogsRequest],
) (*connect.Response[trendbirdv1.GetDMSentLogsResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	limit := int(req.Msg.Limit)
	if limit <= 0 {
		limit = 20
	}

	logs, err := h.uc.GetSentLogs(ctx, userID, limit)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.GetDMSentLogsResponse{
		Logs: converter.DMSentLogSliceToProto(logs),
	}), nil
}

var _ trendbirdv1connect.AutoDMServiceHandler = (*AutoDMHandler)(nil)
