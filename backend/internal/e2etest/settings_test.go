package e2etest

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/proto"

	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/gen/trendbird/v1/trendbirdv1connect"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

func TestSettingsService_GetProfile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db,
			withEmail("test@example.com"),
			withImage("https://pbs.twimg.com/profile.jpg"),
		)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewSettingsServiceClient)
		resp, err := client.GetProfile(context.Background(), connect.NewRequest(&trendbirdv1.GetProfileRequest{}))
		if err != nil {
			t.Fatalf("GetProfile: %v", err)
		}

		got := resp.Msg.GetUser()
		if got.GetName() != user.Name {
			t.Errorf("name: want %s, got %s", user.Name, got.GetName())
		}
		if got.GetEmail() != user.Email {
			t.Errorf("email: want %s, got %s", user.Email, got.GetEmail())
		}
		if got.GetTwitterHandle() != user.TwitterHandle {
			t.Errorf("twitter_handle: want %s, got %s", user.TwitterHandle, got.GetTwitterHandle())
		}
		if got.GetImage() != user.Image {
			t.Errorf("image: want %s, got %s", user.Image, got.GetImage())
		}
	})

	t.Run("cross_user_isolation", func(t *testing.T) {
		env := setupTest(t)
		seedUser(t, env.db, withName("User A"), withEmail("a@example.com"))
		userB := seedUser(t, env.db, withName("User B"), withEmail("b@example.com"))

		client := connectClient(t, env, userB.ID, trendbirdv1connect.NewSettingsServiceClient)
		resp, err := client.GetProfile(context.Background(), connect.NewRequest(&trendbirdv1.GetProfileRequest{}))
		if err != nil {
			t.Fatalf("GetProfile: %v", err)
		}

		got := resp.Msg.GetUser()
		if got.GetName() != "User B" {
			t.Errorf("name: want %q, got %q", "User B", got.GetName())
		}
		if got.GetEmail() != "b@example.com" {
			t.Errorf("email: want %q, got %q", "b@example.com", got.GetEmail())
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.settingsClient.GetProfile(context.Background(), connect.NewRequest(&trendbirdv1.GetProfileRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

func TestSettingsService_UpdateProfile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db, withEmail("old@example.com"))

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewSettingsServiceClient)
		resp, err := client.UpdateProfile(context.Background(), connect.NewRequest(&trendbirdv1.UpdateProfileRequest{
			Email: proto.String("new@example.com"),
		}))
		if err != nil {
			t.Fatalf("UpdateProfile: %v", err)
		}

		// レスポンスの検証
		if got := resp.Msg.GetUser().GetEmail(); got != "new@example.com" {
			t.Errorf("response email: want new@example.com, got %s", got)
		}

		// DB の検証
		var updated model.User
		if err := env.db.First(&updated, "id = ?", user.ID).Error; err != nil {
			t.Fatalf("failed to fetch user: %v", err)
		}
		if updated.Email != "new@example.com" {
			t.Errorf("db email: want new@example.com, got %s", updated.Email)
		}
	})

	t.Run("empty_email", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewSettingsServiceClient)
		_, err := client.UpdateProfile(context.Background(), connect.NewRequest(&trendbirdv1.UpdateProfileRequest{
			Email: proto.String(""),
		}))
		assertConnectCode(t, err, connect.CodeInvalidArgument)
	})

	t.Run("email_duplicate_allowed", func(t *testing.T) {
		env := setupTest(t)
		sharedEmail := "shared@example.com"
		seedUser(t, env.db, withEmail(sharedEmail)) // userA already has this email
		userB := seedUser(t, env.db, withEmail("different@example.com"))

		client := connectClient(t, env, userB.ID, trendbirdv1connect.NewSettingsServiceClient)
		resp, err := client.UpdateProfile(context.Background(), connect.NewRequest(&trendbirdv1.UpdateProfileRequest{
			Email: proto.String(sharedEmail),
		}))
		if err != nil {
			t.Fatalf("UpdateProfile with duplicate email: %v", err)
		}

		if got := resp.Msg.GetUser().GetEmail(); got != sharedEmail {
			t.Errorf("email: want %q, got %q", sharedEmail, got)
		}
	})

	t.Run("partial_update_email_only", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db, withName("Original Name"), withEmail("old@example.com"))

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewSettingsServiceClient)
		resp, err := client.UpdateProfile(context.Background(), connect.NewRequest(&trendbirdv1.UpdateProfileRequest{
			Email: proto.String("new@example.com"),
		}))
		if err != nil {
			t.Fatalf("UpdateProfile: %v", err)
		}

		// email が更新されている
		if got := resp.Msg.GetUser().GetEmail(); got != "new@example.com" {
			t.Errorf("email: want %q, got %q", "new@example.com", got)
		}

		// name は変更されていない
		if got := resp.Msg.GetUser().GetName(); got != "Original Name" {
			t.Errorf("name: want %q, got %q", "Original Name", got)
		}

		// DB でも検証
		var updated model.User
		if err := env.db.First(&updated, "id = ?", user.ID).Error; err != nil {
			t.Fatalf("fetch user: %v", err)
		}
		if updated.Name != "Original Name" {
			t.Errorf("db name: want %q, got %q", "Original Name", updated.Name)
		}
	})

	t.Run("nil_email_noop", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db, withEmail("keep@example.com"))

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewSettingsServiceClient)
		// email フィールドなしで呼び出し
		resp, err := client.UpdateProfile(context.Background(), connect.NewRequest(&trendbirdv1.UpdateProfileRequest{}))
		if err != nil {
			t.Fatalf("UpdateProfile with nil email: %v", err)
		}

		// email は変更されていない
		if got := resp.Msg.GetUser().GetEmail(); got != "keep@example.com" {
			t.Errorf("email: want %q, got %q", "keep@example.com", got)
		}

		// DB でも検証
		var updated model.User
		if err := env.db.First(&updated, "id = ?", user.ID).Error; err != nil {
			t.Fatalf("fetch user: %v", err)
		}
		if updated.Email != "keep@example.com" {
			t.Errorf("db email: want %q, got %q", "keep@example.com", updated.Email)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.settingsClient.UpdateProfile(context.Background(), connect.NewRequest(&trendbirdv1.UpdateProfileRequest{
			Email: proto.String("test@example.com"),
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

func TestSettingsService_GetNotificationSettings(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		seedNotificationSetting(t, env.db, user.ID, withSpikeEnabled(false))

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewSettingsServiceClient)
		resp, err := client.GetNotificationSettings(context.Background(), connect.NewRequest(&trendbirdv1.GetNotificationSettingsRequest{}))
		if err != nil {
			t.Fatalf("GetNotificationSettings: %v", err)
		}

		settings := resp.Msg.GetSettings()
		if settings.GetSpikeEnabled() != false {
			t.Errorf("spike_enabled: want false, got %v", settings.GetSpikeEnabled())
		}
		if settings.GetRisingEnabled() != true {
			t.Errorf("rising_enabled: want true, got %v", settings.GetRisingEnabled())
		}
	})

	t.Run("cross_user_isolation", func(t *testing.T) {
		env := setupTest(t)
		userA := seedUser(t, env.db)
		userB := seedUser(t, env.db)
		seedNotificationSetting(t, env.db, userA.ID, withSpikeEnabled(false))
		seedNotificationSetting(t, env.db, userB.ID, withSpikeEnabled(true))

		client := connectClient(t, env, userB.ID, trendbirdv1connect.NewSettingsServiceClient)
		resp, err := client.GetNotificationSettings(context.Background(), connect.NewRequest(&trendbirdv1.GetNotificationSettingsRequest{}))
		if err != nil {
			t.Fatalf("GetNotificationSettings: %v", err)
		}

		settings := resp.Msg.GetSettings()
		if !settings.GetSpikeEnabled() {
			t.Error("spike_enabled: want true (userB's setting), got false")
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.settingsClient.GetNotificationSettings(context.Background(), connect.NewRequest(&trendbirdv1.GetNotificationSettingsRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

func TestSettingsService_UpdateNotifications(t *testing.T) {
	t.Run("success_partial_update", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		ns := seedNotificationSetting(t, env.db, user.ID) // デフォルト: 全 true

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewSettingsServiceClient)
		resp, err := client.UpdateNotifications(context.Background(), connect.NewRequest(&trendbirdv1.UpdateNotificationsRequest{
			SpikeEnabled: proto.Bool(false),
		}))
		if err != nil {
			t.Fatalf("UpdateNotifications: %v", err)
		}
		if !resp.Msg.GetUpdated() {
			t.Error("expected updated=true")
		}

		// DB で部分更新を検証
		var updated model.NotificationSetting
		if err := env.db.First(&updated, "id = ?", ns.ID).Error; err != nil {
			t.Fatalf("failed to fetch notification_setting: %v", err)
		}
		if updated.SpikeEnabled != false {
			t.Errorf("spike_enabled: want false, got %v", updated.SpikeEnabled)
		}
		if updated.RisingEnabled != true {
			t.Errorf("rising_enabled: want true, got %v", updated.RisingEnabled)
		}
	})

	t.Run("success_update_all_fields", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		ns := seedNotificationSetting(t, env.db, user.ID) // デフォルト: 全 true

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewSettingsServiceClient)
		_, err := client.UpdateNotifications(context.Background(), connect.NewRequest(&trendbirdv1.UpdateNotificationsRequest{
			SpikeEnabled:  proto.Bool(false),
			RisingEnabled: proto.Bool(false),
		}))
		if err != nil {
			t.Fatalf("UpdateNotifications: %v", err)
		}

		// DB で全フィールド更新を検証
		var updated model.NotificationSetting
		if err := env.db.First(&updated, "id = ?", ns.ID).Error; err != nil {
			t.Fatalf("failed to fetch notification_setting: %v", err)
		}
		if updated.SpikeEnabled != false {
			t.Errorf("spike_enabled: want false, got %v", updated.SpikeEnabled)
		}
		if updated.RisingEnabled != false {
			t.Errorf("rising_enabled: want false, got %v", updated.RisingEnabled)
		}
	})

	t.Run("full_toggle_cycle", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		seedNotificationSetting(t, env.db, user.ID) // デフォルト: 全 true

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewSettingsServiceClient)

		// Step 1: 全 true → 全 false
		_, err := client.UpdateNotifications(context.Background(), connect.NewRequest(&trendbirdv1.UpdateNotificationsRequest{
			SpikeEnabled:  proto.Bool(false),
			RisingEnabled: proto.Bool(false),
		}))
		if err != nil {
			t.Fatalf("UpdateNotifications (to false): %v", err)
		}

		// GetNotificationSettings で全 false 確認
		getResp1, err := client.GetNotificationSettings(context.Background(), connect.NewRequest(&trendbirdv1.GetNotificationSettingsRequest{}))
		if err != nil {
			t.Fatalf("GetNotificationSettings after false: %v", err)
		}
		s1 := getResp1.Msg.GetSettings()
		if s1.GetSpikeEnabled() {
			t.Error("spike_enabled: want false after toggle off")
		}
		if s1.GetRisingEnabled() {
			t.Error("rising_enabled: want false after toggle off")
		}

		// Step 2: 全 false → 全 true
		_, err = client.UpdateNotifications(context.Background(), connect.NewRequest(&trendbirdv1.UpdateNotificationsRequest{
			SpikeEnabled:  proto.Bool(true),
			RisingEnabled: proto.Bool(true),
		}))
		if err != nil {
			t.Fatalf("UpdateNotifications (to true): %v", err)
		}

		// GetNotificationSettings で全 true 確認
		getResp2, err := client.GetNotificationSettings(context.Background(), connect.NewRequest(&trendbirdv1.GetNotificationSettingsRequest{}))
		if err != nil {
			t.Fatalf("GetNotificationSettings after true: %v", err)
		}
		s2 := getResp2.Msg.GetSettings()
		if !s2.GetSpikeEnabled() {
			t.Error("spike_enabled: want true after toggle on")
		}
		if !s2.GetRisingEnabled() {
			t.Error("rising_enabled: want true after toggle on")
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.settingsClient.UpdateNotifications(context.Background(), connect.NewRequest(&trendbirdv1.UpdateNotificationsRequest{
			SpikeEnabled: proto.Bool(true),
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}
