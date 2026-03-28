package e2etest

import (
	"context"
	"testing"
	"time"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/gateway"
	"github.com/trendbird/backend/internal/domain/service"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
	"github.com/trendbird/backend/internal/infrastructure/persistence/repository"
	"github.com/trendbird/backend/internal/usecase"
)

func TestTrendFetchBatch_SpikeDetected(t *testing.T) {
	env := setupTest(t)

	topic := seedTopic(t, env.db, withTopicName("Spike Trend"), withTopicKeywords([]string{"spike-kw"}))

	// Seed baseline volumes (stable history: 10 tweets/hour for 20 data points)
	// 20 data points at value=10 ensures z-score > 3.0 when spike=100
	now := time.Now().UTC()
	for i := 20; i >= 1; i-- {
		seedTopicVolume(t, env.db, topic.ID,
			withTopicVolumeTimestamp(now.Add(time.Duration(-i)*time.Hour)),
			withTopicVolumeValue(10),
		)
	}

	// Mock: GetTweetCounts returns high volume (spike)
	env.mockTwitter.GetTweetCountsFn = func(_ context.Context, _ string, _ string, _ time.Time) ([]gateway.TweetCountDataPoint, error) {
		return []gateway.TweetCountDataPoint{
			{Start: now.Add(-30 * time.Minute), End: now, TweetCount: 100},
		}, nil
	}

	topicRepo := repository.NewTopicRepository(env.db)
	topicVolumeRepo := repository.NewTopicVolumeRepository(env.db)
	spikeHistRepo := repository.NewSpikeHistoryRepository(env.db)
	zscoreSvc := service.NewZScoreService()

	uc := usecase.NewTrendFetchUsecase(topicRepo, topicVolumeRepo, spikeHistRepo, nil, env.mockTwitter, env.mockAI, zscoreSvc, "test-bearer")

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify topic status updated to Spike (1)
	var dbTopic model.Topic
	env.db.First(&dbTopic, "id = ?", topic.ID)
	if dbTopic.Status != 1 {
		t.Errorf("topic status: want 1 (Spike), got %d", dbTopic.Status)
	}

	// Verify SpikeHistory created
	var shCount int64
	env.db.Model(&model.SpikeHistory{}).Where("topic_id = ?", topic.ID).Count(&shCount)
	if shCount == 0 {
		t.Error("expected SpikeHistory to be created")
	}
}

func TestTrendFetchBatch_StableNoChange(t *testing.T) {
	env := setupTest(t)

	topic := seedTopic(t, env.db, withTopicName("Stable Trend"), withTopicKeywords([]string{"stable-kw"}))

	// Seed baseline volumes
	now := time.Now().UTC()
	for i := 5; i >= 1; i-- {
		seedTopicVolume(t, env.db, topic.ID,
			withTopicVolumeTimestamp(now.Add(time.Duration(-i)*time.Hour)),
			withTopicVolumeValue(100),
		)
	}

	// Mock: GetTweetCounts returns normal volume
	env.mockTwitter.GetTweetCountsFn = func(_ context.Context, _ string, _ string, _ time.Time) ([]gateway.TweetCountDataPoint, error) {
		return []gateway.TweetCountDataPoint{
			{Start: now.Add(-1 * time.Hour), End: now, TweetCount: 105},
		}, nil
	}

	topicRepo := repository.NewTopicRepository(env.db)
	topicVolumeRepo := repository.NewTopicVolumeRepository(env.db)
	spikeHistRepo := repository.NewSpikeHistoryRepository(env.db)
	zscoreSvc := service.NewZScoreService()

	uc := usecase.NewTrendFetchUsecase(topicRepo, topicVolumeRepo, spikeHistRepo, nil, env.mockTwitter, env.mockAI, zscoreSvc, "test-bearer")

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify topic status stays Stable (3)
	var dbTopic model.Topic
	env.db.First(&dbTopic, "id = ?", topic.ID)
	if dbTopic.Status != 3 {
		t.Errorf("topic status: want 3 (Stable), got %d", dbTopic.Status)
	}

	// Verify no SpikeHistory created
	var shCount int64
	env.db.Model(&model.SpikeHistory{}).Where("topic_id = ?", topic.ID).Count(&shCount)
	if shCount != 0 {
		t.Errorf("expected 0 SpikeHistory, got %d", shCount)
	}
}

func TestTrendFetchBatch_NoTopics(t *testing.T) {
	env := setupTest(t)

	topicRepo := repository.NewTopicRepository(env.db)
	topicVolumeRepo := repository.NewTopicVolumeRepository(env.db)
	spikeHistRepo := repository.NewSpikeHistoryRepository(env.db)
	zscoreSvc := service.NewZScoreService()

	uc := usecase.NewTrendFetchUsecase(topicRepo, topicVolumeRepo, spikeHistRepo, nil, env.mockTwitter, env.mockAI, zscoreSvc, "test-bearer")

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}

func TestTrendFetchBatch_RateLimit(t *testing.T) {
	env := setupTest(t)

	// Create 2 topics
	topic1 := seedTopic(t, env.db, withTopicName("RL Topic 1"), withTopicKeywords([]string{"rl1"}))
	seedTopic(t, env.db, withTopicName("RL Topic 2"), withTopicKeywords([]string{"rl2"}))

	now := time.Now().UTC()
	for i := 5; i >= 1; i-- {
		seedTopicVolume(t, env.db, topic1.ID,
			withTopicVolumeTimestamp(now.Add(time.Duration(-i)*time.Hour)),
			withTopicVolumeValue(10),
		)
	}

	// Mock: first call succeeds, second call returns rate limit error
	callCount := 0
	env.mockTwitter.GetTweetCountsFn = func(_ context.Context, _ string, _ string, _ time.Time) ([]gateway.TweetCountDataPoint, error) {
		callCount++
		if callCount > 1 {
			return nil, apperror.ResourceExhausted("rate limit exceeded")
		}
		return []gateway.TweetCountDataPoint{
			{Start: now.Add(-1 * time.Hour), End: now, TweetCount: 10},
		}, nil
	}

	topicRepo := repository.NewTopicRepository(env.db)
	topicVolumeRepo := repository.NewTopicVolumeRepository(env.db)
	spikeHistRepo := repository.NewSpikeHistoryRepository(env.db)
	zscoreSvc := service.NewZScoreService()

	uc := usecase.NewTrendFetchUsecase(topicRepo, topicVolumeRepo, spikeHistRepo, nil, env.mockTwitter, env.mockAI, zscoreSvc, "test-bearer")

	// Should not return error (partial success is OK)
	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
}

func TestTrendFetchBatch_SpikeStartedAtTracking(t *testing.T) {
	env := setupTest(t)

	topic := seedTopic(t, env.db, withTopicName("SpikeStart Track"), withTopicKeywords([]string{"track-kw"}))

	// Seed 20 baseline data points for reliable z-score spike detection
	now := time.Now().UTC()
	for i := 20; i >= 1; i-- {
		seedTopicVolume(t, env.db, topic.ID,
			withTopicVolumeTimestamp(now.Add(time.Duration(-i)*time.Hour)),
			withTopicVolumeValue(10),
		)
	}

	// Mock: GetTweetCounts returns high volume to trigger spike
	env.mockTwitter.GetTweetCountsFn = func(_ context.Context, _ string, _ string, _ time.Time) ([]gateway.TweetCountDataPoint, error) {
		return []gateway.TweetCountDataPoint{
			{Start: now.Add(-30 * time.Minute), End: now, TweetCount: 100},
		}, nil
	}

	topicRepo := repository.NewTopicRepository(env.db)
	topicVolumeRepo := repository.NewTopicVolumeRepository(env.db)
	spikeHistRepo := repository.NewSpikeHistoryRepository(env.db)
	zscoreSvc := service.NewZScoreService()

	uc := usecase.NewTrendFetchUsecase(topicRepo, topicVolumeRepo, spikeHistRepo, nil, env.mockTwitter, env.mockAI, zscoreSvc, "test-bearer")

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify SpikeStartedAt is set
	var dbTopic model.Topic
	env.db.First(&dbTopic, "id = ?", topic.ID)
	if dbTopic.SpikeStartedAt == nil {
		t.Error("expected SpikeStartedAt to be set on Stable→Spike transition")
	}
}
