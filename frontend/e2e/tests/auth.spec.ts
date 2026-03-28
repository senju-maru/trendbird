import { test, expect, authenticateContext } from '../fixtures/test-base';
import { buildUser, resetSeq } from '../fixtures/factories';
import { LoginPage } from '../pages/login.page';

test.describe('ログインページ', () => {
  test.beforeEach(async ({ page }) => {
    resetSeq();
    await page.addInitScript(() => {
      sessionStorage.removeItem('tb_tutorial_pending');
    });
  });

  test('ロゴ・ログインボタン・ガイダンスが表示される', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();

    await expect(loginPage.heading).toBeVisible();
    await expect(loginPage.xLoginButton).toBeVisible();
    await expect(loginPage.guidanceText).toBeVisible();
  });

  test('X ログインボタン押下で /auth/x に遷移', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await loginPage.goto();

    await page.route('**/auth/x**', (route) => route.abort());

    const requestPromise = page.waitForRequest((req) =>
      req.url().includes('/auth/x'),
    );
    await loginPage.xLoginButton.click();
    const request = await requestPromise;
    expect(request.url()).toContain('/auth/x');
  });

  test('loggedOut=true で「ログアウトしました」表示', async ({ page }) => {
    const loginPage = new LoginPage(page);
    await page.goto('/login?loggedOut=true');

    await expect(loginPage.loggedOutMessage).toBeVisible();
  });
});

test.describe('OAuth コールバック', () => {
  test.beforeEach(async ({ page }) => {
    resetSeq();
    await page.addInitScript(() => {
      sessionStorage.removeItem('tb_tutorial_pending');
    });
  });

  test('code あり + XAuth 成功 → /dashboard リダイレクト', async ({
    context,
    page,
    apiMock,
  }) => {
    // 実フローではバックエンドがCookieを設定済み。テストではauthenticateContextで再現
    await authenticateContext(context);
    await apiMock.mockRPC('AuthService', 'XAuth', {
      user: buildUser(),

      tutorialPending: false,
    });
    await apiMock.setupDefaults();

    await page.goto('/callback?code=test-code');
    await page.waitForURL('**/dashboard');
  });

  test('returnTo あり → returnTo 先へリダイレクト', async ({
    context,
    page,
    apiMock,
  }) => {
    await authenticateContext(context);
    await apiMock.mockRPC('AuthService', 'XAuth', {
      user: buildUser(),

      tutorialPending: false,
    });
    await apiMock.setupDefaults();
    await page.addInitScript(() => {
      localStorage.setItem('tb_return_to', '/settings');
    });

    await page.goto('/callback?code=test-code');
    await page.waitForURL('**/settings');
  });

  test('error=access_denied → /login リダイレクト', async ({ page }) => {
    await page.goto('/callback?error=access_denied');
    await page.waitForURL('**/login');
  });

  test('code なし → エラーメッセージ表示', async ({ page }) => {
    await page.goto('/callback');
    await expect(
      page.getByText('認証コードが取得できませんでした'),
    ).toBeVisible();
  });

  test('XAuth RPC 失敗 → エラー + ログインリンク', async ({
    page,
    apiMock,
  }) => {
    await apiMock.mockRPCError(
      'AuthService',
      'XAuth',
      'internal',
      'サーバーエラー',
    );

    await page.goto('/callback?code=test-code');
    await expect(page.getByText('認証に失敗しました')).toBeVisible();
    await expect(
      page.getByRole('link', { name: 'ログインページに戻る' }),
    ).toBeVisible();
  });
});

test.describe('セッション管理', () => {
  test.beforeEach(async ({ page }) => {
    resetSeq();
    await page.addInitScript(() => {
      sessionStorage.removeItem('tb_tutorial_pending');
    });
  });

  test('getCurrentUser 失敗 → セッションクリア → LP リダイレクト', async ({
    context,
    page,
    apiMock,
  }) => {
    // JWT Cookie あり → ミドルウェアは通過するが、バックエンドがJWTを拒否するケース
    await authenticateContext(context);
    await apiMock.mockRPCError(
      'AuthService',
      'GetCurrentUser',
      'internal',
      'session invalid',
    );

    await page.goto('/dashboard');
    // layout.tsx catch → /api/clear-session → Cookie削除 → / へリダイレクト
    await page.waitForURL(/^http:\/\/localhost:\d+\/$/);
  });

  test('getCurrentUser が Unauthenticated → セッションクリア → LP リダイレクト', async ({
    context,
    page,
    apiMock,
  }) => {
    // JWT Cookie あり → ミドルウェアは通過するが、バックエンドが Unauthenticated を返すケース
    // A-1（internal エラー）とは異なり、authRedirectInterceptor が発火するが、
    // layout.tsx の catch → /api/clear-session → Cookie 削除 → / が最終的に勝つ
    await authenticateContext(context);
    await apiMock.mockRPCError(
      'AuthService',
      'GetCurrentUser',
      'unauthenticated',
      'token expired',
    );

    await page.goto('/dashboard');
    await page.waitForURL(/^http:\/\/localhost:\d+\/$/);
  });

  test('ヘッダーからログアウト → LP リダイレクト', async ({
    context,
    page,
    apiMock,
  }) => {
    await authenticateContext(context);
    await apiMock.setupDefaults();
    await apiMock.mockRPC('AuthService', 'Logout', {});

    await page.goto('/dashboard');
    // ダッシュボードが表示されるまで待機
    await expect(page.getByLabel('Default User')).toBeVisible();

    // アバターをクリックしてメニューを開く
    await page.getByLabel('Default User').click();
    // ヘッダー内のログアウトボタンをクリック（サイドバーにも同名ボタンがあるため限定）
    await page
      .getByRole('banner')
      .getByRole('button', { name: 'ログアウト' })
      .click();

    // router.push('/') → LP へ遷移
    await page.waitForURL(/^http:\/\/localhost:\d+\/$/);
  });
});
