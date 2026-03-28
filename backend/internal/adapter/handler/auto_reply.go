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

var _ trendbirdv1connect.AutoReplyServiceHandler = (*AutoReplyHandler)(nil)

// AutoReplyHandler implements the AutoReplyService Connect RPC handler.
type AutoReplyHandler struct {
	uc *usecase.AutoReplyUsecase
}

// NewAutoReplyHandler creates a new AutoReplyHandler.
func NewAutoReplyHandler(uc *usecase.AutoReplyUsecase) *AutoReplyHandler {
	return &AutoReplyHandler{uc: uc}
}

func (h *AutoReplyHandler) ListAutoReplyRules(
	ctx context.Context,
	req *connect.Request[trendbirdv1.ListAutoReplyRulesRequest],
) (*connect.Response[trendbirdv1.ListAutoReplyRulesResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	rules, err := h.uc.ListRules(ctx, userID)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.ListAutoReplyRulesResponse{
		Rules: converter.AutoReplyRuleSliceToProto(rules),
	}), nil
}

func (h *AutoReplyHandler) CreateAutoReplyRule(
	ctx context.Context,
	req *connect.Request[trendbirdv1.CreateAutoReplyRuleRequest],
) (*connect.Response[trendbirdv1.CreateAutoReplyRuleResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	rule, err := h.uc.CreateRule(ctx, userID, req.Msg.TargetTweetId, req.Msg.TargetTweetText, req.Msg.TriggerKeywords, req.Msg.ReplyTemplate)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.CreateAutoReplyRuleResponse{
		Rule: converter.AutoReplyRuleToProto(rule),
	}), nil
}

func (h *AutoReplyHandler) UpdateAutoReplyRule(
	ctx context.Context,
	req *connect.Request[trendbirdv1.UpdateAutoReplyRuleRequest],
) (*connect.Response[trendbirdv1.UpdateAutoReplyRuleResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	rule, err := h.uc.UpdateRule(ctx, userID, req.Msg.Id, req.Msg.Enabled, req.Msg.TriggerKeywords, req.Msg.ReplyTemplate)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.UpdateAutoReplyRuleResponse{
		Rule: converter.AutoReplyRuleToProto(rule),
	}), nil
}

func (h *AutoReplyHandler) DeleteAutoReplyRule(
	ctx context.Context,
	req *connect.Request[trendbirdv1.DeleteAutoReplyRuleRequest],
) (*connect.Response[trendbirdv1.DeleteAutoReplyRuleResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	if err := h.uc.DeleteRule(ctx, userID, req.Msg.Id); err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.DeleteAutoReplyRuleResponse{}), nil
}

func (h *AutoReplyHandler) GetReplySentLogs(
	ctx context.Context,
	req *connect.Request[trendbirdv1.GetReplySentLogsRequest],
) (*connect.Response[trendbirdv1.GetReplySentLogsResponse], error) {
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

	return connect.NewResponse(&trendbirdv1.GetReplySentLogsResponse{
		Logs: converter.ReplySentLogSliceToProto(logs),
	}), nil
}
