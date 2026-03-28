import { test, expect, authenticateContext } from '../fixtures/test-base';
import { resetSeq } from '../fixtures/factories';

test.describe('ルートガード', () => {
  test.beforeEach(async ({ page }) => {
    resetSeq();
    await page.addInitScript(() => {
      sessionStorage.removeItem('tb_tutorial_pending');
    });
  });

  test('認証済み + /login → /dashboard リダイレクト', async ({
    context,
    page,
    apiMock,
  }) => {
    await authenticateContext(context);
    await apiMock.setupDefaults();

    await page.goto('/login');
    await page.waitForURL('**/dashboard');
  });

  test('未認証 + /dashboard → / リダイレクト', async ({ page }) => {
    await page.goto('/dashboard');
    await page.waitForURL(/^http:\/\/localhost:\d+\/$/);
  });

  test('認証済み + /dashboard → アクセス可', async ({
    context,
    page,
    apiMock,
  }) => {
    await authenticateContext(context);
    await apiMock.setupDefaults();

    await page.goto('/dashboard');
    await expect(page).toHaveURL(/\/dashboard/);
  });

  test('未認証 + /topics → / リダイレクト', async ({ page }) => {
    await page.goto('/topics');
    await page.waitForURL(/^http:\/\/localhost:\d+\/$/);
  });

  test('未認証 + /posts → / リダイレクト', async ({ page }) => {
    await page.goto('/posts');
    await page.waitForURL(/^http:\/\/localhost:\d+\/$/);
  });

  test('未認証 + /notifications → / リダイレクト', async ({ page }) => {
    await page.goto('/notifications');
    await page.waitForURL(/^http:\/\/localhost:\d+\/$/);
  });
});
