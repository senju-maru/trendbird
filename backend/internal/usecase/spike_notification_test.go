package usecase

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/trendbird/backend/internal/domain/entity"
)

// ---------------------------------------------------------------------------
// Mock implementations for SpikeNotificationUsecase tests
// ---------------------------------------------------------------------------

type mockSpikeHistoryRepo struct {
	ListUnnotifiedFn         func(ctx context.Context) ([]*entity.SpikeHistory, error)
	ListUnnotifiedByStatusFn func(ctx context.Context, status entity.TopicStatus) ([]*entity.SpikeHistory, error)
	MarkNotifiedFn      func(ctx context.Context, ids []string, at time.Time) error
	ListByTopicIDsSinceFn    func(ctx context.Context, topicIDs []string, since time.Time) ([]*entity.SpikeHistory, error)

	MarkedIDs []string
	mu        sync.Mutex
}

func (m *mockSpikeHistoryRepo) Create(ctx context.Context, h *entity.SpikeHistory) error { return nil }
func (m *mockSpikeHistoryRepo) ListByTopicID(ctx context.Context, topicID string) ([]*entity.SpikeHistory, error) {
	return nil, nil
}
func (m *mockSpikeHistoryRepo) CountByUserIDCurrentMonth(ctx context.Context, userID string) (int32, error) {
	return 0, nil
}
func (m *mockSpikeHistoryRepo) ListUnnotified(ctx context.Context) ([]*entity.SpikeHistory, error) {
	if m.ListUnnotifiedFn != nil {
		return m.ListUnnotifiedFn(ctx)
	}
	return nil, nil
}
func (m *mockSpikeHistoryRepo) ListUnnotifiedByStatus(ctx context.Context, status entity.TopicStatus) ([]*entity.SpikeHistory, error) {
	if m.ListUnnotifiedByStatusFn != nil {
		return m.ListUnnotifiedByStatusFn(ctx, status)
	}
	// Fallback to ListUnnotifiedFn for backward compatibility
	if m.ListUnnotifiedFn != nil {
		return m.ListUnnotifiedFn(ctx)
	}
	return nil, nil
}
func (m *mockSpikeHistoryRepo) MarkNotified(ctx context.Context, ids []string, at time.Time) error {
	m.mu.Lock()
	m.MarkedIDs = append(m.MarkedIDs, ids...)
	m.mu.Unlock()
	if m.MarkNotifiedFn != nil {
		return m.MarkNotifiedFn(ctx, ids, at)
	}
	return nil
}
func (m *mockSpikeHistoryRepo) ListByTopicIDsSince(ctx context.Context, topicIDs []string, since time.Time) ([]*entity.SpikeHistory, error) {
	if m.ListByTopicIDsSinceFn != nil {
		return m.ListByTopicIDsSinceFn(ctx, topicIDs, since)
	}
	return nil, nil
}

type mockUserTopicRepo struct {
	ListUserIDsByTopicIDFn func(ctx context.Context, topicID string, notificationEnabledOnly bool) ([]string, error)
	ListTopicIDsByUserIDFn func(ctx context.Context, userID string) ([]string, error)
}

func (m *mockUserTopicRepo) Create(ctx context.Context, ut *entity.UserTopic) error { return nil }
func (m *mockUserTopicRepo) Delete(ctx context.Context, userID, topicID string) error { return nil }
func (m *mockUserTopicRepo) DeleteByUserIDAndGenre(ctx context.Context, userID, genreID string) error {
	return nil
}
func (m *mockUserTopicRepo) Exists(ctx context.Context, userID, topicID string) (bool, error) {
	return false, nil
}
func (m *mockUserTopicRepo) CountByUserID(ctx context.Context, userID string) (int, error) {
	return 0, nil
}
func (m *mockUserTopicRepo) CountCreatorByUserID(ctx context.Context, userID string) (int, error) {
	return 0, nil
}
func (m *mockUserTopicRepo) UpdateNotificationEnabled(ctx context.Context, userID, topicID string, enabled bool) error {
	return nil
}
func (m *mockUserTopicRepo) ListUserIDsByTopicID(ctx context.Context, topicID string, notificationEnabledOnly bool) ([]string, error) {
	return m.ListUserIDsByTopicIDFn(ctx, topicID, notificationEnabledOnly)
}
func (m *mockUserTopicRepo) ListTopicIDsByUserID(ctx context.Context, userID string) ([]string, error) {
	if m.ListTopicIDsByUserIDFn != nil {
		return m.ListTopicIDsByUserIDFn(ctx, userID)
	}
	return nil, nil
}

type mockNotiSettingRepo struct {
	settings map[string]*entity.NotificationSetting
}

func (m *mockNotiSettingRepo) FindByUserID(ctx context.Context, userID string) (*entity.NotificationSetting, error) {
	s, ok := m.settings[userID]
	if !ok {
		return &entity.NotificationSetting{UserID: userID, SpikeEnabled: true}, nil
	}
	return s, nil
}
func (m *mockNotiSettingRepo) Upsert(ctx context.Context, setting *entity.NotificationSetting) error {
	return nil
}

type mockNotificationRepo struct {
	CreatedNotifications []*entity.Notification
	CreatedUserIDs       [][]string
	mu                   sync.Mutex
	GetLastNotifiedAtFn  func(ctx context.Context, userID string, notiType entity.NotificationType) (time.Time, error)
}

func (m *mockNotificationRepo) Create(ctx context.Context, n *entity.Notification) error { return nil }
func (m *mockNotificationRepo) CreateForUsers(ctx context.Context, n *entity.Notification, userIDs []string) error {
	m.mu.Lock()
	m.CreatedNotifications = append(m.CreatedNotifications, n)
	m.CreatedUserIDs = append(m.CreatedUserIDs, userIDs)
	m.mu.Unlock()
	return nil
}
func (m *mockNotificationRepo) ListByUserID(ctx context.Context, userID string, limit, offset int) ([]*entity.Notification, int64, error) {
	return nil, 0, nil
}
func (m *mockNotificationRepo) MarkAsRead(ctx context.Context, userID, id string) error { return nil }
func (m *mockNotificationRepo) MarkAllAsReadByUserID(ctx context.Context, userID string) error {
	return nil
}
func (m *mockNotificationRepo) CountUnreadByUserID(ctx context.Context, userID string) (int32, error) {
	return 0, nil
}
func (m *mockNotificationRepo) GetLastNotifiedAt(ctx context.Context, userID string, notiType entity.NotificationType) (time.Time, error) {
	if m.GetLastNotifiedAtFn != nil {
		return m.GetLastNotifiedAtFn(ctx, userID, notiType)
	}
	return time.Time{}, nil
}

type mockUserRepo struct {
	users map[string]*entity.User
}

func (m *mockUserRepo) FindByID(ctx context.Context, id string) (*entity.User, error) {
	return m.users[id], nil
}
func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	return nil, nil
}
func (m *mockUserRepo) FindByTwitterID(ctx context.Context, twitterID string) (*entity.User, error) {
	return nil, nil
}
func (m *mockUserRepo) UpsertByTwitterID(ctx context.Context, input entity.UpsertUserInput) (*entity.User, error) {
	return nil, nil
}
func (m *mockUserRepo) UpdateEmail(ctx context.Context, id, email string) error { return nil }
func (m *mockUserRepo) CompleteTutorial(ctx context.Context, id string) error  { return nil }
func (m *mockUserRepo) Delete(ctx context.Context, id string) error              { return nil }
func (m *mockUserRepo) ListByIDs(ctx context.Context, ids []string) ([]*entity.User, error) {
	var result []*entity.User
	for _, id := range ids {
		if u, ok := m.users[id]; ok {
			result = append(result, u)
		}
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestSpikeNotificationUsecase_Execute_NoSpikes(t *testing.T) {
	uc := NewSpikeNotificationUsecase(
		&mockSpikeHistoryRepo{
			ListUnnotifiedFn: func(ctx context.Context) ([]*entity.SpikeHistory, error) {
				return nil, nil
			},
		},
		&mockUserTopicRepo{},
		&mockNotiSettingRepo{settings: map[string]*entity.NotificationSetting{}},
		&mockNotificationRepo{},
	)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSpikeNotificationUsecase_Execute_SendsNotification(t *testing.T) {
	topicID := "topic-1"
	userID := "user-1"
	spikeID := "spike-1"

	spikeRepo := &mockSpikeHistoryRepo{
		ListUnnotifiedFn: func(ctx context.Context) ([]*entity.SpikeHistory, error) {
			return []*entity.SpikeHistory{
				{
					ID:              spikeID,
					TopicID:         topicID,
					TopicName:       "Go 1.26",
					Timestamp:       time.Now(),
					PeakZScore:      4.2,
					Status:          entity.TopicSpike,
					Summary:         "Go 1.26 release spike",
					DurationMinutes: 30,
				},
			}, nil
		},
	}

	userTopicRepo := &mockUserTopicRepo{
		ListUserIDsByTopicIDFn: func(ctx context.Context, tid string, _ bool) ([]string, error) {
			if tid == topicID {
				return []string{userID}, nil
			}
			return nil, nil
		},
	}

	notiSettingRepo := &mockNotiSettingRepo{
		settings: map[string]*entity.NotificationSetting{
			userID: {UserID: userID, SpikeEnabled: true},
		},
	}

	notiRepo := &mockNotificationRepo{}

	uc := NewSpikeNotificationUsecase(
		spikeRepo, userTopicRepo, notiSettingRepo, notiRepo,
	)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify in-app notification was created
	if len(notiRepo.CreatedNotifications) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notiRepo.CreatedNotifications))
	}
	if notiRepo.CreatedNotifications[0].Title != "Go 1.26 でスパイクを検知" {
		t.Errorf("unexpected notification title: %s", notiRepo.CreatedNotifications[0].Title)
	}

	// Verify spike was marked as notified
	if len(spikeRepo.MarkedIDs) != 1 || spikeRepo.MarkedIDs[0] != spikeID {
		t.Errorf("expected marked IDs [%s], got %v", spikeID, spikeRepo.MarkedIDs)
	}
}

func TestSpikeNotificationUsecase_Execute_SendsNotificationWhenSpikeEnabled(t *testing.T) {
	topicID := "topic-1"
	userID := "user-1"

	spikeRepo := &mockSpikeHistoryRepo{
		ListUnnotifiedFn: func(ctx context.Context) ([]*entity.SpikeHistory, error) {
			return []*entity.SpikeHistory{
				{
					ID:        "spike-1",
					TopicID:   topicID,
					TopicName: "Topic",
					Timestamp: time.Now(),
					PeakZScore: 3.0,
					Status:    entity.TopicSpike,
					Summary:   "Test spike",
				},
			}, nil
		},
	}

	userTopicRepo := &mockUserTopicRepo{
		ListUserIDsByTopicIDFn: func(ctx context.Context, tid string, _ bool) ([]string, error) {
			return []string{userID}, nil
		},
	}

	notiSettingRepo := &mockNotiSettingRepo{
		settings: map[string]*entity.NotificationSetting{
			userID: {UserID: userID, SpikeEnabled: true},
		},
	}

	notiRepo := &mockNotificationRepo{}

	uc := NewSpikeNotificationUsecase(
		spikeRepo, userTopicRepo, notiSettingRepo, notiRepo,
	)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// In-app notification should be created (spike_enabled=true)
	if len(notiRepo.CreatedNotifications) != 1 {
		t.Fatalf("expected 1 in-app notification, got %d", len(notiRepo.CreatedNotifications))
	}
}

