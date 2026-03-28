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

// TopicHandler implements the TopicService Connect RPC handler.
type TopicHandler struct {
	uc *usecase.TopicUsecase
}

// NewTopicHandler creates a new TopicHandler.
func NewTopicHandler(uc *usecase.TopicUsecase) *TopicHandler {
	return &TopicHandler{uc: uc}
}

// ListTopics returns all topics for the user.
func (h *TopicHandler) ListTopics(
	ctx context.Context,
	req *connect.Request[trendbirdv1.ListTopicsRequest],
) (*connect.Response[trendbirdv1.ListTopicsResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	topics, err := h.uc.ListTopics(ctx, userID)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.ListTopicsResponse{
		Topics: converter.TopicSliceToProtoSimple(topics),
	}), nil
}

// GetTopic returns a single topic.
func (h *TopicHandler) GetTopic(
	ctx context.Context,
	req *connect.Request[trendbirdv1.GetTopicRequest],
) (*connect.Response[trendbirdv1.GetTopicResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	output, err := h.uc.GetTopic(ctx, userID, req.Msg.GetId())
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.GetTopicResponse{
		Topic: converter.TopicToProto(&converter.TopicToProtoInput{
			Topic:               output.Topic,
			WeeklySparklineData: output.WeeklySparklineData,
			SpikeHistory:        output.SpikeHistory,
		}),
	}), nil
}

// CreateTopic creates a new topic.
func (h *TopicHandler) CreateTopic(
	ctx context.Context,
	req *connect.Request[trendbirdv1.CreateTopicRequest],
) (*connect.Response[trendbirdv1.CreateTopicResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	input := usecase.CreateTopicInput{
		Name:     req.Msg.GetName(),
		Keywords: req.Msg.GetKeywords(),
		Genre:    req.Msg.GetGenre(),
	}

	topic, err := h.uc.CreateTopic(ctx, userID, input)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.CreateTopicResponse{
		Topic: converter.TopicToProtoSimple(topic),
	}), nil
}

// DeleteTopic deletes a topic.
func (h *TopicHandler) DeleteTopic(
	ctx context.Context,
	req *connect.Request[trendbirdv1.DeleteTopicRequest],
) (*connect.Response[trendbirdv1.DeleteTopicResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	if err := h.uc.DeleteTopic(ctx, userID, req.Msg.GetId()); err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.DeleteTopicResponse{}), nil
}

// UpdateTopicNotification updates a topic's notification setting.
func (h *TopicHandler) UpdateTopicNotification(
	ctx context.Context,
	req *connect.Request[trendbirdv1.UpdateTopicNotificationRequest],
) (*connect.Response[trendbirdv1.UpdateTopicNotificationResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	if err := h.uc.UpdateTopicNotification(ctx, userID, req.Msg.GetId(), req.Msg.GetEnabled()); err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.UpdateTopicNotificationResponse{}), nil
}

// AddGenre adds a genre to the user's selection.
func (h *TopicHandler) AddGenre(
	ctx context.Context,
	req *connect.Request[trendbirdv1.AddGenreRequest],
) (*connect.Response[trendbirdv1.AddGenreResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	if err := h.uc.AddGenre(ctx, userID, req.Msg.GetGenre()); err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.AddGenreResponse{}), nil
}

// RemoveGenre removes a genre and its associated topics.
func (h *TopicHandler) RemoveGenre(
	ctx context.Context,
	req *connect.Request[trendbirdv1.RemoveGenreRequest],
) (*connect.Response[trendbirdv1.RemoveGenreResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	if err := h.uc.RemoveGenre(ctx, userID, req.Msg.GetGenre()); err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.RemoveGenreResponse{}), nil
}

// ListUserGenres returns the genres selected by the user.
func (h *TopicHandler) ListUserGenres(
	ctx context.Context,
	req *connect.Request[trendbirdv1.ListUserGenresRequest],
) (*connect.Response[trendbirdv1.ListUserGenresResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	genres, err := h.uc.ListUserGenres(ctx, userID)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.ListUserGenresResponse{
		Genres: genres,
	}), nil
}

// ListGenres returns all genre master data.
func (h *TopicHandler) ListGenres(
	ctx context.Context,
	req *connect.Request[trendbirdv1.ListGenresRequest],
) (*connect.Response[trendbirdv1.ListGenresResponse], error) {
	genres, err := h.uc.ListGenres(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.ListGenresResponse{
		Genres: converter.GenreSliceToProto(genres),
	}), nil
}

// SuggestTopics returns fuzzy-matched topic suggestions.
func (h *TopicHandler) SuggestTopics(
	ctx context.Context,
	req *connect.Request[trendbirdv1.SuggestTopicsRequest],
) (*connect.Response[trendbirdv1.SuggestTopicsResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	var genre string
	if req.Msg.Genre != nil {
		genre = *req.Msg.Genre
	}

	input := usecase.SuggestTopicsInput{
		Query: req.Msg.GetQuery(),
		Genre: genre,
		Limit: int(req.Msg.GetLimit()),
	}

	suggestions, err := h.uc.SuggestTopics(ctx, userID, input)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.SuggestTopicsResponse{
		Suggestions: converter.TopicSuggestionSliceToProto(suggestions),
	}), nil
}

var _ trendbirdv1connect.TopicServiceHandler = (*TopicHandler)(nil)
