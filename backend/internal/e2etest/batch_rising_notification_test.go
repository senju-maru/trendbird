package e2etest

import (
	"context"
	"testing"
	"time"

	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
	"github.com/trendbird/backend/internal/infrastructure/persistence/repository"
	"github.com/trendbird/backend/internal/usecase"
)

func TestRisingNotificationBatch_E2E(t *testing.T) {
	env := setupTest(t)

	// Seed data: user → topic → user_topic → notification_setting → spike_history (status=2 Rising)
	user := seedUser(t, env.db, withEmail("rising-test@example.com"))
	topic := seedTopic(t, env.db, withTopicName("Rust 2.0"))
	seedUserTopic(t, env.db, user.ID, topic.ID, withNotificationEnabled(true))
	seedNotificationSetting(t, env.db, user.ID, withRisingEnabled(true))

	// Insert unnotified rising history (status=2)
	rising := seedSpikeHistory(t, env.db, topic.ID,
		withSpikeStatus(2),
		withSpikePeakZScore(3.2),
		withSpikeSummary("Rust 2.0 rising trend detected"),
	)

	// Verify notified_at is NULL
	var sh model.SpikeHistory
	if err := env.db.First(&sh, "id = ?", rising.ID).Error; err != nil {
		t.Fatalf("failed to find spike history: %v", err)
	}
	if sh.NotifiedAt != nil {
		t.Fatalf("expected notified_at to be NULL, got %v", sh.NotifiedAt)
	}

	// Build usecase with real repos
	spikeHistRepo := repository.NewSpikeHistoryRepository(env.db)
	userTopicRepo := repository.NewUserTopicRepository(env.db)
	notiSettingRepo := repository.NewNotificationSettingRepository(env.db)
	notiRepo := repository.NewNotificationRepository(env.db)
	uc := usecase.NewRisingNotificationUsecase(
		spikeHistRepo, userTopicRepo, notiSettingRepo, notiRepo,
	)

	ctx := context.Background()
	if err := uc.Execute(ctx); err != nil {
		t.Fatalf("batch execute failed: %v", err)
	}

	// Verify notified_at is now set
	if err := env.db.First(&sh, "id = ?", rising.ID).Error; err != nil {
		t.Fatalf("failed to find spike history after batch: %v", err)
	}
	if sh.NotifiedAt == nil {
		t.Fatal("expected notified_at to be set after batch, but it's still NULL")
	}
	if time.Since(*sh.NotifiedAt) > 5*time.Minute {
		t.Errorf("notified_at too old: %v", sh.NotifiedAt)
	}

	// Verify in-app notification was created
	var unCount int64
	env.db.Raw("SELECT COUNT(*) FROM user_notifications WHERE user_id = ?", user.ID).Scan(&unCount)
	if unCount == 0 {
		t.Fatal("expected at least 1 in-app notification to be created")
	}

	// Verify idempotency: running again should not create more notifications
	beforeCount := unCount

	if err := uc.Execute(ctx); err != nil {
		t.Fatalf("second batch execute failed: %v", err)
	}

	var afterCount int64
	env.db.Raw("SELECT COUNT(*) FROM user_notifications WHERE user_id = ?", user.ID).Scan(&afterCount)
	if afterCount != beforeCount {
		t.Errorf("expected no additional notifications on re-run, but got %d more", afterCount-beforeCount)
	}
}

func TestRisingNotificationBatch_NoCooldownSuppression(t *testing.T) {
	// OSS版: クールダウン=0 のため、直前の通知があっても抑制されない
	env := setupTest(t)

	user := seedUser(t, env.db, withEmail("cooldown-rising@example.com"))
	topic := seedTopic(t, env.db, withTopicName("Rising Cooldown Topic"))
	seedUserTopic(t, env.db, user.ID, topic.ID, withNotificationEnabled(true))
	seedNotificationSetting(t, env.db, user.ID, withRisingEnabled(true))

	// 30分前の通知を作成（クールダウン=0 なので抑制されない）
	noti := seedNotification(t, env.db, user.ID, withNotificationType(1)) // type=1 Trend
	env.db.Exec("UPDATE notifications SET created_at = ? WHERE id = ?", time.Now().Add(-30*time.Minute), noti.ID)

	// 未通知の rising を作成 (status=2)
	rising := seedSpikeHistory(t, env.db, topic.ID,
		withSpikeStatus(2),
		withSpikePeakZScore(3.2),
		withSpikeSummary("Rising cooldown test"),
	)

	spikeHistRepo := repository.NewSpikeHistoryRepository(env.db)
	userTopicRepo := repository.NewUserTopicRepository(env.db)
	notiSettingRepo := repository.NewNotificationSettingRepository(env.db)
	notiRepo := repository.NewNotificationRepository(env.db)
	uc := usecase.NewRisingNotificationUsecase(
		spikeHistRepo, userTopicRepo, notiSettingRepo, notiRepo,
	)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("batch execute failed: %v", err)
	}

	// rising は処理済み（notified_at が設定される）
	var sh model.SpikeHistory
	if err := env.db.First(&sh, "id = ?", rising.ID).Error; err != nil {
		t.Fatalf("failed to find spike history: %v", err)
	}
	if sh.NotifiedAt == nil {
		t.Fatal("expected notified_at to be set")
	}

	// 新規 user_notification が作成されている（seed の1件 + バッチの1件 = 2件）
	var unCount int64
	env.db.Raw("SELECT COUNT(*) FROM user_notifications WHERE user_id = ?", user.ID).Scan(&unCount)
	if unCount < 2 {
		t.Errorf("user_notifications: want >= 2, got %d", unCount)
	}
}

func TestRisingNotificationBatch_SkipsWhenRisingDisabled(t *testing.T) {
	env := setupTest(t)

	// Seed user with rising_enabled=false
	user := seedUser(t, env.db, withEmail("no-rising@example.com"))
	topic := seedTopic(t, env.db, withTopicName("Disabled Rising Topic"))
	seedUserTopic(t, env.db, user.ID, topic.ID, withNotificationEnabled(true))
	seedNotificationSetting(t, env.db, user.ID, withRisingEnabled(false))

	// Insert unnotified rising history (status=2)
	rising := seedSpikeHistory(t, env.db, topic.ID,
		withSpikeStatus(2),
		withSpikePeakZScore(2.8),
	)

	spikeHistRepo := repository.NewSpikeHistoryRepository(env.db)
	userTopicRepo := repository.NewUserTopicRepository(env.db)
	notiSettingRepo := repository.NewNotificationSettingRepository(env.db)
	notiRepo := repository.NewNotificationRepository(env.db)
	uc := usecase.NewRisingNotificationUsecase(
		spikeHistRepo, userTopicRepo, notiSettingRepo, notiRepo,
	)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("batch execute failed: %v", err)
	}

	// notified_at should be set (to prevent re-processing)
	var sh model.SpikeHistory
	if err := env.db.First(&sh, "id = ?", rising.ID).Error; err != nil {
		t.Fatalf("failed to find spike history: %v", err)
	}
	if sh.NotifiedAt == nil {
		t.Fatal("expected notified_at to be set even when rising disabled")
	}

	// No in-app notifications should be created
	var unCount int64
	env.db.Raw("SELECT COUNT(*) FROM user_notifications WHERE user_id = ?", user.ID).Scan(&unCount)
	if unCount != 0 {
		t.Errorf("expected 0 in-app notifications, got %d", unCount)
	}
}

func TestRisingNotificationBatch_IgnoresSpikeRecords(t *testing.T) {
	env := setupTest(t)

	// Seed user with rising_enabled=true
	user := seedUser(t, env.db, withEmail("spike-only@example.com"))
	topic := seedTopic(t, env.db, withTopicName("Spike Only Topic"))
	seedUserTopic(t, env.db, user.ID, topic.ID, withNotificationEnabled(true))
	seedNotificationSetting(t, env.db, user.ID, withRisingEnabled(true))

	// Insert unnotified spike history with status=1 (Spike, NOT Rising)
	seedSpikeHistory(t, env.db, topic.ID,
		withSpikeStatus(1),
		withSpikePeakZScore(5.0),
	)

	spikeHistRepo := repository.NewSpikeHistoryRepository(env.db)
	userTopicRepo := repository.NewUserTopicRepository(env.db)
	notiSettingRepo := repository.NewNotificationSettingRepository(env.db)
	notiRepo := repository.NewNotificationRepository(env.db)
	uc := usecase.NewRisingNotificationUsecase(
		spikeHistRepo, userTopicRepo, notiSettingRepo, notiRepo,
	)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("batch execute failed: %v", err)
	}

	// No in-app notifications should be created
	var unCount int64
	env.db.Raw("SELECT COUNT(*) FROM user_notifications WHERE user_id = ?", user.ID).Scan(&unCount)
	if unCount != 0 {
		t.Errorf("expected 0 in-app notifications, got %d", unCount)
	}
}
