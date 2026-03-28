import { test, expect, authenticateContext } from '../fixtures/test-base';
import {
  resetSeq,
} from '../fixtures/factories';
import { expectToast } from '../helpers/assertions';
import { SettingsPage } from '../pages/settings.page';
import { create, toJson } from '@bufbuild/protobuf';
import {
  UpdateProfileResponseSchema,
} from '../../src/gen/trendbird/v1/settings_pb';
import {
  UserSchema,
} from '../../src/gen/trendbird/v1/auth_pb';
import {
  TwitterConnectionInfoSchema,
  TwitterConnectionStatus,
  GetConnectionInfoResponseSchema,
  DisconnectTwitterResponseSchema,
} from '../../src/gen/trendbird/v1/twitter_pb';

test.describe('設定ページ', () => {
  test.beforeEach(async ({ context, page, apiMock }) => {
    resetSeq();
    await authenticateContext(context);
    await apiMock.setupDefaults();
    await page.addInitScript(() => {
      sessionStorage.removeItem('tb_tutorial_pending');
    });
  });

  // ─── タブナビゲーション ──────────────────────────────────────

  test.describe('タブナビゲーション', () => {
    test('デフォルトでプロフィールタブが表示される', async ({ page }) => {
      const settings = new SettingsPage(page);
      await settings.goto();

      await expect(settings.heading).toBeVisible();
      await expect(settings.emailInput).toBeVisible();
      await expect(settings.saveButton).toBeVisible();
    });

    test('URLクエリパラメータでタブが選択される', async ({ page }) => {
      const settings = new SettingsPage(page);
      await settings.goto('twitter');

      // X連携タブがアクティブ
      await expect(page.getByText('Xアカウント連携')).toBeVisible();

      // dashboard に遷移
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');

      // ブラウザバック
      await page.goBack();
      await page.waitForURL('**/settings?tab=twitter');

      // X連携タブが再び表示される
      await expect(page.getByText('Xアカウント連携')).toBeVisible();
    });

    test('タブ切替でコンテンツが切り替わる', async ({ page }) => {
      const settings = new SettingsPage(page);
      await settings.goto();

      // プロフィール → 通知
      await settings.switchTab('notifications');
      await expect(page.getByText('盛り上がり通知')).toBeVisible();

      // 通知 → X連携
      await settings.switchTab('twitter');
      await expect(page.getByText('Xアカウント連携')).toBeVisible();

      // X連携 → アカウント
      await settings.switchTab('account');
      await expect(settings.logoutButton).toBeVisible();
    });
  });

  // ─── プロフィールタブ ────────────────────────────────────────

  test.describe('プロフィールタブ', () => {
    test('ユーザー名・Xハンドルが表示される', async ({ page }) => {
      const settings = new SettingsPage(page);
      await settings.goto();

      await expect(page.getByText('Default User')).toBeVisible();
      await expect(page.getByText('defaultuser')).toBeVisible();
    });

    test('メールアドレスを保存できる', async ({ page, apiMock }) => {
      const user = create(UserSchema, {
        id: 'default-user',
        name: 'Default User',
        email: 'new@example.com',
        twitterHandle: 'defaultuser',

        createdAt: '2026-01-01T00:00:00Z',
      });
      await apiMock.mockRPC('SettingsService', 'UpdateProfile',
        toJson(UpdateProfileResponseSchema,
          create(UpdateProfileResponseSchema, { user })));

      const settings = new SettingsPage(page);
      await settings.goto();

      await settings.emailInput.fill('new@example.com');
      await settings.saveButton.click();

      await expect(page.getByText('メールアドレスを保存しました')).toBeVisible();
    });

    test('無効なメールアドレスでバリデーションエラー', async ({ page }) => {
      const settings = new SettingsPage(page);
      await settings.goto();

      await settings.emailInput.fill('invalid-email');
      await settings.saveButton.click();

      await expect(page.getByText('有効なメールアドレスを入力してください')).toBeVisible();
    });

    test('空のメールアドレスでバリデーションエラー', async ({ page }) => {
      const settings = new SettingsPage(page);
      await settings.goto();

      await settings.emailInput.fill('');
      await settings.saveButton.click();

      await expect(page.getByText('有効なメールアドレスを入力してください')).toBeVisible();
    });

    test('プラス記号付きメールアドレスが有効', async ({ page, apiMock }) => {
      await apiMock.mockRPC('SettingsService', 'UpdateProfile',
        toJson(UpdateProfileResponseSchema,
          create(UpdateProfileResponseSchema, {
            user: create(UserSchema, {
              id: 'default-user',
              name: 'Default User',
              email: 'user+tag@example.com',
              twitterHandle: 'defaultuser',
      
              createdAt: '2026-01-01T00:00:00Z',
            }),
          })));

      const settings = new SettingsPage(page);
      await settings.goto();

      await settings.emailInput.fill('user+tag@example.com');
      await settings.saveButton.click();

      await expect(page.getByText('メールアドレスを保存しました')).toBeVisible();
    });

    test('プロフィール保存失敗でエラー toast', async ({ page, apiMock }) => {
      await apiMock.mockRPCError('SettingsService', 'UpdateProfile', 'internal', 'server error');

      const settings = new SettingsPage(page);
      await settings.goto();

      await settings.emailInput.fill('new@example.com');
      await settings.saveButton.click();

      await expect(page.getByText('保存がうまくいきませんでした')).toBeVisible({ timeout: 5_000 });
    });
  });

  // ─── 通知タブ ────────────────────────────────────────────────

  test.describe('通知タブ', () => {
    test('トグルON/OFFで API 呼び出し + toast 表示', async ({ page, apiMock }) => {
      await apiMock.mockRPC('SettingsService', 'UpdateNotifications', { updated: true });

      const settings = new SettingsPage(page);
      await settings.goto('notifications');

      // デフォルトで spikeEnabled=true → クリックでOFF
      await settings.notificationToggle('盛り上がり通知').click();

      await expect(page.getByText('盛り上がり通知を無効にしました')).toBeVisible();
    });

    test('トグル失敗時にロールバック + エラー toast', async ({ page, apiMock }) => {
      await apiMock.mockRPCError('SettingsService', 'UpdateNotifications', 'internal', 'server error');

      const settings = new SettingsPage(page);
      await settings.goto('notifications');

      // デフォルトで spikeEnabled=true → クリックでOFF（失敗するのでロールバック）
      await settings.notificationToggle('盛り上がり通知').click();

      await expect(page.getByText('通知設定の変更がうまくいきませんでした')).toBeVisible();
    });

    test('2つのトグルが全て表示される', async ({ page }) => {
      const settings = new SettingsPage(page);
      await settings.goto('notifications');

      await expect(page.getByText('盛り上がり通知')).toBeVisible();
      await expect(page.getByText('上昇中通知')).toBeVisible();
    });
  });

  // ─── X連携タブ ───────────────────────────────────────────────

  test.describe('X連携タブ', () => {
    test('未接続時に連携ボタンが表示される', async ({ page }) => {
      const settings = new SettingsPage(page);
      await settings.goto('twitter');

      await expect(settings.connectButton).toBeVisible();
      await expect(page.getByText('投稿機能を利用するにはXアカウントの連携が必要です。')).toBeVisible();
    });

    test('接続済みでユーザー情報が表示される', async ({ page, apiMock }) => {
      await apiMock.clearMock('TwitterService', 'GetConnectionInfo');
      await apiMock.mockRPC('TwitterService', 'GetConnectionInfo',
        toJson(GetConnectionInfoResponseSchema,
          create(GetConnectionInfoResponseSchema, {
            info: create(TwitterConnectionInfoSchema, {
              status: TwitterConnectionStatus.CONNECTED,
              connectedAt: '2026-01-15T10:00:00Z',
            }),
          })));

      const settings = new SettingsPage(page);
      await settings.goto('twitter');

      await expect(page.getByText('Default User')).toBeVisible();
      await expect(settings.reconnectButton).toBeVisible();
      await expect(settings.disconnectButton).toBeVisible();
      await expect(page.getByText('接続日時:')).toBeVisible();
    });

    test('接続済みで再接続ボタンが表示され、return_to をセットする', async ({ page, apiMock }) => {
      await apiMock.clearMock('TwitterService', 'GetConnectionInfo');
      await apiMock.mockRPC('TwitterService', 'GetConnectionInfo',
        toJson(GetConnectionInfoResponseSchema,
          create(GetConnectionInfoResponseSchema, {
            info: create(TwitterConnectionInfoSchema, {
              status: TwitterConnectionStatus.CONNECTED,
              connectedAt: '2026-01-15T10:00:00Z',
            }),
          })));

      const settings = new SettingsPage(page);
      await settings.goto('twitter');

      await expect(settings.reconnectButton).toBeVisible();

      // handleReconnect: localStorage.setItem → window.location.href の順で実行される
      // OAuth URL への遷移をインターセプトし、localStorage の値を検証する
      await page.route('**/auth/x', (route) => {
        route.fulfill({ status: 200, contentType: 'text/html', body: '<html><body>intercepted</body></html>' });
      });

      await settings.reconnectButton.click();

      // インターセプトされたページに遷移する → 元のオリジンに戻って localStorage を確認
      await page.waitForTimeout(500);
      await page.goto('/settings');
      const returnTo = await page.evaluate(() => localStorage.getItem('tb_return_to'));
      expect(returnTo).toBe('/settings?tab=twitter');
    });

    test('連携解除 → 確認ダイアログ → /login へ', async ({ page, apiMock }) => {
      await apiMock.clearMock('TwitterService', 'GetConnectionInfo');
      await apiMock.mockRPC('TwitterService', 'GetConnectionInfo',
        toJson(GetConnectionInfoResponseSchema,
          create(GetConnectionInfoResponseSchema, {
            info: create(TwitterConnectionInfoSchema, {
              status: TwitterConnectionStatus.CONNECTED,
              connectedAt: '2026-01-15T10:00:00Z',
            }),
          })));
      await apiMock.mockRPC('TwitterService', 'DisconnectTwitter',
        toJson(DisconnectTwitterResponseSchema,
          create(DisconnectTwitterResponseSchema, {})));
      const CLEAR_JWT = { 'Set-Cookie': 'tb_jwt=; HttpOnly; SameSite=Lax; Path=/; Max-Age=0' };
      await apiMock.mockRPC('AuthService', 'Logout', {}, 200, CLEAR_JWT);

      const settings = new SettingsPage(page);
      await settings.goto('twitter');

      // 連携解除ボタンクリック → 確認ダイアログ
      await settings.disconnectButton.click();
      await expect(page.getByText('X連携を解除')).toBeVisible();

      // 確認ダイアログで「連携解除」をクリック
      await page.getByRole('button', { name: '連携解除' }).nth(1).click();
      await page.waitForURL('**/login');
    });
  });

  // ─── アカウントタブ ──────────────────────────────────────────

  test.describe('アカウントタブ', () => {
    test('アカウント削除 → 確認ダイアログ → /login へ', async ({ page, apiMock }) => {
      const CLEAR_JWT = { 'Set-Cookie': 'tb_jwt=; HttpOnly; SameSite=Lax; Path=/; Max-Age=0' };
      await apiMock.mockRPC('AuthService', 'DeleteAccount', {}, 200, CLEAR_JWT);

      const settings = new SettingsPage(page);
      await settings.goto('account');

      // 削除ボタンクリック → 確認ダイアログ
      await settings.deleteAccountButton.click();
      await expect(page.getByText('この操作は取り消せません。すべてのデータが完全に削除されます。')).toBeVisible();

      // 「削除する」をクリック
      await page.getByRole('button', { name: '削除する' }).click();
      await page.waitForURL('**/login');
    });
  });

  // ─── ダイアログ操作 ─────────────────────────────────────────

  test.describe('ダイアログ操作', () => {
    test('アカウント削除ダイアログ → Escape → 閉じる', async ({ page }) => {
      const settings = new SettingsPage(page);
      await settings.goto('account');

      // 削除ダイアログを開く
      await settings.deleteAccountButton.click();
      await expect(page.getByText('この操作は取り消せません。すべてのデータが完全に削除されます。')).toBeVisible();

      // Escape で閉じる
      await page.keyboard.press('Escape');

      // ダイアログが閉じてページが維持される
      await expect(page.getByText('この操作は取り消せません。すべてのデータが完全に削除されます。')).not.toBeVisible();
      await expect(settings.logoutButton).toBeVisible();
    });
  });
});
