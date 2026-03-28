import { test, expect, authenticateContext } from '../fixtures/test-base';
import {
  resetSeq,
  TopicStatus,
  NotificationType,
  buildNotification,
} from '../fixtures/factories';
import { NotificationsPage } from '../pages/notifications.page';
import { create, toJson } from '@bufbuild/protobuf';
import {
  NotificationSchema,
  ListNotificationsResponseSchema,
  MarkAsReadResponseSchema,
  MarkAllAsReadResponseSchema,
} from '../../src/gen/trendbird/v1/notification_pb';

test.describe('通知ページ', () => {
  test.beforeEach(async ({ context, page, apiMock }) => {
    resetSeq();
    await authenticateContext(context);
    await apiMock.setupDefaults();
    await page.addInitScript(() => {
      sessionStorage.removeItem('tb_tutorial_pending');
    });
  });

  test.describe('通知一覧表示', () => {
    test('通知カード一覧が表示される（未読ドット・タイトル・メッセージ・タイムスタンプ）', async ({ page, apiMock }) => {
      const trendNotif = buildNotification({
        type: NotificationType.TREND,
        title: 'AIトピックが盛り上がっています',
        message: 'z-score が 3.5 を超えました',
        isRead: false,
        topicId: 'topic-1',
        topicName: 'AI Topic',
        topicStatus: TopicStatus.SPIKE,
      });
      const systemNotif = buildNotification({
        type: NotificationType.SYSTEM,
        title: '新機能のお知らせ',
        message: '予約投稿機能がリリースされました',
        isRead: true,
      });

      await apiMock.clearMock('NotificationService', 'ListNotifications');
      await apiMock.mockRPC('NotificationService', 'ListNotifications',
        toJson(ListNotificationsResponseSchema,
          create(ListNotificationsResponseSchema, {
            notifications: [
              create(NotificationSchema, {
                id: 'notif-1',
                type: NotificationType.TREND,
                title: 'AIトピックが盛り上がっています',
                message: 'z-score が 3.5 を超えました',
                timestamp: '2026-01-01T12:00:00Z',
                isRead: false,
                topicId: 'topic-1',
                topicName: 'AI Topic',
                topicStatus: TopicStatus.SPIKE,
              }),
              create(NotificationSchema, {
                id: 'notif-2',
                type: NotificationType.SYSTEM,
                title: '新機能のお知らせ',
                message: '予約投稿機能がリリースされました',
                timestamp: '2026-01-01T10:00:00Z',
                isRead: true,
              }),
            ],
          })));

      const notifications = new NotificationsPage(page);
      await notifications.goto();

      await expect(notifications.heading).toBeVisible();
      await expect(page.getByText('AIトピックが盛り上がっています')).toBeVisible();
      await expect(page.getByText('z-score が 3.5 を超えました')).toBeVisible();
      await expect(page.getByText('新機能のお知らせ')).toBeVisible();
    });

    test('タブ切替でフィルタリングされる', async ({ page, apiMock }) => {
      await apiMock.clearMock('NotificationService', 'ListNotifications');
      await apiMock.mockRPC('NotificationService', 'ListNotifications',
        toJson(ListNotificationsResponseSchema,
          create(ListNotificationsResponseSchema, {
            notifications: [
              create(NotificationSchema, {
                id: 'notif-trend',
                type: NotificationType.TREND,
                title: 'AIトピックが急上昇',
                message: 'トレンドメッセージ',
                timestamp: '2026-01-01T12:00:00Z',
                isRead: false,
                topicId: 'topic-1',
                topicName: 'Topic 1',
              }),
              create(NotificationSchema, {
                id: 'notif-system',
                type: NotificationType.SYSTEM,
                title: 'メンテナンスのお知らせ',
                message: 'システムメッセージ',
                timestamp: '2026-01-01T10:00:00Z',
                isRead: false,
              }),
            ],
          })));

      const notifications = new NotificationsPage(page);
      await notifications.goto();

      // 「すべて」タブでは両方表示
      await expect(page.getByText('AIトピックが急上昇')).toBeVisible();
      await expect(page.getByText('メンテナンスのお知らせ')).toBeVisible();

      // 「トレンド通知」タブ → トレンドのみ
      await notifications.trendTab.click();
      await expect(page.getByText('AIトピックが急上昇')).toBeVisible();
      await expect(page.getByText('メンテナンスのお知らせ')).not.toBeVisible();

      // 「運営」タブ → システムのみ
      await notifications.systemTab.click();
      await expect(page.getByText('メンテナンスのお知らせ')).toBeVisible();
      await expect(page.getByText('AIトピックが急上昇')).not.toBeVisible();
    });
  });

  test.describe('既読操作', () => {
    test('カードクリックで既読化 + ページ遷移', async ({ page, apiMock }) => {
      await apiMock.clearMock('NotificationService', 'ListNotifications');
      await apiMock.mockRPC('NotificationService', 'ListNotifications',
        toJson(ListNotificationsResponseSchema,
          create(ListNotificationsResponseSchema, {
            notifications: [
              create(NotificationSchema, {
                id: 'notif-click',
                type: NotificationType.TREND,
                title: 'クリックテスト通知',
                message: 'クリックで既読になる',
                timestamp: '2026-01-01T12:00:00Z',
                isRead: false,
                topicId: 'default-topic-1',
                topicName: 'Default Topic 1',
              }),
            ],
          })));
      await apiMock.mockRPC('NotificationService', 'MarkAsRead',
        toJson(MarkAsReadResponseSchema, create(MarkAsReadResponseSchema, {})));
      const notifications = new NotificationsPage(page);
      await notifications.goto();

      // 通知カードをクリック
      await page.getByText('クリックテスト通知').click();
      // トピック詳細ページに遷移
      await page.waitForURL('**/dashboard/default-topic-1');
    });

    test('システム通知クリックで actionUrl に遷移', async ({ page, apiMock }) => {
      await apiMock.clearMock('NotificationService', 'ListNotifications');
      await apiMock.mockRPC('NotificationService', 'ListNotifications',
        toJson(ListNotificationsResponseSchema,
          create(ListNotificationsResponseSchema, {
            notifications: [
              create(NotificationSchema, {
                id: 'notif-system-action',
                type: NotificationType.SYSTEM,
                title: 'お知らせ',
                message: '新機能が追加されました',
                timestamp: '2026-01-01T12:00:00Z',
                isRead: false,
                actionUrl: '/dashboard',
              }),
            ],
          })));
      await apiMock.mockRPC('NotificationService', 'MarkAsRead',
        toJson(MarkAsReadResponseSchema, create(MarkAsReadResponseSchema, {})));

      const notifications = new NotificationsPage(page);
      await notifications.goto();

      // システム通知をクリック → actionUrl に遷移
      await page.getByText('お知らせ').click();
      await page.waitForURL('**/dashboard');
    });

    test('「すべて既読にする」で全未読ドット消失', async ({ page, apiMock }) => {
      await apiMock.clearMock('NotificationService', 'ListNotifications');
      await apiMock.mockRPC('NotificationService', 'ListNotifications',
        toJson(ListNotificationsResponseSchema,
          create(ListNotificationsResponseSchema, {
            notifications: [
              create(NotificationSchema, {
                id: 'notif-unread-1',
                type: NotificationType.TREND,
                title: '未読通知1',
                message: 'メッセージ1',
                timestamp: '2026-01-01T12:00:00Z',
                isRead: false,
                topicId: 'topic-1',
              }),
              create(NotificationSchema, {
                id: 'notif-unread-2',
                type: NotificationType.TREND,
                title: '未読通知2',
                message: 'メッセージ2',
                timestamp: '2026-01-01T11:00:00Z',
                isRead: false,
                topicId: 'topic-2',
              }),
            ],
          })));
      await apiMock.mockRPC('NotificationService', 'MarkAllAsRead',
        toJson(MarkAllAsReadResponseSchema, create(MarkAllAsReadResponseSchema, {})));

      const notifications = new NotificationsPage(page);
      await notifications.goto();

      // 「すべて既読にする」ボタンが表示される
      await expect(notifications.markAllReadButton).toBeVisible();

      // クリック
      await notifications.markAllReadButton.click();

      // ボタンが消える（未読がなくなったため）
      await expect(notifications.markAllReadButton).not.toBeVisible();
    });
  });

  test.describe('一括既読の状態遷移', () => {
    test('一括既読後も通知カードは表示されたまま', async ({ page, apiMock }) => {
      await apiMock.clearMock('NotificationService', 'ListNotifications');
      await apiMock.mockRPC('NotificationService', 'ListNotifications',
        toJson(ListNotificationsResponseSchema,
          create(ListNotificationsResponseSchema, {
            notifications: [
              create(NotificationSchema, {
                id: 'notif-keep',
                type: NotificationType.TREND,
                title: '既読後も残る通知',
                message: 'テストメッセージ',
                timestamp: '2026-01-01T12:00:00Z',
                isRead: false,
                topicId: 'topic-1',
              }),
            ],
          })));
      await apiMock.mockRPC('NotificationService', 'MarkAllAsRead',
        toJson(MarkAllAsReadResponseSchema, create(MarkAllAsReadResponseSchema, {})));

      const notifications = new NotificationsPage(page);
      await notifications.goto();

      await expect(page.getByText('既読後も残る通知')).toBeVisible();
      await notifications.markAllReadButton.click();

      // ボタンは非表示（未読がなくなったため）
      await expect(notifications.markAllReadButton).not.toBeVisible();
      // 通知カード自体は表示されたまま
      await expect(page.getByText('既読後も残る通知')).toBeVisible();
    });
  });

  test.describe('状態表示', () => {
    test('通知0件で空状態が表示される', async ({ page }) => {
      // setupDefaults で空の ListNotifications が返る
      const notifications = new NotificationsPage(page);
      await notifications.goto();

      await expect(notifications.emptyState).toBeVisible();
    });

    test('API エラーでリトライボタンが表示される', async ({ page, apiMock }) => {
      await apiMock.clearMock('NotificationService', 'ListNotifications');
      await apiMock.mockRPCError('NotificationService', 'ListNotifications', 'internal', 'サーバーエラー');

      const notifications = new NotificationsPage(page);
      await notifications.goto();

      await expect(notifications.errorState).toBeVisible();
      await expect(notifications.retryButton).toBeVisible();
    });
  });

  test.describe('リトライ', () => {
    test('リトライクリック → 成功 → 通知リスト表示', async ({ page, apiMock }) => {
      // 1回目: エラー
      await apiMock.clearMock('NotificationService', 'ListNotifications');
      await apiMock.mockRPCError('NotificationService', 'ListNotifications', 'internal', 'サーバーエラー');

      const notifications = new NotificationsPage(page);
      await notifications.goto();

      await expect(notifications.errorState).toBeVisible();
      await expect(notifications.retryButton).toBeVisible();

      // 2回目: 成功に切り替え
      await apiMock.clearMock('NotificationService', 'ListNotifications');
      await apiMock.mockRPC('NotificationService', 'ListNotifications',
        toJson(ListNotificationsResponseSchema,
          create(ListNotificationsResponseSchema, {
            notifications: [
              create(NotificationSchema, {
                id: 'notif-retry',
                type: NotificationType.TREND,
                title: 'リトライ後の通知',
                message: 'リトライで復旧しました',
                timestamp: '2026-01-01T12:00:00Z',
                isRead: false,
                topicId: 'topic-1',
                topicName: 'Topic 1',
              }),
            ],
          })));

      // リトライボタンクリック
      await notifications.retryButton.click();

      // エラー表示が消え、通知が表示される
      await expect(notifications.errorState).not.toBeVisible();
      await expect(page.getByText('リトライ後の通知')).toBeVisible();
      await expect(page.getByText('リトライで復旧しました')).toBeVisible();
    });
  });
});
