package e2etest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"connectrpc.com/connect"

	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/gen/trendbird/v1/trendbirdv1connect"
	"github.com/trendbird/backend/internal/domain/gateway"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// ---------------------------------------------------------------------------
// TestTwitterService_GetConnectionInfo
// ---------------------------------------------------------------------------

func TestTwitterService_GetConnectionInfo(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)

		user := seedUser(t, env.db)
		tc := seedTwitterConnection(t, env.db, user.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTwitterServiceClient)
		resp, err := client.GetConnectionInfo(context.Background(), connect.NewRequest(&trendbirdv1.GetConnectionInfoRequest{}))
		if err != nil {
			t.Fatalf("GetConnectionInfo: %v", err)
		}

		info := resp.Msg.GetInfo()
		if info.GetStatus() != trendbirdv1.TwitterConnectionStatus_TWITTER_CONNECTION_STATUS_CONNECTED {
			t.Errorf("status: want CONNECTED, got %v", info.GetStatus())
		}
		if info.ConnectedAt == nil || *info.ConnectedAt == "" {
			t.Errorf("connected_at should not be empty")
		}
		// last_tested_at フィールド検証 (B3)
		// seed 時に last_tested_at は nil なので、レスポンスも nil（未設定）であること
		if tc.LastTestedAt == nil && info.LastTestedAt != nil {
			t.Errorf("last_tested_at: want nil, got %v", *info.LastTestedAt)
		}
	})

	t.Run("disconnected", func(t *testing.T) {
		env := setupTest(t)

		user := seedUser(t, env.db)
		// 接続なし

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTwitterServiceClient)
		_, err := client.GetConnectionInfo(context.Background(), connect.NewRequest(&trendbirdv1.GetConnectionInfoRequest{}))
		assertConnectCode(t, err, connect.CodeNotFound)
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.twitterClient.GetConnectionInfo(context.Background(), connect.NewRequest(&trendbirdv1.GetConnectionInfoRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

// ---------------------------------------------------------------------------
// TestTwitterService_TestConnection
// ---------------------------------------------------------------------------

func TestTwitterService_TestConnection(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)

		user := seedUser(t, env.db)
		seedTwitterConnection(t, env.db, user.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTwitterServiceClient)
		resp, err := client.TestConnection(context.Background(), connect.NewRequest(&trendbirdv1.TestConnectionRequest{}))
		if err != nil {
			t.Fatalf("TestConnection: %v", err)
		}

		info := resp.Msg.GetInfo()
		if info.GetStatus() != trendbirdv1.TwitterConnectionStatus_TWITTER_CONNECTION_STATUS_CONNECTED {
			t.Errorf("status: want CONNECTED, got %v", info.GetStatus())
		}
		// last_tested_at が更新されているか
		if info.LastTestedAt == nil || *info.LastTestedAt == "" {
			t.Errorf("last_tested_at should be updated")
		}

		// DB 直接検証: status=CONNECTED + last_tested_at が更新されていること (E3)
		var dbConn model.TwitterConnection
		if err := env.db.Where("user_id = ?", user.ID).First(&dbConn).Error; err != nil {
			t.Fatalf("query twitter_connection: %v", err)
		}
		if dbConn.Status != 3 { // Connected
			t.Errorf("DB status: want 3 (Connected), got %d", dbConn.Status)
		}
		if dbConn.LastTestedAt == nil {
			t.Error("DB last_tested_at should be set after TestConnection")
		}
	})

	t.Run("success_token_refresh", func(t *testing.T) {
		env := setupTest(t)

		user := seedUser(t, env.db)
		// トークン期限切れの接続を seed
		seedTwitterConnection(t, env.db, user.ID,
			withTokenExpiresAt(time.Now().Add(-1*time.Hour)),
		)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTwitterServiceClient)
		_, err := client.TestConnection(context.Background(), connect.NewRequest(&trendbirdv1.TestConnectionRequest{}))
		if err != nil {
			t.Fatalf("TestConnection: %v", err)
		}

		// DB のアクセストークンがリフレッシュ後の値に更新されていることを検証
		var tc model.TwitterConnection
		if err := env.db.Where("user_id = ?", user.ID).First(&tc).Error; err != nil {
			t.Fatalf("query twitter_connection: %v", err)
		}
		if tc.AccessToken != "refreshed-access-token" {
			t.Errorf("access_token: want refreshed-access-token, got %s", tc.AccessToken)
		}
	})

	t.Run("token_refresh_failure", func(t *testing.T) {
		env := setupTest(t)

		user := seedUser(t, env.db)
		// トークン期限切れの接続を seed
		seedTwitterConnection(t, env.db, user.ID,
			withTokenExpiresAt(time.Now().Add(-1*time.Hour)),
		)

		// RefreshTokenFn をエラー返却に差し替え
		env.mockTwitter.RefreshTokenFn = func(_ context.Context, _ string) (*gateway.OAuthTokenResponse, error) {
			return nil, fmt.Errorf("refresh token expired")
		}

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTwitterServiceClient)
		_, err := client.TestConnection(context.Background(), connect.NewRequest(&trendbirdv1.TestConnectionRequest{}))
		assertConnectCode(t, err, connect.CodeInternal)

		// DB: twitter_connection の status が Error になっていることを検証
		var tc model.TwitterConnection
		if err := env.db.Where("user_id = ?", user.ID).First(&tc).Error; err != nil {
			t.Fatalf("query twitter_connection: %v", err)
		}
		if tc.Status != 4 { // Error
			t.Errorf("status: want 4 (Error), got %d", tc.Status)
		}
		if tc.ErrorMessage == nil || *tc.ErrorMessage == "" {
			t.Error("error_message should not be empty")
		}
	})

	t.Run("failure_verify_credentials", func(t *testing.T) {
		env := setupTest(t)

		user := seedUser(t, env.db)
		seedTwitterConnection(t, env.db, user.ID)

		// VerifyCredentials を失敗させる
		env.mockTwitter.VerifyCredentialsFn = func(_ context.Context, _ string) error {
			return fmt.Errorf("invalid or expired token")
		}

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTwitterServiceClient)
		_, err := client.TestConnection(context.Background(), connect.NewRequest(&trendbirdv1.TestConnectionRequest{}))
		assertConnectCode(t, err, connect.CodeInternal)

		// DB: status=4(Error), error_message 非空
		var tc model.TwitterConnection
		if err := env.db.Where("user_id = ?", user.ID).First(&tc).Error; err != nil {
			t.Fatalf("query twitter_connection: %v", err)
		}
		if tc.Status != 4 { // Error
			t.Errorf("status: want 4 (Error), got %d", tc.Status)
		}
		if tc.ErrorMessage == nil || *tc.ErrorMessage == "" {
			t.Errorf("error_message should not be empty")
		}
	})
}

// ---------------------------------------------------------------------------
// TestTwitterService_DisconnectTwitter
// ---------------------------------------------------------------------------

func TestTwitterService_DisconnectTwitter(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)

		user := seedUser(t, env.db)
		seedTwitterConnection(t, env.db, user.ID)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewTwitterServiceClient)
		_, err := client.DisconnectTwitter(context.Background(), connect.NewRequest(&trendbirdv1.DisconnectTwitterRequest{}))
		if err != nil {
			t.Fatalf("DisconnectTwitter: %v", err)
		}

		// DB: twitter_connections が 0件
		var count int64
		env.db.Model(&model.TwitterConnection{}).Where("user_id = ?", user.ID).Count(&count)
		if count != 0 {
			t.Errorf("twitter_connections count: want 0, got %d", count)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.twitterClient.DisconnectTwitter(context.Background(), connect.NewRequest(&trendbirdv1.DisconnectTwitterRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}
