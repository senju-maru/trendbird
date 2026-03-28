package usecase

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/domain/gateway"
	"github.com/trendbird/backend/internal/domain/repository"
)

// CreateTopicInput holds the parameters for creating a new topic.
type CreateTopicInput struct {
	Name     string
	Keywords []string
	Genre    string // genre slug
}

// GetTopicOutput holds the result of GetTopic with detail data.
type GetTopicOutput struct {
	Topic               *entity.Topic
	WeeklySparklineData []*entity.TopicVolume
	SpikeHistory        []*entity.SpikeHistory
}

// TopicUsecase handles topic management operations.
type TopicUsecase struct {
	topicRepo       repository.TopicRepository
	userTopicRepo   repository.UserTopicRepository
	userGenreRepo   repository.UserGenreRepository
	genreRepo       repository.GenreRepository
	activityRepo    repository.ActivityRepository
	topicVolumeRepo repository.TopicVolumeRepository
	spikeHistRepo   repository.SpikeHistoryRepository
	aiGateway       gateway.AIGateway
	twitterGW       gateway.TwitterGateway
	bearerToken     string
	txManager       repository.TransactionManager
}

// NewTopicUsecase creates a new TopicUsecase.
func NewTopicUsecase(
	topicRepo repository.TopicRepository,
	userTopicRepo repository.UserTopicRepository,
	userGenreRepo repository.UserGenreRepository,
	genreRepo repository.GenreRepository,
	activityRepo repository.ActivityRepository,
	topicVolumeRepo repository.TopicVolumeRepository,
	spikeHistRepo repository.SpikeHistoryRepository,
	aiGateway gateway.AIGateway,
	twitterGW gateway.TwitterGateway,
	bearerToken string,
	txManager repository.TransactionManager,
) *TopicUsecase {
	return &TopicUsecase{
		topicRepo:       topicRepo,
		userTopicRepo:   userTopicRepo,
		userGenreRepo:   userGenreRepo,
		genreRepo:       genreRepo,
		activityRepo:    activityRepo,
		topicVolumeRepo: topicVolumeRepo,
		spikeHistRepo:   spikeHistRepo,
		aiGateway:       aiGateway,
		twitterGW:       twitterGW,
		bearerToken:     bearerToken,
		txManager:       txManager,
	}
}

// ListTopics returns all topics for the user.
func (u *TopicUsecase) ListTopics(ctx context.Context, userID string) ([]*entity.Topic, error) {
	return u.topicRepo.ListByUserID(ctx, userID)
}

// GetTopic returns a single topic with detail data after verifying user-topic link.
func (u *TopicUsecase) GetTopic(ctx context.Context, userID string, topicID string) (*GetTopicOutput, error) {
	topic, err := u.topicRepo.FindByIDForUser(ctx, topicID, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	weeklyVolumes, err := u.topicVolumeRepo.ListByTopicIDAndRange(ctx, topicID, now.AddDate(0, 0, -7), now)
	if err != nil {
		return nil, err
	}

	spikeHistory, err := u.spikeHistRepo.ListByTopicID(ctx, topicID)
	if err != nil {
		return nil, err
	}

	return &GetTopicOutput{
		Topic:               topic,
		WeeklySparklineData: weeklyVolumes,
		SpikeHistory:        spikeHistory,
	}, nil
}

// CreateTopic creates a new topic.
// If the genre is not yet registered in user_genres, it is automatically added.
// If a topic with the same name+genre already exists, it reuses the existing topic.
// Genre registration, topic creation, and user-topic linking run inside a single transaction.
func (u *TopicUsecase) CreateTopic(ctx context.Context, userID string, input CreateTopicInput) (*entity.Topic, error) {
	// Resolve genre slug → genre ID (read-only, outside transaction)
	genre, err := u.genreRepo.FindBySlug(ctx, input.Genre)
	if err != nil {
		return nil, err
	}

	// Wrap all writes in a transaction to prevent partial failures and TOCTOU races.
	var topic *entity.Topic
	var isNewTopic bool
	if err := u.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		// Auto-register genre if not yet registered
		exists, err := u.userGenreRepo.ExistsByUserIDAndGenre(txCtx, userID, genre.ID)
		if err != nil {
			return err
		}
		if !exists {
			if err := u.userGenreRepo.Create(txCtx, &entity.UserGenre{
				UserID:  userID,
				GenreID: genre.ID,
			}); err != nil {
				return err
			}
		}

		// Find or create shared topic.
		// topicRepo.Create uses ON CONFLICT DO NOTHING so concurrent inserts are safe.
		existing, err := u.topicRepo.FindByNameAndGenre(txCtx, input.Name, genre.ID)
		if err != nil {
			return err
		}

		isNewTopic = existing == nil

		if isNewTopic {
			newTopic := &entity.Topic{
				Name:      input.Name,
				Keywords:  input.Keywords,
				GenreID:   genre.ID,
				GenreSlug: genre.Slug,
				Status:    entity.TopicStable,
			}
			if err := u.topicRepo.Create(txCtx, newTopic); err != nil {
				return err
			}
			// ON CONFLICT DO NOTHING: if ID is empty, another user created it concurrently
			if newTopic.ID == "" {
				refetched, err := u.topicRepo.FindByNameAndGenre(txCtx, input.Name, genre.ID)
				if err != nil {
					return err
				}
				topic = refetched
				isNewTopic = false
			} else {
				topic = newTopic
			}
		} else {
			topic = existing
		}

		// Create user-topic link (idempotent) with is_creator flag
		return u.userTopicRepo.Create(txCtx, &entity.UserTopic{
			UserID:              userID,
			TopicID:             topic.ID,
			NotificationEnabled: true,
			IsCreator:           isNewTopic,
		})
	}); err != nil {
		return nil, err
	}

	topic.NotificationEnabled = true

	// Best effort: record activity
	RecordActivity(ctx, u.activityRepo, userID, entity.ActivityTopicAdded, input.Name, "トピック「"+input.Name+"」を追加しました")

	// Best effort: fetch initial data for brand-new topics so the detail page isn't empty.
	if isNewTopic && topic.ID != "" && len(input.Keywords) > 0 && u.bearerToken != "" {
		u.collectInitialData(ctx, topic)
	}

	return topic, nil
}

// collectInitialData fetches tweet volumes for a brand-new topic
// so the topic detail page has data immediately after creation.
// All errors are logged and swallowed — initial data fetch must never fail the CreateTopic call.
func (u *TopicUsecase) collectInitialData(ctx context.Context, topic *entity.Topic) {
	now := time.Now().UTC()

	// Fetch tweet volumes (last 6 hours, hourly)
	volumeQuery := strings.Join(topic.Keywords, " OR ")
	startTime := now.Add(-6 * time.Hour).Truncate(time.Hour)
	points, err := u.twitterGW.GetTweetCounts(ctx, u.bearerToken, volumeQuery, startTime)
	if err != nil {
		slog.Warn("initial tweet count fetch failed", "topic_id", topic.ID, "error", err)
	} else if len(points) > 0 {
		volumes := make([]*entity.TopicVolume, len(points))
		var sum int
		for i, p := range points {
			volumes[i] = &entity.TopicVolume{
				TopicID:   topic.ID,
				Timestamp: p.Start,
				Value:     int32(p.TweetCount),
			}
			sum += p.TweetCount
		}
		if err := u.topicVolumeRepo.BulkCreate(ctx, volumes); err != nil {
			slog.Warn("initial volume bulk create failed", "topic_id", topic.ID, "error", err)
		} else {
			slog.Info("initial volumes collected", "topic_id", topic.ID, "count", len(volumes))

			// Update topic with initial volume stats so the detail page shows real numbers
			latest := volumes[len(volumes)-1].Value
			mean := int32(sum / len(points))
			topic.CurrentVolume = latest
			topic.BaselineVolume = mean
			if err := u.topicRepo.Update(ctx, topic); err != nil {
				slog.Warn("initial volume topic update failed", "topic_id", topic.ID, "error", err)
			}
		}
	}
}

func timePtr(t time.Time) *time.Time { return &t }

// AddGenre adds a genre to the user's selection. Idempotent — succeeds if already exists.
func (u *TopicUsecase) AddGenre(ctx context.Context, userID string, genreSlug string) error {
	// Resolve genre slug → genre ID
	genre, err := u.genreRepo.FindBySlug(ctx, genreSlug)
	if err != nil {
		return err
	}

	exists, err := u.userGenreRepo.ExistsByUserIDAndGenre(ctx, userID, genre.ID)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	return u.userGenreRepo.Create(ctx, &entity.UserGenre{
		UserID:  userID,
		GenreID: genre.ID,
	})
}

// RemoveGenre removes a genre and all associated user-topic links within a transaction.
func (u *TopicUsecase) RemoveGenre(ctx context.Context, userID string, genreSlug string) error {
	// Resolve genre slug → genre ID
	genre, err := u.genreRepo.FindBySlug(ctx, genreSlug)
	if err != nil {
		return err
	}

	return u.txManager.RunInTransaction(ctx, func(txCtx context.Context) error {
		if err := u.userTopicRepo.DeleteByUserIDAndGenre(txCtx, userID, genre.ID); err != nil {
			return err
		}
		return u.userGenreRepo.DeleteByUserIDAndGenre(txCtx, userID, genre.ID)
	})
}

// ListUserGenres returns the genre slugs selected by the user.
func (u *TopicUsecase) ListUserGenres(ctx context.Context, userID string) ([]string, error) {
	genres, err := u.userGenreRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]string, len(genres))
	for i, g := range genres {
		result[i] = g.GenreSlug
	}
	return result, nil
}

// ListGenres returns all genre master records.
func (u *TopicUsecase) ListGenres(ctx context.Context) ([]*entity.Genre, error) {
	return u.genreRepo.List(ctx)
}

// DeleteTopic removes the user-topic link (shared topic remains).
func (u *TopicUsecase) DeleteTopic(ctx context.Context, userID string, topicID string) error {
	topic, err := u.topicRepo.FindByIDForUser(ctx, topicID, userID)
	if err != nil {
		return err
	}

	if err := u.userTopicRepo.Delete(ctx, userID, topicID); err != nil {
		return err
	}

	// Best effort: record activity
	RecordActivity(ctx, u.activityRepo, userID, entity.ActivityTopicRemoved, topic.Name, "トピック「"+topic.Name+"」を削除しました")

	return nil
}

// UpdateTopicNotification updates the notification enabled flag for a user-topic link.
func (u *TopicUsecase) UpdateTopicNotification(ctx context.Context, userID string, topicID string, enabled bool) error {
	// Verify user-topic link exists
	if _, err := u.topicRepo.FindByIDForUser(ctx, topicID, userID); err != nil {
		return err
	}

	return u.userTopicRepo.UpdateNotificationEnabled(ctx, userID, topicID, enabled)
}

// SuggestTopicsInput holds the parameters for suggesting topics.
type SuggestTopicsInput struct {
	Query string
	Genre string
	Limit int
}

// SuggestTopics returns topic suggestions, excluding user's existing topics.
// If Query is empty and Genre is set, returns topics from that genre.
// If Query is set, performs pg_trgm fuzzy search across all genres.
func (u *TopicUsecase) SuggestTopics(ctx context.Context, userID string, input SuggestTopicsInput) ([]*entity.TopicSuggestion, error) {
	if input.Limit <= 0 {
		input.Limit = 10
	}

	excludeIDs, err := u.userTopicRepo.ListTopicIDsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if input.Query == "" && input.Genre != "" {
		return u.topicRepo.ListByGenreExcluding(ctx, input.Genre, excludeIDs, input.Limit)
	}

	return u.topicRepo.SuggestByName(ctx, input.Query, excludeIDs, input.Limit)
}

