package e2etest

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"

	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/gen/trendbird/v1/trendbirdv1connect"
	"github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

func TestNotificationService_ListNotifications(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		// 通知3件を seed（Type=TREND×2 + SYSTEM×1、1件は既読）
		n1 := seedNotification(t, env.db, user.ID)
		n2 := seedNotification(t, env.db, user.ID, withNotificationType(2)) // SYSTEM
		n3 := seedNotification(t, env.db, user.ID, withNotificationRead(true))

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewNotificationServiceClient)
		resp, err := client.ListNotifications(context.Background(), connect.NewRequest(&trendbirdv1.ListNotificationsRequest{}))
		if err != nil {
			t.Fatalf("ListNotifications: %v", err)
		}

		notifications := resp.Msg.GetNotifications()
		if got := len(notifications); got != 3 {
			t.Fatalf("expected 3 notifications, got %d", got)
		}

		// created_at DESC 順で返却されるので n3 → n2 → n1 の順
		wantOrder := []*model.Notification{n3, n2, n1}
		wantIsRead := []bool{true, false, false} // n3=read, n2=unread, n1=unread
		for i, want := range wantOrder {
			got := notifications[i]
			if got.GetId() != want.ID {
				t.Errorf("[%d] id: want %s, got %s", i, want.ID, got.GetId())
			}
			if int32(got.GetType()) != want.Type {
				t.Errorf("[%d] type: want %d, got %d", i, want.Type, got.GetType())
			}
			if got.GetTitle() != want.Title {
				t.Errorf("[%d] title: want %s, got %s", i, want.Title, got.GetTitle())
			}
			if got.GetMessage() != want.Message {
				t.Errorf("[%d] message: want %s, got %s", i, want.Message, got.GetMessage())
			}
			if got.GetIsRead() != wantIsRead[i] {
				t.Errorf("[%d] is_read: want %v, got %v", i, wantIsRead[i], got.GetIsRead())
			}

			// timestamp は RFC3339 フォーマット
			ts, parseErr := time.Parse(time.RFC3339, got.GetTimestamp())
			if parseErr != nil {
				t.Errorf("[%d] timestamp parse error: %v", i, parseErr)
			} else if ts.Sub(want.CreatedAt).Abs() > 2*time.Second {
				t.Errorf("[%d] timestamp: want ~%v, got %v", i, want.CreatedAt, ts)
			}
		}

		// 降順ソートの追加検証: 各 timestamp が前の要素以下であること
		for i := 1; i < len(notifications); i++ {
			prev, _ := time.Parse(time.RFC3339, notifications[i-1].GetTimestamp())
			curr, _ := time.Parse(time.RFC3339, notifications[i].GetTimestamp())
			if curr.After(prev) {
				t.Errorf("notifications not in DESC order: [%d]=%v > [%d]=%v", i, curr, i-1, prev)
			}
		}
	})

	t.Run("empty", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewNotificationServiceClient)
		resp, err := client.ListNotifications(context.Background(), connect.NewRequest(&trendbirdv1.ListNotificationsRequest{}))
		if err != nil {
			t.Fatalf("ListNotifications: %v", err)
		}

		if got := len(resp.Msg.GetNotifications()); got != 0 {
			t.Fatalf("expected 0 notifications, got %d", got)
		}
	})

	t.Run("cross_user_isolation", func(t *testing.T) {
		env := setupTest(t)
		owner := seedUser(t, env.db)
		other := seedUser(t, env.db)

		// owner に2件、other に1件
		seedNotification(t, env.db, owner.ID)
		seedNotification(t, env.db, owner.ID)
		otherNotif := seedNotification(t, env.db, other.ID)

		// other が ListNotifications → 自分の1件のみ取得
		client := connectClient(t, env, other.ID, trendbirdv1connect.NewNotificationServiceClient)
		resp, err := client.ListNotifications(context.Background(), connect.NewRequest(&trendbirdv1.ListNotificationsRequest{}))
		if err != nil {
			t.Fatalf("ListNotifications: %v", err)
		}

		notifications := resp.Msg.GetNotifications()
		if got := len(notifications); got != 1 {
			t.Fatalf("expected 1 notification for other, got %d", got)
		}
		if notifications[0].GetTitle() != otherNotif.Title {
			t.Errorf("title: want %s, got %s", otherNotif.Title, notifications[0].GetTitle())
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		// 認証なしクライアント（env.notificationClient）で呼び出し
		_, err := env.notificationClient.ListNotifications(context.Background(), connect.NewRequest(&trendbirdv1.ListNotificationsRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

func TestNotificationService_MarkAsRead(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)
		n := seedNotification(t, env.db, user.ID) // 未読

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewNotificationServiceClient)
		_, err := client.MarkAsRead(context.Background(), connect.NewRequest(&trendbirdv1.MarkAsReadRequest{
			Id: n.ID,
		}))
		if err != nil {
			t.Fatalf("MarkAsRead: %v", err)
		}

		// DB (user_notifications) で既読になっていることを確認
		var un model.UserNotification
		if err := env.db.Where("user_id = ? AND notification_id = ?", user.ID, n.ID).First(&un).Error; err != nil {
			t.Fatalf("failed to fetch user_notification: %v", err)
		}
		if !un.IsRead {
			t.Error("expected is_read=true after MarkAsRead")
		}
	})

	t.Run("not_found", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewNotificationServiceClient)
		_, err := client.MarkAsRead(context.Background(), connect.NewRequest(&trendbirdv1.MarkAsReadRequest{
			Id: "00000000-0000-0000-0000-000000000000",
		}))
		assertConnectCode(t, err, connect.CodeNotFound)
	})

	// 実装は WHERE id = ? AND user_id = ? で検索するため、他ユーザーの通知は
	// 「見つからない」として CodeNotFound を返す（CodePermissionDenied ではない）。
	t.Run("other_user_not_found", func(t *testing.T) {
		env := setupTest(t)
		owner := seedUser(t, env.db)
		other := seedUser(t, env.db)
		n := seedNotification(t, env.db, owner.ID) // owner の通知

		// other が owner の通知を既読にしようとする → NotFound
		client := connectClient(t, env, other.ID, trendbirdv1connect.NewNotificationServiceClient)
		_, err := client.MarkAsRead(context.Background(), connect.NewRequest(&trendbirdv1.MarkAsReadRequest{
			Id: n.ID,
		}))
		assertConnectCode(t, err, connect.CodeNotFound)

		// DB (user_notifications) 上で通知が未読のままであることを検証
		var ownerUN model.UserNotification
		if err := env.db.Where("user_id = ? AND notification_id = ?", owner.ID, n.ID).First(&ownerUN).Error; err != nil {
			t.Fatalf("failed to fetch user_notification: %v", err)
		}
		if ownerUN.IsRead {
			t.Error("expected is_read=false, notification should remain unread")
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.notificationClient.MarkAsRead(context.Background(), connect.NewRequest(&trendbirdv1.MarkAsReadRequest{
			Id: "some-id",
		}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}

func TestNotificationService_MarkAllAsRead(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		env := setupTest(t)
		user := seedUser(t, env.db)

		// 未読3件 + 既読1件
		seedNotification(t, env.db, user.ID)
		seedNotification(t, env.db, user.ID)
		seedNotification(t, env.db, user.ID)
		seedNotification(t, env.db, user.ID, withNotificationRead(true))

		client := connectClient(t, env, user.ID, trendbirdv1connect.NewNotificationServiceClient)
		_, err := client.MarkAllAsRead(context.Background(), connect.NewRequest(&trendbirdv1.MarkAllAsReadRequest{}))
		if err != nil {
			t.Fatalf("MarkAllAsRead: %v", err)
		}

		// DB (user_notifications) 上で全4件が is_read=true であることを確認
		var unreadCount int64
		if err := env.db.Model(&model.UserNotification{}).Where("user_id = ? AND is_read = ?", user.ID, false).Count(&unreadCount).Error; err != nil {
			t.Fatalf("failed to count unread: %v", err)
		}
		if unreadCount != 0 {
			t.Errorf("expected 0 unread notifications, got %d", unreadCount)
		}

		var totalCount int64
		if err := env.db.Model(&model.UserNotification{}).Where("user_id = ?", user.ID).Count(&totalCount).Error; err != nil {
			t.Fatalf("failed to count total: %v", err)
		}
		if totalCount != 4 {
			t.Errorf("expected 4 total notifications, got %d", totalCount)
		}
	})

	t.Run("cross_user_isolation", func(t *testing.T) {
		env := setupTest(t)
		owner := seedUser(t, env.db)
		other := seedUser(t, env.db)

		// owner に未読2件、other に未読1件
		seedNotification(t, env.db, owner.ID)
		seedNotification(t, env.db, owner.ID)
		seedNotification(t, env.db, other.ID)

		// other が MarkAllAsRead
		client := connectClient(t, env, other.ID, trendbirdv1connect.NewNotificationServiceClient)
		_, err := client.MarkAllAsRead(context.Background(), connect.NewRequest(&trendbirdv1.MarkAllAsReadRequest{}))
		if err != nil {
			t.Fatalf("MarkAllAsRead: %v", err)
		}

		// other の通知は全て既読
		var otherUnread int64
		if err := env.db.Model(&model.UserNotification{}).Where("user_id = ? AND is_read = ?", other.ID, false).Count(&otherUnread).Error; err != nil {
			t.Fatalf("failed to count other's unread: %v", err)
		}
		if otherUnread != 0 {
			t.Errorf("expected 0 unread for other, got %d", otherUnread)
		}

		// owner の通知は未読のまま
		var ownerUnread int64
		if err := env.db.Model(&model.UserNotification{}).Where("user_id = ? AND is_read = ?", owner.ID, false).Count(&ownerUnread).Error; err != nil {
			t.Fatalf("failed to count owner's unread: %v", err)
		}
		if ownerUnread != 2 {
			t.Errorf("expected 2 unread for owner, got %d", ownerUnread)
		}
	})

	t.Run("unauthenticated", func(t *testing.T) {
		env := setupTest(t)

		_, err := env.notificationClient.MarkAllAsRead(context.Background(), connect.NewRequest(&trendbirdv1.MarkAllAsReadRequest{}))
		assertConnectCode(t, err, connect.CodeUnauthenticated)
	})
}
