package e2etest

import (
	"context"
	"testing"
	"time"

	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
	"github.com/trendbird/backend/internal/infrastructure/persistence/repository"
	"github.com/trendbird/backend/internal/usecase"
)

func TestSpikeNotificationBatch_E2E(t *testing.T) {
	env := setupTest(t)

	// Seed data: user → topic → user_topic → notification_setting → spike_history
	user := seedUser(t, env.db, withEmail("batch-test@example.com"))
	topic := seedTopic(t, env.db, withTopicName("Go 1.26"))
	seedUserTopic(t, env.db, user.ID, topic.ID, withNotificationEnabled(true))
	seedNotificationSetting(t, env.db, user.ID, withSpikeEnabled(true))

	// Insert unnotified spike history
	spike := seedSpikeHistory(t, env.db, topic.ID, withSpikePeakZScore(4.5), withSpikeSummary("Go 1.26 release spike"))

	// Verify notified_at is NULL
	var sh model.SpikeHistory
	if err := env.db.First(&sh, "id = ?", spike.ID).Error; err != nil {
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
	uc := usecase.NewSpikeNotificationUsecase(
		spikeHistRepo, userTopicRepo, notiSettingRepo, notiRepo,
	)

	ctx := context.Background()
	if err := uc.Execute(ctx); err != nil {
		t.Fatalf("batch execute failed: %v", err)
	}

	// Verify notified_at is now set
	if err := env.db.First(&sh, "id = ?", spike.ID).Error; err != nil {
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

func TestSpikeNotificationBatch_NoCooldownSuppression(t *testing.T) {
	// OSS版: クールダウン=0 のため、直前の通知があっても抑制されない
	env := setupTest(t)

	user := seedUser(t, env.db, withEmail("cooldown-spike@example.com"))
	topic := seedTopic(t, env.db, withTopicName("Cooldown Topic"))
	seedUserTopic(t, env.db, user.ID, topic.ID, withNotificationEnabled(true))
	seedNotificationSetting(t, env.db, user.ID, withSpikeEnabled(true))

	// 30分前の通知を作成（クールダウン=0 なので抑制されない）
	noti := seedNotification(t, env.db, user.ID, withNotificationType(1)) // type=1 Trend
	env.db.Exec("UPDATE notifications SET created_at = ? WHERE id = ?", time.Now().Add(-30*time.Minute), noti.ID)

	// 未通知の spike を作成
	spike := seedSpikeHistory(t, env.db, topic.ID, withSpikePeakZScore(4.5), withSpikeSummary("Cooldown test spike"))

	spikeHistRepo := repository.NewSpikeHistoryRepository(env.db)
	userTopicRepo := repository.NewUserTopicRepository(env.db)
	notiSettingRepo := repository.NewNotificationSettingRepository(env.db)
	notiRepo := repository.NewNotificationRepository(env.db)
	uc := usecase.NewSpikeNotificationUsecase(
		spikeHistRepo, userTopicRepo, notiSettingRepo, notiRepo,
	)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("batch execute failed: %v", err)
	}

	// spike は処理済み（notified_at が設定される）
	var sh model.SpikeHistory
	if err := env.db.First(&sh, "id = ?", spike.ID).Error; err != nil {
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

func TestSpikeNotificationBatch_CooldownExpired(t *testing.T) {
	env := setupTest(t)

	// cooldown=0 (OSS版): 古い通知があっても送信される
	user := seedUser(t, env.db, withEmail("expired-cooldown@example.com"))
	topic := seedTopic(t, env.db, withTopicName("Expired Cooldown Topic"))
	seedUserTopic(t, env.db, user.ID, topic.ID, withNotificationEnabled(true))
	seedNotificationSetting(t, env.db, user.ID, withSpikeEnabled(true))

	// 1441分前の通知（cooldown 超過）
	noti := seedNotification(t, env.db, user.ID, withNotificationType(1))
	env.db.Exec("UPDATE notifications SET created_at = ? WHERE id = ?", time.Now().Add(-1441*time.Minute), noti.ID)

	// 未通知の spike を作成
	spike := seedSpikeHistory(t, env.db, topic.ID, withSpikePeakZScore(4.5), withSpikeSummary("Expired cooldown spike"))

	spikeHistRepo := repository.NewSpikeHistoryRepository(env.db)
	userTopicRepo := repository.NewUserTopicRepository(env.db)
	notiSettingRepo := repository.NewNotificationSettingRepository(env.db)
	notiRepo := repository.NewNotificationRepository(env.db)
	uc := usecase.NewSpikeNotificationUsecase(
		spikeHistRepo, userTopicRepo, notiSettingRepo, notiRepo,
	)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("batch execute failed: %v", err)
	}

	// spike は処理済み
	var sh model.SpikeHistory
	if err := env.db.First(&sh, "id = ?", spike.ID).Error; err != nil {
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

func TestSpikeNotificationBatch_MultipleSpikesPerTopic(t *testing.T) {
	env := setupTest(t)

	user := seedUser(t, env.db, withEmail("multi-spike@example.com"))
	topic := seedTopic(t, env.db, withTopicName("Multi Spike Topic"))
	seedUserTopic(t, env.db, user.ID, topic.ID, withNotificationEnabled(true))
	seedNotificationSetting(t, env.db, user.ID, withSpikeEnabled(true))

	// 同一トピックに2件の未通知 spike_history
	spike1 := seedSpikeHistory(t, env.db, topic.ID, withSpikePeakZScore(4.0), withSpikeSummary("Spike 1"))
	spike2 := seedSpikeHistory(t, env.db, topic.ID, withSpikePeakZScore(5.0), withSpikeSummary("Spike 2"))

	spikeHistRepo := repository.NewSpikeHistoryRepository(env.db)
	userTopicRepo := repository.NewUserTopicRepository(env.db)
	notiSettingRepo := repository.NewNotificationSettingRepository(env.db)
	notiRepo := repository.NewNotificationRepository(env.db)
	uc := usecase.NewSpikeNotificationUsecase(
		spikeHistRepo, userTopicRepo, notiSettingRepo, notiRepo,
	)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("batch execute failed: %v", err)
	}

	// 両方の spike が処理済み（notified_at が設定されている）
	var sh1, sh2 model.SpikeHistory
	if err := env.db.First(&sh1, "id = ?", spike1.ID).Error; err != nil {
		t.Fatalf("find spike1: %v", err)
	}
	if sh1.NotifiedAt == nil {
		t.Error("spike1: notified_at should be set")
	}

	if err := env.db.First(&sh2, "id = ?", spike2.ID).Error; err != nil {
		t.Fatalf("find spike2: %v", err)
	}
	if sh2.NotifiedAt == nil {
		t.Error("spike2: notified_at should be set")
	}

	// 各 spike ごとに通知が作成される（クールダウンはバッチ実行前に1回チェック）
	var unCount int64
	env.db.Raw("SELECT COUNT(*) FROM user_notifications WHERE user_id = ?", user.ID).Scan(&unCount)
	if unCount != 2 {
		t.Errorf("user_notifications: want 2 (one per spike), got %d", unCount)
	}

	// 冪等性: 2回目の実行では追加通知なし
	beforeCount := unCount

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("second batch execute failed: %v", err)
	}

	var afterCount int64
	env.db.Raw("SELECT COUNT(*) FROM user_notifications WHERE user_id = ?", user.ID).Scan(&afterCount)
	if afterCount != beforeCount {
		t.Errorf("expected no additional notifications on re-run, but got %d more", afterCount-beforeCount)
	}
}

func TestSpikeNotificationBatch_NoSubscribers(t *testing.T) {
	env := setupTest(t)

	// Seed topic with spike but no subscribers
	topic := seedTopic(t, env.db, withTopicName("Orphan Topic"))
	spike := seedSpikeHistory(t, env.db, topic.ID)

	spikeHistRepo := repository.NewSpikeHistoryRepository(env.db)
	userTopicRepo := repository.NewUserTopicRepository(env.db)
	notiSettingRepo := repository.NewNotificationSettingRepository(env.db)
	notiRepo := repository.NewNotificationRepository(env.db)
	uc := usecase.NewSpikeNotificationUsecase(
		spikeHistRepo, userTopicRepo, notiSettingRepo, notiRepo,
	)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("batch execute failed: %v", err)
	}

	// Should still be marked as notified even with no subscribers
	var sh model.SpikeHistory
	if err := env.db.First(&sh, "id = ?", spike.ID).Error; err != nil {
		t.Fatalf("failed to find spike history: %v", err)
	}
	if sh.NotifiedAt == nil {
		t.Fatal("expected notified_at to be set even with no subscribers")
	}
}

func TestSpikeNotificationBatch_NoCooldown(t *testing.T) {
	// OSS版: クールダウン=0 のため、直前に通知があっても再通知される
	env := setupTest(t)

	user := seedUser(t, env.db, withEmail("nocooldown@example.com"))
	topic := seedTopic(t, env.db, withTopicName("Cooldown Topic"))
	seedUserTopic(t, env.db, user.ID, topic.ID, withNotificationEnabled(true))
	seedNotificationSetting(t, env.db, user.ID, withSpikeEnabled(true))

	// 1分前の通知を作成（クールダウン=0 なので通知される）
	noti := seedNotification(t, env.db, user.ID, withNotificationType(1))
	env.db.Exec("UPDATE notifications SET created_at = ? WHERE id = ?", time.Now().Add(-1*time.Minute), noti.ID)

	// 未通知 spike を作成
	seedSpikeHistory(t, env.db, topic.ID, withSpikePeakZScore(4.5), withSpikeSummary("No cooldown spike"))

	spikeHistRepo := repository.NewSpikeHistoryRepository(env.db)
	userTopicRepo := repository.NewUserTopicRepository(env.db)
	notiSettingRepo := repository.NewNotificationSettingRepository(env.db)
	notiRepo := repository.NewNotificationRepository(env.db)
	uc := usecase.NewSpikeNotificationUsecase(
		spikeHistRepo, userTopicRepo, notiSettingRepo, notiRepo,
	)

	if err := uc.Execute(context.Background()); err != nil {
		t.Fatalf("batch execute failed: %v", err)
	}

	// クールダウン=0: 直前の通知があっても新規通知が作成される
	var unCount int64
	env.db.Raw("SELECT COUNT(*) FROM user_notifications WHERE user_id = ?", user.ID).Scan(&unCount)
	if unCount != 2 { // seed 1件 + バッチ 1件
		t.Errorf("user_notifications: want 2, got %d", unCount)
	}
}
