package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/trendbird/backend/internal/domain/entity"
)

func TestRisingNotificationUsecase_Execute_NoRisingRecords(t *testing.T) {
	uc := NewRisingNotificationUsecase(
		&mockSpikeHistoryRepo{
			ListUnnotifiedByStatusFn: func(ctx context.Context, status entity.TopicStatus) ([]*entity.SpikeHistory, error) {
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

func TestRisingNotificationUsecase_Execute_SendsNotification(t *testing.T) {
	topicID := "topic-1"
	userID := "user-1"
	risingID := "rising-1"

	spikeRepo := &mockSpikeHistoryRepo{
		ListUnnotifiedByStatusFn: func(ctx context.Context, status entity.TopicStatus) ([]*entity.SpikeHistory, error) {
			if status != entity.TopicRising {
				t.Errorf("expected TopicRising status, got %d", status)
			}
			return []*entity.SpikeHistory{
				{
					ID:              risingID,
					TopicID:         topicID,
					TopicName:       "AI Trends",
					Timestamp:       time.Now(),
					PeakZScore:      2.5,
					Status:          entity.TopicRising,
					Summary:         "AI Trends rising trend",
					DurationMinutes: 0,
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
			userID: {UserID: userID, RisingEnabled: true},
		},
	}

	notiRepo := &mockNotificationRepo{}

	uc := NewRisingNotificationUsecase(
		spikeRepo, userTopicRepo, notiSettingRepo, notiRepo,
	)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify in-app notification was created
	if len(notiRepo.CreatedNotifications) != 1 {
		t.Fatalf("expected 1 notification, got %d", len(notiRepo.CreatedNotifications))
	}
	if notiRepo.CreatedNotifications[0].Title != "AI Trends で上昇トレンドを検知" {
		t.Errorf("unexpected notification title: %s", notiRepo.CreatedNotifications[0].Title)
	}

	// Verify rising was marked as notified
	if len(spikeRepo.MarkedIDs) != 1 || spikeRepo.MarkedIDs[0] != risingID {
		t.Errorf("expected marked IDs [%s], got %v", risingID, spikeRepo.MarkedIDs)
	}
}

func TestRisingNotificationUsecase_Execute_SkipsUserWithRisingDisabled(t *testing.T) {
	topicID := "topic-1"
	userID := "user-1"

	spikeRepo := &mockSpikeHistoryRepo{
		ListUnnotifiedByStatusFn: func(ctx context.Context, status entity.TopicStatus) ([]*entity.SpikeHistory, error) {
			return []*entity.SpikeHistory{
				{
					ID:        "rising-1",
					TopicID:   topicID,
					TopicName: "Topic",
					Timestamp: time.Now(),
					PeakZScore: 2.0,
					Status:    entity.TopicRising,
					Summary:   "Test rising",
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
			userID: {UserID: userID, RisingEnabled: false},
		},
	}

	notiRepo := &mockNotificationRepo{}

	uc := NewRisingNotificationUsecase(
		spikeRepo, userTopicRepo, notiSettingRepo, notiRepo,
	)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// No in-app notification should be created (rising_enabled=false)
	if len(notiRepo.CreatedNotifications) != 0 {
		t.Errorf("expected 0 in-app notifications, got %d", len(notiRepo.CreatedNotifications))
	}
}
