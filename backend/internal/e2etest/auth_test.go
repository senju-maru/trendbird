package e2etest

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/golang-jwt/jwt/v5"

	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/gen/trendbird/v1/trendbirdv1connect"
	"github.com/trendbird/backend/internal/domain/gateway"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

func TestAuthService_XAuth(t *testing.T) {
	t.Run("new_user_registration", func(t *testing.T) {
		env := setupTest(t)

		// モック設定: ExchangeCode → トークン返却（デフォルトのまま）
		// モック設定: GetUserInfo → 新規ユーザー情報を返却
		env.mockTwitter.GetUserInfoFn = func(_ context.Context, _ string) (*gateway.TwitterUserInfo, error) {
			return &gateway.TwitterUserInfo{
				ID:       "x-new-user",
				Name:     "New User",
				Username: "newuser",
				Email:    "new@example.com",
				Image:    "https://img.example.com/new.jpg",
			}, nil
		}

		// cookieInterceptor 付き AuthServiceClient を生成
		client := trendbirdv1connect.NewAuthServiceClient(
			env.server.Client(),
			env.server.URL,
			connect.WithInterceptors(cookieInterceptor("tb_cv=test-code-verifier")),
		)

		resp, err := client.XAuth(context.Background(), connect.NewRequest(&trendbirdv1.XAuthRequest{
			OauthCode: "test-auth-code",
		}))
		if err != nil {
			t.Fatalf("XAuth: %v", err)
		}

		// --- レスポンス検証 ---
		if !resp.Msg.GetTutorialPending() {
			t.Error("tutorial_pending: want true, got false")
		}

		user := resp.Msg.GetUser()
		if user.GetName() != "New User" {
			t.Errorf("name: want %q, got %q", "New User", user.GetName())
		}
		if user.GetTwitterHandle() != "newuser" {
			t.Errorf("twitter_handle: want %q, got %q", "newuser", user.GetTwitterHandle())
		}
		// --- DB 検証: users テーブル ---
		var dbUser model.User
		if err := env.db.First(&dbUser, "twitter_id = ?", "x-new-user").Error; err != nil {
			t.Fatalf("user not found in DB: %v", err)
		}
		if dbUser.Name != "New User" {
			t.Errorf("db user name: want %q, got %q", "New User", dbUser.Name)
		}

		// --- DB 検証: twitter_connections テーブル ---
		var dbConn model.TwitterConnection
		if err := env.db.First(&dbConn, "user_id = ?", dbUser.ID).Error; err != nil {
			t.Fatalf("twitter_connection not found in DB: %v", err)
		}
		if dbConn.Status != 3 { // Connected
			t.Errorf("twitter_connection status: want 3 (Connected), got %d", dbConn.Status)
		}

		// --- DB 検証: notification_settings テーブル（デフォルト値で作成） ---
		var dbNotiSetting model.NotificationSetting
		if err := env.db.First(&dbNotiSetting, "user_id = ?", dbUser.ID).Error; err != nil {
			t.Fatalf("notification_setting not found in DB: %v", err)
		}
		if !dbNotiSetting.SpikeEnabled {
			t.Error("notification_setting spike_enabled: want true, got false")
		}
		if !dbNotiSetting.RisingEnabled {
			t.Error("notification_setting rising_enabled: want true, got false")
		}
		// --- Set-Cookie 検証 ---
		header := resp.Header()

		// tb_jwt が存在し値が非空
		jwtCookie, ok := findSetCookie(header, "tb_jwt")
		if !ok {
			t.Error("Set-Cookie tb_jwt not found in response")
		} else if jwtCookie == "tb_jwt=" {
			t.Error("Set-Cookie tb_jwt value is empty")
		}

		// tb_cv が Max-Age=0 でクリア
		assertSetCookieCleared(t, header, "tb_cv")

		// tb_state が Max-Age=0 でクリア
		assertSetCookieCleared(t, header, "tb_state")
	})

	t.Run("existing_user_relogin", func(t *testing.T) {
		env := setupTest(t)

		// 既存ユーザーを seed（tutorial_completed=true がデフォルト）
		user := seedUser(t, env.db,
			withTwitterID("x-existing"),
			withName("Old Name"),
			withEmail("old@example.com"),
			withTwitterHandle("existinguser"),
		)

		seedTwitterConnection(t, env.db, user.ID)

		// モック設定: GetUserInfo → 更新されたユーザー情報を返却
		env.mockTwitter.GetUserInfoFn = func(_ context.Context, _ string) (*gateway.TwitterUserInfo, error) {
			return &gateway.TwitterUserInfo{
				ID:       "x-existing",
				Name:     "Updated Name",
				Username: "existinguser",
				Email:    "updated@example.com",
				Image:    "https://img.example.com/updated.jpg",
			}, nil
		}

		client := trendbirdv1connect.NewAuthServiceClient(
			env.server.Client(),
			env.server.URL,
			connect.WithInterceptors(cookieInterceptor("tb_cv=test-code-verifier")),
		)

		resp, err := client.XAuth(context.Background(), connect.NewRequest(&trendbirdv1.XAuthRequest{
			OauthCode: "test-auth-code",
		}))
		if err != nil {
			t.Fatalf("XAuth: %v", err)
		}

		// --- レスポンス検証: tutorial_pending が false ---
		if resp.Msg.GetTutorialPending() {
			t.Error("tutorial_pending: want false, got true")
		}

		// --- DB 検証: users テーブルの name が更新されていること ---
		var dbUser model.User
		if err := env.db.First(&dbUser, "id = ?", user.ID).Error; err != nil {
			t.Fatalf("user not found in DB: %v", err)
		}
		if dbUser.Name != "Updated Name" {
			t.Errorf("db user name: want %q, got %q", "Updated Name", dbUser.Name)
		}

		// --- DB 検証: twitter_connections のアクセストークンが新しいモック値に更新 ---
		var dbConn model.TwitterConnection
		if err := env.db.First(&dbConn, "user_id = ?", user.ID).Error; err != nil {
			t.Fatalf("twitter_connection not found in DB: %v", err)
		}
		if dbConn.AccessToken != "test-access-token" {
			t.Errorf("access_token: want %q, got %q", "test-access-token", dbConn.AccessToken)
		}

		// --- DB 検証: notification_settings は新規作成されない ---
		var notiCount int64
		env.db.Model(&model.NotificationSetting{}).Where("user_id = ?", user.ID).Count(&notiCount)
		if notiCount != 0 {
			t.Errorf("notification_settings count: want 0, got %d", notiCount)
		}

	})

	t.Run("email_not_provided", func(t *testing.T) {
		env := setupTest(t)

		// モック設定: GetUserInfo → Email="" で返却
		env.mockTwitter.GetUserInfoFn = func(_ context.Context, _ string) (*gateway.TwitterUserInfo, error) {
			return &gateway.TwitterUserInfo{
				ID:       "x-no-email-user",
				Name:     "No Email User",
				Username: "noemail",
				Email:    "",
				Image:    "https://img.example.com/noemail.jpg",
			}, nil
		}

		client := trendbirdv1connect.NewAuthServiceClient(
			env.server.Client(),
			env.server.URL,
			connect.WithInterceptors(cookieInterceptor("tb_cv=test-code-verifier")),
		)

		resp, err := client.XAuth(context.Background(), connect.NewRequest(&trendbirdv1.XAuthRequest{
			OauthCode: "test-auth-code",
		}))
		if err != nil {
			t.Fatalf("XAuth: %v", err)
		}

		// 新規ユーザーとして登録成功
		if !resp.Msg.GetTutorialPending() {
			t.Error("tutorial_pending: want true, got false")
		}

		// DB: users.email="" が保存されていること
		var dbUser model.User
		if err := env.db.First(&dbUser, "twitter_id = ?", "x-no-email-user").Error; err != nil {
			t.Fatalf("user not found in DB: %v", err)
		}
		if dbUser.Email != "" {
			t.Errorf("db user email: want empty, got %q", dbUser.Email)
		}

		// notification_settings が作成されていること
		var notiCount int64
		env.db.Model(&model.NotificationSetting{}).Where("user_id = ?", dbUser.ID).Count(&notiCount)
		if notiCount != 1 {
			t.Errorf("notification_settings count: want 1, got %d", notiCount)
		}

	})

	t.Run("existing_user_image_update", func(t *testing.T) {
		env := setupTest(t)

		// 既存ユーザーを seed（tutorial_completed=true がデフォルト）
		user := seedUser(t, env.db,
			withTwitterID("x-image-update"),
			withName("Old Name"),
			withEmail("stable@example.com"),
			withImage("old.jpg"),
			withTwitterHandle("imageuser"),
		)
		seedTwitterConnection(t, env.db, user.ID)

		// モック設定: GetUserInfo → image と name が更新、email は変更
		env.mockTwitter.GetUserInfoFn = func(_ context.Context, _ string) (*gateway.TwitterUserInfo, error) {
			return &gateway.TwitterUserInfo{
				ID:       "x-image-update",
				Name:     "New Name",
				Username: "imageuser",
				Email:    "changed@example.com",
				Image:    "new.jpg",
			}, nil
		}

		client := trendbirdv1connect.NewAuthServiceClient(
			env.server.Client(),
			env.server.URL,
			connect.WithInterceptors(cookieInterceptor("tb_cv=test-code-verifier")),
		)

		resp, err := client.XAuth(context.Background(), connect.NewRequest(&trendbirdv1.XAuthRequest{
			OauthCode: "test-auth-code",
		}))
		if err != nil {
			t.Fatalf("XAuth: %v", err)
		}

		if resp.Msg.GetTutorialPending() {
			t.Error("tutorial_pending: want false, got true")
		}

		// DB 検証
		var dbUser model.User
		if err := env.db.First(&dbUser, "id = ?", user.ID).Error; err != nil {
			t.Fatalf("user not found in DB: %v", err)
		}

		// image は更新される（DoUpdates に含まれる）
		if dbUser.Image != "new.jpg" {
			t.Errorf("db user image: want %q, got %q", "new.jpg", dbUser.Image)
		}
		// name は更新される（DoUpdates に含まれる）
		if dbUser.Name != "New Name" {
			t.Errorf("db user name: want %q, got %q", "New Name", dbUser.Name)
		}
		// email は変更されない（DoUpdates に email 未含有）
		if dbUser.Email != "stable@example.com" {
			t.Errorf("db user email: want %q (unchanged), got %q", "stable@example.com", dbUser.Email)
		}
	})

	t.Run("exchange_code_failure", func(t *testing.T) {
		env := setupTest(t)

		// モック設定: ExchangeCode → エラー返却
		env.mockTwitter.ExchangeCodeFn = func(_ context.Context, _, _ string) (*gateway.OAuthTokenResponse, error) {
			return nil, fmt.Errorf("oauth error")
		}

		client := trendbirdv1connect.NewAuthServiceClient(
			env.server.Client(),
			env.server.URL,
			connect.WithInterceptors(cookieInterceptor("tb_cv=test-code-verifier")),
		)

		_, err := client.XAuth(context.Background(), connect.NewRequest(&trendbirdv1.XAuthRequest{
			OauthCode: "test-auth-code",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})

	t.Run("get_user_info_failure", func(t *testing.T) {
		env := setupTest(t)

		// モック設定: ExchangeCode → 成功（デフォルト）
		// モック設定: GetUserInfo → エラー返却
		env.mockTwitter.GetUserInfoFn = func(_ context.Context, _ string) (*gateway.TwitterUserInfo, error) {
			return nil, fmt.Errorf("api error")
		}

		client := trendbirdv1connect.NewAuthServiceClient(
			env.server.Client(),
			env.server.URL,
			connect.WithInterceptors(cookieInterceptor("tb_cv=test-code-verifier")),
		)

		_, err := client.XAuth(context.Background(), connect.NewRequest(&trendbirdv1.XAuthRequest{
			OauthCode: "test-auth-code",
		}))
		assertConnectCode(t, err, connect.CodeInternal)
	})
}

func TestAuthService_GetCurrentUser(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db, withEmail("me@example.com"), withName("Test User"), withTwitterHandle("testhandle"))

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAuthServiceClient)
		resp, err := client.GetCurrentUser(context.Background(), connect.NewRequest(&trendbirdv1.GetCurrentUserRequest{}))
		if err != nil {
			t.Fatalf("GetCurrentUser: %v", err)
		}

		got := resp.Msg.GetUser()
		if got.GetName() != "Test User" {
			t.Errorf("name: want %q, got %q", "Test User", got.GetName())
		}
		if got.GetEmail() != "me@example.com" {
			t.Errorf("email: want %q, got %q", "me@example.com", got.GetEmail())
		}
		if got.GetTwitterHandle() != "testhandle" {
			t.Errorf("twitter_handle: want %q, got %q", "testhandle", got.GetTwitterHandle())
		}
	})

	t.Run("cross_user_isolation", func(t *testing.T) {
		env := setupTest(t)
		userA := seedUser(t, env.db, withName("User A"), withEmail("a@example.com"))
		userB := seedUser(t, env.db, withName("User B"), withEmail("b@example.com"))

		// userA のセッションで GetCurrentUser → userA のデータのみ返る
		clientA := connectClient(t, env, userA.ID, trendbirdv1connect.NewAuthServiceClient)
		respA, err := clientA.GetCurrentUser(context.Background(), connect.NewRequest(&trendbirdv1.GetCurrentUserRequest{}))
		if err != nil {
			t.Fatalf("GetCurrentUser userA: %v", err)
		}

		if respA.Msg.GetUser().GetName() != "User A" {
			t.Errorf("name: want %q, got %q", "User A", respA.Msg.GetUser().GetName())
		}

		// userB のセッションで GetCurrentUser → userB のデータのみ返る
		clientB := connectClient(t, env, userB.ID, trendbirdv1connect.NewAuthServiceClient)
		respB, err := clientB.GetCurrentUser(context.Background(), connect.NewRequest(&trendbirdv1.GetCurrentUserRequest{}))
		if err != nil {
			t.Fatalf("GetCurrentUser userB: %v", err)
		}

		if respB.Msg.GetUser().GetName() != "User B" {
			t.Errorf("name: want %q, got %q", "User B", respB.Msg.GetUser().GetName())
		}
	})

	t.Run("tutorial_pending_true", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db, withTutorialCompleted(false))

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAuthServiceClient)
		resp, err := client.GetCurrentUser(context.Background(), connect.NewRequest(&trendbirdv1.GetCurrentUserRequest{}))
		if err != nil {
			t.Fatalf("GetCurrentUser: %v", err)
		}

		if !resp.Msg.GetTutorialPending() {
			t.Error("tutorial_pending: want true, got false")
		}
	})

	t.Run("tutorial_pending_false", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db) // tutorial_completed=true (default)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAuthServiceClient)
		resp, err := client.GetCurrentUser(context.Background(), connect.NewRequest(&trendbirdv1.GetCurrentUserRequest{}))
		if err != nil {
			t.Fatalf("GetCurrentUser: %v", err)
		}

		if resp.Msg.GetTutorialPending() {
			t.Error("tutorial_pending: want false, got true")
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.authClient.GetCurrentUser(context.Background(), connect.NewRequest(&trendbirdv1.GetCurrentUserRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

func TestAuthService_CompleteTutorial(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db, withTutorialCompleted(false))

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAuthServiceClient)
		_, err := client.CompleteTutorial(context.Background(), connect.NewRequest(&trendbirdv1.CompleteTutorialRequest{}))
		if err != nil {
			t.Fatalf("CompleteTutorial: %v", err)
		}

		// DB 検証: tutorial_completed = true
		var dbUser model.User
		if err := env.db.First(&dbUser, "id = ?", user.ID).Error; err != nil {
			t.Fatalf("user not found in DB: %v", err)
		}
		if !dbUser.TutorialCompleted {
			t.Error("tutorial_completed: want true, got false")
		}
	})

	t.Run("idempotent", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db) // tutorial_completed=true (already done)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAuthServiceClient)
		_, err := client.CompleteTutorial(context.Background(), connect.NewRequest(&trendbirdv1.CompleteTutorialRequest{}))
		if err != nil {
			t.Fatalf("CompleteTutorial (idempotent): %v", err)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.authClient.CompleteTutorial(context.Background(), connect.NewRequest(&trendbirdv1.CompleteTutorialRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

func TestAuthService_Logout(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		seedTwitterConnection(t, env.db, user.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAuthServiceClient)
		resp, err := client.Logout(context.Background(), connect.NewRequest(&trendbirdv1.LogoutRequest{}))
		if err != nil {
			t.Fatalf("Logout: %v", err)
		}

		// --- DB 検証: twitter_connections が削除されていること ---
		var connCount int64
		env.db.Model(&model.TwitterConnection{}).Where("user_id = ?", user.ID).Count(&connCount)
		if connCount != 0 {
			t.Errorf("twitter_connections count: want 0, got %d", connCount)
		}

		// --- Set-Cookie 検証: tb_jwt が Max-Age=0 でクリア ---
		assertSetCookieCleared(t, resp.Header(), "tb_jwt")
	})

	t.Run("token_reuse_after_logout", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		seedTwitterConnection(t, env.db, user.ID)

		token := generateTestToken(t, user.ID)

		// ログアウト実行
		logoutClient := trendbirdv1connect.NewAuthServiceClient(
			env.server.Client(),
			env.server.URL,
			connect.WithInterceptors(authTokenInterceptor(token)),
		)
		_, err := logoutClient.Logout(context.Background(), connect.NewRequest(&trendbirdv1.LogoutRequest{}))
		if err != nil {
			t.Fatalf("Logout: %v", err)
		}

		// ログアウト後に同じトークンで GetCurrentUser を呼ぶ。
		// 現在の実装ではステートレス JWT（ブラックリストなし）のため、
		// ユーザーレコードが残っている限り成功する。
		// 将来ブラックリスト実装時にこのアサーションを変更する。
		reuseClient := trendbirdv1connect.NewAuthServiceClient(
			env.server.Client(),
			env.server.URL,
			connect.WithInterceptors(authTokenInterceptor(token)),
		)
		_, err = reuseClient.GetCurrentUser(context.Background(), connect.NewRequest(&trendbirdv1.GetCurrentUserRequest{}))
		if err != nil {
			t.Errorf("GetCurrentUser after logout: expected success (stateless JWT), got %v", err)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.authClient.Logout(context.Background(), connect.NewRequest(&trendbirdv1.LogoutRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

func TestAuthService_DeleteAccount(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		seedTwitterConnection(t, env.db, user.ID)
		tp := seedTopic(t, env.db)
		seedUserTopic(t, env.db, user.ID, tp.ID)
		seedNotificationSetting(t, env.db, user.ID)
		seedNotification(t, env.db, user.ID)
		seedActivity(t, env.db, user.ID)
		seedAIGenerationLog(t, env.db, user.ID)
		seedPost(t, env.db, user.ID)
		// generated_posts を直接 seed（専用ヘルパーなし）
		if err := env.db.Create(&model.GeneratedPost{
			UserID:  user.ID,
			Style:   1,
			Content: "generated content",
		}).Error; err != nil {
			t.Fatalf("seed generated_post: %v", err)
		}

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAuthServiceClient)
		resp, err := client.DeleteAccount(context.Background(), connect.NewRequest(&trendbirdv1.DeleteAccountRequest{}))
		if err != nil {
			t.Fatalf("DeleteAccount: %v", err)
		}

		// --- DB 検証: users が削除されていること ---
		var userCount int64
		env.db.Model(&model.User{}).Where("id = ?", user.ID).Count(&userCount)
		if userCount != 0 {
			t.Errorf("users count: want 0, got %d", userCount)
		}

		// --- DB 検証: twitter_connections が削除されていること ---
		var connCount int64
		env.db.Model(&model.TwitterConnection{}).Where("user_id = ?", user.ID).Count(&connCount)
		if connCount != 0 {
			t.Errorf("twitter_connections count: want 0, got %d", connCount)
		}

		// --- DB 検証: CASCADE で user_topics リンクが全削除されていること ---
		var utCount int64
		env.db.Model(&model.UserTopic{}).Where("user_id = ?", user.ID).Count(&utCount)
		if utCount != 0 {
			t.Errorf("user_topics count: want 0, got %d", utCount)
		}

		var notiSettingCount int64
		env.db.Model(&model.NotificationSetting{}).Where("user_id = ?", user.ID).Count(&notiSettingCount)
		if notiSettingCount != 0 {
			t.Errorf("notification_settings count: want 0, got %d", notiSettingCount)
		}

		var notiCount int64
		env.db.Model(&model.UserNotification{}).Where("user_id = ?", user.ID).Count(&notiCount)
		if notiCount != 0 {
			t.Errorf("user_notifications count: want 0, got %d", notiCount)
		}

		var activityCount int64
		env.db.Model(&model.Activity{}).Where("user_id = ?", user.ID).Count(&activityCount)
		if activityCount != 0 {
			t.Errorf("activities count: want 0, got %d", activityCount)
		}

		var aiGenCount int64
		env.db.Model(&model.AIGenerationLog{}).Where("user_id = ?", user.ID).Count(&aiGenCount)
		if aiGenCount != 0 {
			t.Errorf("ai_generation_logs count: want 0, got %d", aiGenCount)
		}

		var postCount int64
		env.db.Model(&model.Post{}).Where("user_id = ?", user.ID).Count(&postCount)
		if postCount != 0 {
			t.Errorf("posts count: want 0, got %d", postCount)
		}

		var genPostCount int64
		env.db.Model(&model.GeneratedPost{}).Where("user_id = ?", user.ID).Count(&genPostCount)
		if genPostCount != 0 {
			t.Errorf("generated_posts count: want 0, got %d", genPostCount)
		}

		// --- Set-Cookie 検証: tb_jwt が Max-Age=0 でクリア ---
		assertSetCookieCleared(t, resp.Header(), "tb_jwt")
	})

	t.Run("cascade_auto_dm_rules", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		seedTwitterConnection(t, env.db, user.ID)

		// AutoDMRule と DMSentLog を seed
		rule := seedAutoDMRule(t, env.db, user.ID)
		seedDMSentLog(t, env.db, user.ID, rule.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAuthServiceClient)
		_, err := client.DeleteAccount(context.Background(), connect.NewRequest(&trendbirdv1.DeleteAccountRequest{}))
		if err != nil {
			t.Fatalf("DeleteAccount: %v", err)
		}

		// --- DB 検証: auto_dm_rules が CASCADE 削除されていること ---
		var ruleCount int64
		env.db.Model(&model.AutoDMRule{}).Where("user_id = ?", user.ID).Count(&ruleCount)
		if ruleCount != 0 {
			t.Errorf("auto_dm_rules count: want 0, got %d", ruleCount)
		}

		// --- DB 検証: dm_sent_logs が CASCADE 削除されていること ---
		var logCount int64
		env.db.Model(&model.DMSentLog{}).Where("user_id = ?", user.ID).Count(&logCount)
		if logCount != 0 {
			t.Errorf("dm_sent_logs count: want 0, got %d", logCount)
		}
	})

	t.Run("full_cascade_with_user_genres", func(t *testing.T) {
		env := setupTest(t)

		user := seedUser(t, env.db)
		seedTwitterConnection(t, env.db, user.ID)
		seedUserGenre(t, env.db, user.ID, "technology")
		seedUserGenre(t, env.db, user.ID, "marketing")

		// Shared topic: both user and userB subscribe
		tp := seedTopic(t, env.db, withTopicGenre("technology"))
		seedUserTopic(t, env.db, user.ID, tp.ID)

		userB := seedUser(t, env.db)
		seedUserTopic(t, env.db, userB.ID, tp.ID)
		seedUserGenre(t, env.db, userB.ID, "technology")

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewAuthServiceClient)
		_, err := client.DeleteAccount(context.Background(), connect.NewRequest(&trendbirdv1.DeleteAccountRequest{}))
		if err != nil {
			t.Fatalf("DeleteAccount: %v", err)
		}

		// user_genres CASCADE 削除
		var ugCount int64
		env.db.Model(&model.UserGenre{}).Where("user_id = ?", user.ID).Count(&ugCount)
		if ugCount != 0 {
			t.Errorf("user_genres count: want 0, got %d", ugCount)
		}

		// topic は残存（共有リソース）
		var topicCount int64
		env.db.Model(&model.Topic{}).Where("id = ?", tp.ID).Count(&topicCount)
		if topicCount != 1 {
			t.Errorf("topics count: want 1, got %d", topicCount)
		}

		// userB の user_topics は不変
		var utCount int64
		env.db.Model(&model.UserTopic{}).Where("user_id = ? AND topic_id = ?", userB.ID, tp.ID).Count(&utCount)
		if utCount != 1 {
			t.Errorf("userB user_topics: want 1, got %d", utCount)
		}

		// userB の user_genres も不変
		var ugBCount int64
		env.db.Model(&model.UserGenre{}).Where("user_id = ?", userB.ID).Count(&ugBCount)
		if ugBCount != 1 {
			t.Errorf("userB user_genres: want 1, got %d", ugBCount)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.authClient.DeleteAccount(context.Background(), connect.NewRequest(&trendbirdv1.DeleteAccountRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

// ---------------------------------------------------------------------------
// TestAuthService_XAuth_ConcurrentSameUser
// ---------------------------------------------------------------------------

func TestAuthService_XAuth_ConcurrentSameUser(t *testing.T) {
	env := setupTest(t)

	// Both goroutines will return the same twitter_id
	env.mockTwitter.GetUserInfoFn = func(_ context.Context, _ string) (*gateway.TwitterUserInfo, error) {
		return &gateway.TwitterUserInfo{
			ID:       "x-concurrent-same",
			Name:     "Concurrent User",
			Username: "concurrentuser",
			Email:    "concurrent@example.com",
			Image:    "https://img.example.com/concurrent.jpg",
		}, nil
	}

	type result struct {
		tutorialPending bool
		err       error
	}
	ch := make(chan result, 2)

	for range 2 {
		go func() {
			client := trendbirdv1connect.NewAuthServiceClient(
				env.server.Client(),
				env.server.URL,
				connect.WithInterceptors(cookieInterceptor("tb_cv=test-code-verifier")),
			)
			resp, err := client.XAuth(context.Background(), connect.NewRequest(&trendbirdv1.XAuthRequest{
				OauthCode: "test-auth-code",
			}))
			if err != nil {
				ch <- result{err: err}
				return
			}
			ch <- result{tutorialPending: resp.Msg.GetTutorialPending()}
		}()
	}

	r1 := <-ch
	r2 := <-ch

	// 少なくとも1つは成功
	if r1.err != nil && r2.err != nil {
		t.Fatalf("both goroutines failed: %v / %v", r1.err, r2.err)
	}

	// users テーブルに1行のみ
	var userCount int64
	env.db.Model(&model.User{}).Where("twitter_id = ?", "x-concurrent-same").Count(&userCount)
	if userCount != 1 {
		t.Errorf("users count: want 1, got %d", userCount)
	}

	// twitter_connections テーブルに1行のみ
	var tcCount int64
	env.db.Raw("SELECT COUNT(*) FROM twitter_connections tc JOIN users u ON tc.user_id = u.id WHERE u.twitter_id = ?", "x-concurrent-same").Scan(&tcCount)
	if tcCount != 1 {
		t.Errorf("twitter_connections count: want 1, got %d", tcCount)
	}

	// 少なくとも1つは tutorialPending=true
	if !r1.tutorialPending && !r2.tutorialPending && r1.err == nil && r2.err == nil {
		t.Error("expected at least one tutorialPending=true")
	}
}

// ---------------------------------------------------------------------------
// TestAuthService_DeleteAccount_Concurrent
// ---------------------------------------------------------------------------

func TestAuthService_DeleteAccount_Concurrent(t *testing.T) {
	env := setupTest(t)

	user := seedUser(t, env.db)
	seedTwitterConnection(t, env.db, user.ID)

	type result struct {
		err error
	}
	ch := make(chan result, 2)

	for range 2 {
		go func() {
			client := connectClient(t, env, user.ID, trendbirdv1connect.NewAuthServiceClient)
			_, err := client.DeleteAccount(context.Background(), connect.NewRequest(&trendbirdv1.DeleteAccountRequest{}))
			ch <- result{err: err}
		}()
	}

	r1 := <-ch
	r2 := <-ch

	// 少なくとも1つは成功
	successCount := 0
	if r1.err == nil {
		successCount++
	}
	if r2.err == nil {
		successCount++
	}
	if successCount == 0 {
		t.Fatalf("both goroutines failed: %v / %v", r1.err, r2.err)
	}

	// users テーブルが0行
	var userCount int64
	env.db.Model(&model.User{}).Where("id = ?", user.ID).Count(&userCount)
	if userCount != 0 {
		t.Errorf("users count: want 0, got %d", userCount)
	}
}

// ---------------------------------------------------------------------------
// TestAuthMiddleware_TokenValidation
// ---------------------------------------------------------------------------

func TestAuthMiddleware_TokenValidation(t *testing.T) {
	now := time.Now()

	t.Run("valid_token_success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		token := generateTestToken(t, user.ID)

		client := trendbirdv1connect.NewAuthServiceClient(
			env.server.Client(),
			env.server.URL,
			connect.WithInterceptors(authTokenInterceptor(token)),
		)
		resp, err := client.GetCurrentUser(context.Background(), connect.NewRequest(&trendbirdv1.GetCurrentUserRequest{}))
		if err != nil {
			t.Fatalf("GetCurrentUser with valid token: %v", err)
		}
		if resp.Msg.GetUser().GetId() != user.ID {
			t.Errorf("user id: want %q, got %q", user.ID, resp.Msg.GetUser().GetId())
		}
	})

	t.Run("expired_token", func(t *testing.T) {
		env := setupTest(t)

		token := generateCustomToken(t, []byte(testJWTSecret), jwt.MapClaims{
			"sub": "some-user-id",
			"iat": now.Add(-2 * time.Hour).Unix(),
			"exp": now.Add(-1 * time.Hour).Unix(),
		}, jwt.SigningMethodHS256)

		client := trendbirdv1connect.NewAuthServiceClient(
			env.server.Client(),
			env.server.URL,
			connect.WithInterceptors(authTokenInterceptor(token)),
		)
		_, err := client.GetCurrentUser(context.Background(), connect.NewRequest(&trendbirdv1.GetCurrentUserRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})

	t.Run("wrong_secret", func(t *testing.T) {
		env := setupTest(t)

		token := generateCustomToken(t, []byte("wrong-secret-key"), jwt.MapClaims{
			"sub": "some-user-id",
			"iat": now.Unix(),
			"exp": now.Add(1 * time.Hour).Unix(),
		}, jwt.SigningMethodHS256)

		client := trendbirdv1connect.NewAuthServiceClient(
			env.server.Client(),
			env.server.URL,
			connect.WithInterceptors(authTokenInterceptor(token)),
		)
		_, err := client.GetCurrentUser(context.Background(), connect.NewRequest(&trendbirdv1.GetCurrentUserRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})

	t.Run("missing_sub_claim", func(t *testing.T) {
		env := setupTest(t)

		token := generateCustomToken(t, []byte(testJWTSecret), jwt.MapClaims{
			"iat": now.Unix(),
			"exp": now.Add(1 * time.Hour).Unix(),
		}, jwt.SigningMethodHS256)

		client := trendbirdv1connect.NewAuthServiceClient(
			env.server.Client(),
			env.server.URL,
			connect.WithInterceptors(authTokenInterceptor(token)),
		)
		_, err := client.GetCurrentUser(context.Background(), connect.NewRequest(&trendbirdv1.GetCurrentUserRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})

	t.Run("empty_token_string", func(t *testing.T) {
		env := setupTest(t)

		client := trendbirdv1connect.NewAuthServiceClient(
			env.server.Client(),
			env.server.URL,
			connect.WithInterceptors(authTokenInterceptor("")),
		)
		_, err := client.GetCurrentUser(context.Background(), connect.NewRequest(&trendbirdv1.GetCurrentUserRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})

	t.Run("malformed_non_jwt_string", func(t *testing.T) {
		env := setupTest(t)

		client := trendbirdv1connect.NewAuthServiceClient(
			env.server.Client(),
			env.server.URL,
			connect.WithInterceptors(authTokenInterceptor("not-a-jwt-token-at-all")),
		)
		_, err := client.GetCurrentUser(context.Background(), connect.NewRequest(&trendbirdv1.GetCurrentUserRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})

	t.Run("bearer_absent_cookie_fallback", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		token := generateTestToken(t, user.ID)

		// Bearer ヘッダーなし、Cookie のみで認証
		client := trendbirdv1connect.NewAuthServiceClient(
			env.server.Client(),
			env.server.URL,
			connect.WithInterceptors(cookieInterceptor("tb_jwt="+token)),
		)
		resp, err := client.GetCurrentUser(context.Background(), connect.NewRequest(&trendbirdv1.GetCurrentUserRequest{}))
		if err != nil {
			t.Fatalf("GetCurrentUser via cookie: %v", err)
		}
		if resp.Msg.GetUser().GetId() != user.ID {
			t.Errorf("user id: want %q, got %q", user.ID, resp.Msg.GetUser().GetId())
		}
	})

	t.Run("rs256_algorithm_rejected", func(t *testing.T) {
		env := setupTest(t)

		rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			t.Fatalf("rsa.GenerateKey: %v", err)
		}

		claims := jwt.MapClaims{
			"sub": "some-user-id",
			"iat": now.Unix(),
			"exp": now.Add(1 * time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		signed, err := token.SignedString(rsaKey)
		if err != nil {
			t.Fatalf("SignedString RS256: %v", err)
		}

		client := trendbirdv1connect.NewAuthServiceClient(
			env.server.Client(),
			env.server.URL,
			connect.WithInterceptors(authTokenInterceptor(signed)),
		)
		_, err = client.GetCurrentUser(context.Background(), connect.NewRequest(&trendbirdv1.GetCurrentUserRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

// ---------------------------------------------------------------------------
// TestAuthHTTP_XAuthRedirect (D1)
// ---------------------------------------------------------------------------

func TestAuthHTTP_XAuthRedirect(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)

		// リダイレクトを追跡しないクライアント
		client := env.server.Client()
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}

		req, err := http.NewRequest(http.MethodGet, env.server.URL+"/auth/x", nil)
		if err != nil {
			t.Fatalf("create request: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("do request: %v", err)
		}
		defer resp.Body.Close()

		// 302 リダイレクトであること
		if resp.StatusCode != http.StatusFound {
			t.Errorf("status: want %d, got %d", http.StatusFound, resp.StatusCode)
		}

		// Location ヘッダが認可 URL であること
		location := resp.Header.Get("Location")
		if location == "" {
			t.Error("expected Location header to be set")
		}
		if location != "https://x.com/authorize?test=1" {
			t.Errorf("Location: want %q, got %q", "https://x.com/authorize?test=1", location)
		}

		// tb_cv Cookie が設定されていること
		cvCookie, ok := findSetCookie(resp.Header, "tb_cv")
		if !ok {
			t.Error("Set-Cookie tb_cv not found in response")
		} else if cvCookie == "tb_cv=" {
			t.Error("Set-Cookie tb_cv value is empty")
		}

		// tb_state Cookie が設定されていること
		stateCookie, ok := findSetCookie(resp.Header, "tb_state")
		if !ok {
			t.Error("Set-Cookie tb_state not found in response")
		} else if stateCookie == "tb_state=" {
			t.Error("Set-Cookie tb_state value is empty")
		}
	})

	t.Run("gateway_error", func(t *testing.T) {
		env := setupTest(t)

		// BuildAuthorizationURL をエラー返却に差し替え
		env.mockTwitter.BuildAuthorizationURLFn = func(_ context.Context) (*gateway.OAuthStartResult, error) {
			return nil, fmt.Errorf("oauth provider unavailable")
		}

		client := env.server.Client()
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}

		req, err := http.NewRequest(http.MethodGet, env.server.URL+"/auth/x", nil)
		if err != nil {
			t.Fatalf("create request: %v", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("do request: %v", err)
		}
		defer resp.Body.Close()

		// 500 レスポンス
		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("status: want %d, got %d", http.StatusInternalServerError, resp.StatusCode)
		}
	})
}
