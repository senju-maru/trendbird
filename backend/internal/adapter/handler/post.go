package handler

import (
	"context"
	"time"

	"connectrpc.com/connect"

	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/gen/trendbird/v1/trendbirdv1connect"
	"github.com/trendbird/backend/internal/adapter/converter"
	"github.com/trendbird/backend/internal/adapter/middleware"
	"github.com/trendbird/backend/internal/adapter/presenter"
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/usecase"
)

// PostHandler implements the PostService Connect RPC handler.
type PostHandler struct {
	uc *usecase.PostUsecase
}

// NewPostHandler creates a new PostHandler.
func NewPostHandler(uc *usecase.PostUsecase) *PostHandler {
	return &PostHandler{uc: uc}
}

// GeneratePosts generates AI post content for a topic.
func (h *PostHandler) GeneratePosts(
	ctx context.Context,
	req *connect.Request[trendbirdv1.GeneratePostsRequest],
) (*connect.Response[trendbirdv1.GeneratePostsResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	var style *entity.PostStyle
	if req.Msg.Style != nil {
		s := entity.PostStyle(*req.Msg.Style)
		style = &s
	}

	posts, err := h.uc.GeneratePosts(ctx, userID, req.Msg.GetTopicId(), style)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.GeneratePostsResponse{
		Posts: converter.GeneratedPostSliceToProto(posts),
	}), nil
}

// ListDrafts returns the user's draft posts with statistics.
func (h *PostHandler) ListDrafts(
	ctx context.Context,
	req *connect.Request[trendbirdv1.ListDraftsRequest],
) (*connect.Response[trendbirdv1.ListDraftsResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	drafts, stats, _, err := h.uc.ListDrafts(ctx, userID, 50, 0)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.ListDraftsResponse{
		Drafts: converter.PostSliceToScheduledPostProto(drafts),
		Stats:  postStatsToProto(stats),
	}), nil
}

// CreateDraft creates a new draft post.
func (h *PostHandler) CreateDraft(
	ctx context.Context,
	req *connect.Request[trendbirdv1.CreateDraftRequest],
) (*connect.Response[trendbirdv1.CreateDraftResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	draft, err := h.uc.CreateDraft(ctx, userID, req.Msg.GetContent(), req.Msg.TopicId)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.CreateDraftResponse{
		Draft: converter.PostToScheduledPostProto(draft),
	}), nil
}

// UpdateDraft updates an existing draft post.
func (h *PostHandler) UpdateDraft(
	ctx context.Context,
	req *connect.Request[trendbirdv1.UpdateDraftRequest],
) (*connect.Response[trendbirdv1.UpdateDraftResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	draft, err := h.uc.UpdateDraft(ctx, userID, req.Msg.GetId(), req.Msg.GetContent())
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.UpdateDraftResponse{
		Draft: converter.PostToScheduledPostProto(draft),
	}), nil
}

// DeleteDraft deletes a draft post.
func (h *PostHandler) DeleteDraft(
	ctx context.Context,
	req *connect.Request[trendbirdv1.DeleteDraftRequest],
) (*connect.Response[trendbirdv1.DeleteDraftResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	if err := h.uc.DeleteDraft(ctx, userID, req.Msg.GetId()); err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.DeleteDraftResponse{}), nil
}

// SchedulePost schedules a draft for future publishing.
func (h *PostHandler) SchedulePost(
	ctx context.Context,
	req *connect.Request[trendbirdv1.SchedulePostRequest],
) (*connect.Response[trendbirdv1.SchedulePostResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.Msg.GetScheduledAt())
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	post, err := h.uc.SchedulePost(ctx, userID, req.Msg.GetId(), scheduledAt)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.SchedulePostResponse{
		Draft: converter.PostToScheduledPostProto(post),
	}), nil
}

// PublishPost publishes a post to X.
func (h *PostHandler) PublishPost(
	ctx context.Context,
	req *connect.Request[trendbirdv1.PublishPostRequest],
) (*connect.Response[trendbirdv1.PublishPostResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	post, err := h.uc.PublishPost(ctx, userID, req.Msg.GetId())
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.PublishPostResponse{
		Post: converter.PostToPostHistoryProto(post),
	}), nil
}

// ListPostHistory returns the user's published posts.
func (h *PostHandler) ListPostHistory(
	ctx context.Context,
	req *connect.Request[trendbirdv1.ListPostHistoryRequest],
) (*connect.Response[trendbirdv1.ListPostHistoryResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	posts, _, err := h.uc.ListPostHistory(ctx, userID, 50, 0)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.ListPostHistoryResponse{
		Posts: converter.PostSliceToPostHistoryProto(posts),
	}), nil
}

// GetPostStats returns the user's post statistics.
func (h *PostHandler) GetPostStats(
	ctx context.Context,
	req *connect.Request[trendbirdv1.GetPostStatsRequest],
) (*connect.Response[trendbirdv1.GetPostStatsResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	stats, err := h.uc.GetPostStats(ctx, userID)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.GetPostStatsResponse{
		Stats: postStatsToProto(stats),
	}), nil
}

// postStatsToProto converts PostStatsResult to proto PostStats.
func postStatsToProto(stats *usecase.PostStatsResult) *trendbirdv1.PostStats {
	if stats == nil {
		return nil
	}
	return &trendbirdv1.PostStats{
		TotalPublished:     stats.TotalPublished,
		TotalScheduled:     stats.TotalScheduled,
		TotalDrafts:        stats.TotalDrafts,
		TotalFailed:        stats.TotalFailed,
		ThisMonthPublished: stats.ThisMonthPublished,
	}
}

var _ trendbirdv1connect.PostServiceHandler = (*PostHandler)(nil)
