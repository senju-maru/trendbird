package usecase

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/trendbird/backend/internal/domain/apperror"
	"github.com/trendbird/backend/internal/domain/entity"
	"github.com/trendbird/backend/internal/domain/gateway"
	"github.com/trendbird/backend/internal/domain/service"
)

// ---------------------------------------------------------------------------
// Mock implementations for TrendFetchUsecase tests
// ---------------------------------------------------------------------------

type mockTopicRepo struct {
	ListAllFn func(ctx context.Context) ([]*entity.Topic, error)
	UpdateFn  func(ctx context.Context, topic *entity.Topic) error

	mu       sync.Mutex
	Updated  []*entity.Topic
}

func (m *mockTopicRepo) FindByID(ctx context.Context, id string) (*entity.Topic, error) {
	return nil, nil
}
func (m *mockTopicRepo) FindByIDForUser(ctx context.Context, id string, userID string) (*entity.Topic, error) {
	return nil, nil
}
func (m *mockTopicRepo) FindByNameAndGenre(ctx context.Context, name string, genre string) (*entity.Topic, error) {
	return nil, nil
}
func (m *mockTopicRepo) ListAll(ctx context.Context) ([]*entity.Topic, error) {
	if m.ListAllFn != nil {
		return m.ListAllFn(ctx)
	}
	return nil, nil
}
func (m *mockTopicRepo) ListByUserID(ctx context.Context, userID string) ([]*entity.Topic, error) {
	return nil, nil
}
func (m *mockTopicRepo) GetLatestUpdatedAtByUserID(ctx context.Context, userID string) (*time.Time, error) {
	return nil, nil
}
func (m *mockTopicRepo) Create(ctx context.Context, topic *entity.Topic) error { return nil }
func (m *mockTopicRepo) Update(ctx context.Context, topic *entity.Topic) error {
	m.mu.Lock()
	m.Updated = append(m.Updated, topic)
	m.mu.Unlock()
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, topic)
	}
	return nil
}
func (m *mockTopicRepo) Delete(ctx context.Context, id string) error { return nil }
func (m *mockTopicRepo) SuggestByName(ctx context.Context, query string, excludeIDs []string, limit int) ([]*entity.TopicSuggestion, error) {
	return nil, nil
}
func (m *mockTopicRepo) ListByGenreExcluding(ctx context.Context, genreSlug string, excludeIDs []string, limit int) ([]*entity.TopicSuggestion, error) {
	return nil, nil
}

type mockTopicVolumeRepo struct {
	BulkCreateFn          func(ctx context.Context, volumes []*entity.TopicVolume) error
	ListByTopicIDAndRangeFn func(ctx context.Context, topicID string, from, to time.Time) ([]*entity.TopicVolume, error)

	mu            sync.Mutex
	CreatedVolumes []*entity.TopicVolume
}

func (m *mockTopicVolumeRepo) BulkCreate(ctx context.Context, volumes []*entity.TopicVolume) error {
	m.mu.Lock()
	m.CreatedVolumes = append(m.CreatedVolumes, volumes...)
	m.mu.Unlock()
	if m.BulkCreateFn != nil {
		return m.BulkCreateFn(ctx, volumes)
	}
	return nil
}
func (m *mockTopicVolumeRepo) ListByTopicIDAndRange(ctx context.Context, topicID string, from, to time.Time) ([]*entity.TopicVolume, error) {
	if m.ListByTopicIDAndRangeFn != nil {
		return m.ListByTopicIDAndRangeFn(ctx, topicID, from, to)
	}
	return nil, nil
}

type mockTwitterGW struct {
	GetTweetCountsFn func(ctx context.Context, accessToken string, query string, startTime time.Time) ([]gateway.TweetCountDataPoint, error)
}

func (m *mockTwitterGW) BuildAuthorizationURL(ctx context.Context) (*gateway.OAuthStartResult, error) {
	return nil, nil
}
func (m *mockTwitterGW) ExchangeCode(ctx context.Context, code string, codeVerifier string) (*gateway.OAuthTokenResponse, error) {
	return nil, nil
}
func (m *mockTwitterGW) RefreshToken(ctx context.Context, refreshToken string) (*gateway.OAuthTokenResponse, error) {
	return nil, nil
}
func (m *mockTwitterGW) GetUserInfo(ctx context.Context, accessToken string) (*gateway.TwitterUserInfo, error) {
	return nil, nil
}
func (m *mockTwitterGW) SearchRecentTweets(ctx context.Context, accessToken string, input gateway.SearchTweetsInput) ([]gateway.Tweet, error) {
	return nil, nil
}
func (m *mockTwitterGW) GetTweetCounts(ctx context.Context, accessToken string, query string, startTime time.Time) ([]gateway.TweetCountDataPoint, error) {
	if m.GetTweetCountsFn != nil {
		return m.GetTweetCountsFn(ctx, accessToken, query, startTime)
	}
	return nil, nil
}
func (m *mockTwitterGW) PostTweet(ctx context.Context, accessToken string, text string) (string, error) {
	return "", nil
}
func (m *mockTwitterGW) PostReply(_ context.Context, _ string, _ string, _ string) (string, error) {
	return "", nil
}
func (m *mockTwitterGW) DeleteTweet(ctx context.Context, accessToken string, tweetID string) error {
	return nil
}
func (m *mockTwitterGW) VerifyCredentials(ctx context.Context, accessToken string) error {
	return nil
}
func (m *mockTwitterGW) SendDirectMessage(ctx context.Context, accessToken string, recipientID string, text string) error {
	return nil
}
func (m *mockTwitterGW) RevokeToken(ctx context.Context, accessToken string, refreshToken string) error {
	return nil
}

// mockSpikeHistRepoForTrend is a minimal mock for SpikeHistoryRepository used in trend fetch tests.
type mockSpikeHistRepoForTrend struct {
	CreateFn func(ctx context.Context, history *entity.SpikeHistory) error

	mu      sync.Mutex
	Created []*entity.SpikeHistory
}

func (m *mockSpikeHistRepoForTrend) Create(ctx context.Context, h *entity.SpikeHistory) error {
	m.mu.Lock()
	m.Created = append(m.Created, h)
	m.mu.Unlock()
	if m.CreateFn != nil {
		return m.CreateFn(ctx, h)
	}
	return nil
}
func (m *mockSpikeHistRepoForTrend) ListByTopicID(ctx context.Context, topicID string) ([]*entity.SpikeHistory, error) {
	return nil, nil
}
func (m *mockSpikeHistRepoForTrend) CountByUserIDCurrentMonth(ctx context.Context, userID string) (int32, error) {
	return 0, nil
}
func (m *mockSpikeHistRepoForTrend) ListUnnotified(ctx context.Context) ([]*entity.SpikeHistory, error) {
	return nil, nil
}
func (m *mockSpikeHistRepoForTrend) ListUnnotifiedByStatus(ctx context.Context, status entity.TopicStatus) ([]*entity.SpikeHistory, error) {
	return nil, nil
}
func (m *mockSpikeHistRepoForTrend) MarkNotified(ctx context.Context, ids []string, at time.Time) error {
	return nil
}
func (m *mockSpikeHistRepoForTrend) ListByTopicIDsSince(ctx context.Context, topicIDs []string, since time.Time) ([]*entity.SpikeHistory, error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// Helper to build a standard usecase with defaults
// ---------------------------------------------------------------------------

func newTestTrendFetchUsecase(
	topicRepo *mockTopicRepo,
	volumeRepo *mockTopicVolumeRepo,
	spikeRepo *mockSpikeHistRepoForTrend,
	twitterGW *mockTwitterGW,
) *TrendFetchUsecase {
	return NewTrendFetchUsecase(
		topicRepo, volumeRepo, spikeRepo, nil,
		twitterGW, nil, service.NewZScoreService(), "test-bearer-token",
	)
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestTrendFetch_NoTopics(t *testing.T) {
	topicRepo := &mockTopicRepo{
		ListAllFn: func(ctx context.Context) ([]*entity.Topic, error) {
			return nil, nil
		},
	}
	twitterGW := &mockTwitterGW{}
	uc := newTestTrendFetchUsecase(topicRepo, &mockTopicVolumeRepo{}, &mockSpikeHistRepoForTrend{}, twitterGW)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTrendFetch_SkipsEmptyKeywords(t *testing.T) {
	var apiCalled bool
	topicRepo := &mockTopicRepo{
		ListAllFn: func(ctx context.Context) ([]*entity.Topic, error) {
			return []*entity.Topic{
				{ID: "t1", Name: "No Keywords", Keywords: nil, Status: entity.TopicStable},
			}, nil
		},
	}
	twitterGW := &mockTwitterGW{
		GetTweetCountsFn: func(ctx context.Context, _ string, _ string, _ time.Time) ([]gateway.TweetCountDataPoint, error) {
			apiCalled = true
			return nil, nil
		},
	}
	uc := newTestTrendFetchUsecase(topicRepo, &mockTopicVolumeRepo{}, &mockSpikeHistRepoForTrend{}, twitterGW)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if apiCalled {
		t.Error("expected API not to be called for topic with empty keywords")
	}
}

func TestTrendFetch_NormalFlow_UpdatesTopic(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Hour)

	topicRepo := &mockTopicRepo{
		ListAllFn: func(ctx context.Context) ([]*entity.Topic, error) {
			return []*entity.Topic{
				{ID: "t1", Name: "Go", Keywords: []string{"golang"}, Status: entity.TopicStable},
			}, nil
		},
	}

	twitterGW := &mockTwitterGW{
		GetTweetCountsFn: func(ctx context.Context, _ string, query string, _ time.Time) ([]gateway.TweetCountDataPoint, error) {
			return []gateway.TweetCountDataPoint{
				{Start: now.Add(-1 * time.Hour), End: now, TweetCount: 100},
			}, nil
		},
	}

	// Return enough historical data for z-score calculation
	volumeRepo := &mockTopicVolumeRepo{
		ListByTopicIDAndRangeFn: func(ctx context.Context, topicID string, from, to time.Time) ([]*entity.TopicVolume, error) {
			volumes := make([]*entity.TopicVolume, 24)
			for i := range volumes {
				volumes[i] = &entity.TopicVolume{
					TopicID:   topicID,
					Timestamp: now.Add(time.Duration(-24+i) * time.Hour),
					Value:     50,
				}
			}
			// Latest value different to show update
			volumes[23].Value = 50
			return volumes, nil
		},
	}

	spikeRepo := &mockSpikeHistRepoForTrend{}
	uc := newTestTrendFetchUsecase(topicRepo, volumeRepo, spikeRepo, twitterGW)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify topic was updated
	if len(topicRepo.Updated) != 1 {
		t.Fatalf("expected 1 topic update, got %d", len(topicRepo.Updated))
	}
	updated := topicRepo.Updated[0]
	if updated.ZScore == nil {
		t.Error("expected z-score to be set")
	}

	// No spike should be created (all values stable)
	if len(spikeRepo.Created) != 0 {
		t.Errorf("expected no spike history, got %d", len(spikeRepo.Created))
	}
}

func TestTrendFetch_SpikeDetection_CreatesSpikeHistory(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Hour)

	topicRepo := &mockTopicRepo{
		ListAllFn: func(ctx context.Context) ([]*entity.Topic, error) {
			return []*entity.Topic{
				{ID: "t1", Name: "AI News", Keywords: []string{"ChatGPT", "OpenAI"}, Status: entity.TopicStable},
			}, nil
		},
	}

	twitterGW := &mockTwitterGW{
		GetTweetCountsFn: func(ctx context.Context, _ string, _ string, _ time.Time) ([]gateway.TweetCountDataPoint, error) {
			return []gateway.TweetCountDataPoint{
				{Start: now.Add(-1 * time.Hour), End: now, TweetCount: 500},
			}, nil
		},
	}

	// Historical: all 100 except last which is 500 → z = (500-100)/0 → high z-score
	volumeRepo := &mockTopicVolumeRepo{
		ListByTopicIDAndRangeFn: func(ctx context.Context, topicID string, from, to time.Time) ([]*entity.TopicVolume, error) {
			volumes := make([]*entity.TopicVolume, 48)
			for i := range 47 {
				volumes[i] = &entity.TopicVolume{
					TopicID:   topicID,
					Timestamp: now.Add(time.Duration(-48+i) * time.Hour),
					Value:     100,
				}
			}
			// Spike value at the end
			volumes[47] = &entity.TopicVolume{
				TopicID:   topicID,
				Timestamp: now.Add(-1 * time.Hour),
				Value:     500,
			}
			return volumes, nil
		},
	}

	spikeRepo := &mockSpikeHistRepoForTrend{}
	uc := newTestTrendFetchUsecase(topicRepo, volumeRepo, spikeRepo, twitterGW)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Topic should be updated to Spike
	if len(topicRepo.Updated) != 1 {
		t.Fatalf("expected 1 topic update, got %d", len(topicRepo.Updated))
	}
	if topicRepo.Updated[0].Status != entity.TopicSpike {
		t.Errorf("expected Spike status, got %d", topicRepo.Updated[0].Status)
	}
	if topicRepo.Updated[0].SpikeStartedAt == nil {
		t.Error("expected SpikeStartedAt to be set")
	}

	// Spike history should be created
	if len(spikeRepo.Created) != 1 {
		t.Fatalf("expected 1 spike history, got %d", len(spikeRepo.Created))
	}
	if spikeRepo.Created[0].TopicID != "t1" {
		t.Errorf("expected topic ID t1, got %s", spikeRepo.Created[0].TopicID)
	}
}

func TestTrendFetch_SpikeEnd_ClearsSpikeStartedAt(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Hour)
	spikeStart := now.Add(-1 * time.Hour)

	topicRepo := &mockTopicRepo{
		ListAllFn: func(ctx context.Context) ([]*entity.Topic, error) {
			return []*entity.Topic{
				{
					ID: "t1", Name: "Test", Keywords: []string{"test"},
					Status: entity.TopicSpike, SpikeStartedAt: &spikeStart,
				},
			}, nil
		},
	}

	twitterGW := &mockTwitterGW{
		GetTweetCountsFn: func(ctx context.Context, _ string, _ string, _ time.Time) ([]gateway.TweetCountDataPoint, error) {
			return []gateway.TweetCountDataPoint{
				{Start: now.Add(-1 * time.Hour), End: now, TweetCount: 50},
			}, nil
		},
	}

	// All stable values → Stable status → should clear SpikeStartedAt
	volumeRepo := &mockTopicVolumeRepo{
		ListByTopicIDAndRangeFn: func(ctx context.Context, topicID string, from, to time.Time) ([]*entity.TopicVolume, error) {
			volumes := make([]*entity.TopicVolume, 24)
			for i := range volumes {
				volumes[i] = &entity.TopicVolume{
					TopicID:   topicID,
					Timestamp: now.Add(time.Duration(-24+i) * time.Hour),
					Value:     50,
				}
			}
			return volumes, nil
		},
	}

	spikeRepo := &mockSpikeHistRepoForTrend{}
	uc := newTestTrendFetchUsecase(topicRepo, volumeRepo, spikeRepo, twitterGW)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(topicRepo.Updated) != 1 {
		t.Fatalf("expected 1 topic update, got %d", len(topicRepo.Updated))
	}
	if topicRepo.Updated[0].SpikeStartedAt != nil {
		t.Error("expected SpikeStartedAt to be cleared")
	}
	if topicRepo.Updated[0].Status == entity.TopicSpike {
		t.Error("expected status to not be Spike")
	}

	// No new spike history should be created
	if len(spikeRepo.Created) != 0 {
		t.Errorf("expected no spike history, got %d", len(spikeRepo.Created))
	}
}

func TestTrendFetch_StableToRising_CreatesRisingHistory(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Hour)

	topicRepo := &mockTopicRepo{
		ListAllFn: func(ctx context.Context) ([]*entity.Topic, error) {
			return []*entity.Topic{
				{ID: "t1", Name: "Rising Topic", Keywords: []string{"rising"}, Status: entity.TopicStable},
			}, nil
		},
	}

	twitterGW := &mockTwitterGW{
		GetTweetCountsFn: func(ctx context.Context, _ string, _ string, _ time.Time) ([]gateway.TweetCountDataPoint, error) {
			return []gateway.TweetCountDataPoint{
				{Start: now.Add(-1 * time.Hour), End: now, TweetCount: 160},
			}, nil
		},
	}

	// 24 values of 80, 23 values of 120, latest (160) → z ≈ 2.74 → Rising
	volumeRepo := &mockTopicVolumeRepo{
		ListByTopicIDAndRangeFn: func(ctx context.Context, topicID string, from, to time.Time) ([]*entity.TopicVolume, error) {
			volumes := make([]*entity.TopicVolume, 48)
			for i := range 24 {
				volumes[i] = &entity.TopicVolume{
					TopicID:   topicID,
					Timestamp: now.Add(time.Duration(-48+i) * time.Hour),
					Value:     80,
				}
			}
			for i := 24; i < 47; i++ {
				volumes[i] = &entity.TopicVolume{
					TopicID:   topicID,
					Timestamp: now.Add(time.Duration(-48+i) * time.Hour),
					Value:     120,
				}
			}
			volumes[47] = &entity.TopicVolume{
				TopicID:   topicID,
				Timestamp: now.Add(-1 * time.Hour),
				Value:     160,
			}
			return volumes, nil
		},
	}

	spikeRepo := &mockSpikeHistRepoForTrend{}
	uc := newTestTrendFetchUsecase(topicRepo, volumeRepo, spikeRepo, twitterGW)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(topicRepo.Updated) != 1 {
		t.Fatalf("expected 1 topic update, got %d", len(topicRepo.Updated))
	}
	if topicRepo.Updated[0].Status != entity.TopicRising {
		t.Errorf("expected Rising status, got %d", topicRepo.Updated[0].Status)
	}

	// Rising history should be created for Stable → Rising
	if len(spikeRepo.Created) != 1 {
		t.Fatalf("expected 1 rising history, got %d", len(spikeRepo.Created))
	}
	if spikeRepo.Created[0].Status != entity.TopicRising {
		t.Errorf("expected TopicRising status, got %d", spikeRepo.Created[0].Status)
	}
	if spikeRepo.Created[0].TopicID != "t1" {
		t.Errorf("expected topic ID t1, got %s", spikeRepo.Created[0].TopicID)
	}
}

func TestTrendFetch_SpikeToRising_DoesNotCreateHistory(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Hour)
	spikeStart := now.Add(-2 * time.Hour)

	topicRepo := &mockTopicRepo{
		ListAllFn: func(ctx context.Context) ([]*entity.Topic, error) {
			return []*entity.Topic{
				{
					ID: "t1", Name: "Cooldown Topic", Keywords: []string{"cooldown"},
					Status: entity.TopicSpike, SpikeStartedAt: &spikeStart,
				},
			}, nil
		},
	}

	twitterGW := &mockTwitterGW{
		GetTweetCountsFn: func(ctx context.Context, _ string, _ string, _ time.Time) ([]gateway.TweetCountDataPoint, error) {
			return []gateway.TweetCountDataPoint{
				{Start: now.Add(-1 * time.Hour), End: now, TweetCount: 160},
			}, nil
		},
	}

	// Same volume data → Rising status, but previous was Spike → no history
	volumeRepo := &mockTopicVolumeRepo{
		ListByTopicIDAndRangeFn: func(ctx context.Context, topicID string, from, to time.Time) ([]*entity.TopicVolume, error) {
			volumes := make([]*entity.TopicVolume, 48)
			for i := range 24 {
				volumes[i] = &entity.TopicVolume{
					TopicID:   topicID,
					Timestamp: now.Add(time.Duration(-48+i) * time.Hour),
					Value:     80,
				}
			}
			for i := 24; i < 47; i++ {
				volumes[i] = &entity.TopicVolume{
					TopicID:   topicID,
					Timestamp: now.Add(time.Duration(-48+i) * time.Hour),
					Value:     120,
				}
			}
			volumes[47] = &entity.TopicVolume{
				TopicID:   topicID,
				Timestamp: now.Add(-1 * time.Hour),
				Value:     160,
			}
			return volumes, nil
		},
	}

	spikeRepo := &mockSpikeHistRepoForTrend{}
	uc := newTestTrendFetchUsecase(topicRepo, volumeRepo, spikeRepo, twitterGW)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(topicRepo.Updated) != 1 {
		t.Fatalf("expected 1 topic update, got %d", len(topicRepo.Updated))
	}
	if topicRepo.Updated[0].Status != entity.TopicRising {
		t.Errorf("expected Rising status, got %d", topicRepo.Updated[0].Status)
	}

	// No rising history should be created for Spike → Rising (cooldown)
	if len(spikeRepo.Created) != 0 {
		t.Errorf("expected no history for Spike→Rising, got %d", len(spikeRepo.Created))
	}
}

func TestTrendFetch_RateLimitBreaksLoop(t *testing.T) {
	callCount := 0
	topicRepo := &mockTopicRepo{
		ListAllFn: func(ctx context.Context) ([]*entity.Topic, error) {
			return []*entity.Topic{
				{ID: "t1", Name: "Topic 1", Keywords: []string{"kw1"}, Status: entity.TopicStable},
				{ID: "t2", Name: "Topic 2", Keywords: []string{"kw2"}, Status: entity.TopicStable},
				{ID: "t3", Name: "Topic 3", Keywords: []string{"kw3"}, Status: entity.TopicStable},
			}, nil
		},
	}

	twitterGW := &mockTwitterGW{
		GetTweetCountsFn: func(ctx context.Context, _ string, _ string, _ time.Time) ([]gateway.TweetCountDataPoint, error) {
			callCount++
			if callCount >= 2 {
				return nil, apperror.ResourceExhausted("twitter: rate limit exceeded")
			}
			return []gateway.TweetCountDataPoint{
				{Start: time.Now().Add(-1 * time.Hour), End: time.Now(), TweetCount: 100},
			}, nil
		},
	}

	volumeRepo := &mockTopicVolumeRepo{
		ListByTopicIDAndRangeFn: func(ctx context.Context, topicID string, from, to time.Time) ([]*entity.TopicVolume, error) {
			return []*entity.TopicVolume{
				{Value: 100}, {Value: 100}, {Value: 100},
			}, nil
		},
	}

	uc := newTestTrendFetchUsecase(topicRepo, volumeRepo, &mockSpikeHistRepoForTrend{}, twitterGW)

	// Should succeed partially (first topic processed, then rate limit)
	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 API calls before rate limit, got %d", callCount)
	}
}

func TestTrendFetch_InsufficientHistory_SkipsStatusUpdate(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Hour)

	topicRepo := &mockTopicRepo{
		ListAllFn: func(ctx context.Context) ([]*entity.Topic, error) {
			return []*entity.Topic{
				{ID: "t1", Name: "New Topic", Keywords: []string{"new"}, Status: entity.TopicStable},
			}, nil
		},
	}

	twitterGW := &mockTwitterGW{
		GetTweetCountsFn: func(ctx context.Context, _ string, _ string, _ time.Time) ([]gateway.TweetCountDataPoint, error) {
			return []gateway.TweetCountDataPoint{
				{Start: now.Add(-1 * time.Hour), End: now, TweetCount: 100},
			}, nil
		},
	}

	// Only 1 data point → insufficient for z-score
	volumeRepo := &mockTopicVolumeRepo{
		ListByTopicIDAndRangeFn: func(ctx context.Context, topicID string, from, to time.Time) ([]*entity.TopicVolume, error) {
			return []*entity.TopicVolume{
				{TopicID: topicID, Timestamp: now.Add(-1 * time.Hour), Value: 100},
			}, nil
		},
	}

	uc := newTestTrendFetchUsecase(topicRepo, volumeRepo, &mockSpikeHistRepoForTrend{}, twitterGW)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Volume should be created but topic should NOT be updated
	if len(volumeRepo.CreatedVolumes) != 1 {
		t.Errorf("expected 1 volume created, got %d", len(volumeRepo.CreatedVolumes))
	}
	if len(topicRepo.Updated) != 0 {
		t.Errorf("expected 0 topic updates (insufficient data), got %d", len(topicRepo.Updated))
	}
}

func TestTrendFetch_AllTopicsFail_ReturnsError(t *testing.T) {
	topicRepo := &mockTopicRepo{
		ListAllFn: func(ctx context.Context) ([]*entity.Topic, error) {
			return []*entity.Topic{
				{ID: "t1", Name: "Topic 1", Keywords: []string{"kw1"}, Status: entity.TopicStable},
			}, nil
		},
	}

	twitterGW := &mockTwitterGW{
		GetTweetCountsFn: func(ctx context.Context, _ string, _ string, _ time.Time) ([]gateway.TweetCountDataPoint, error) {
			return nil, fmt.Errorf("network error")
		},
	}

	uc := newTestTrendFetchUsecase(topicRepo, &mockTopicVolumeRepo{}, &mockSpikeHistRepoForTrend{}, twitterGW)

	err := uc.Execute(context.Background())
	if err == nil {
		t.Fatal("expected error when all topics fail")
	}
}
