import { Page, expect } from '@playwright/test';

/** toast 通知の表示検証 */
export async function expectToast(page: Page, text: string) {
  const toast = page.getByText(text);
  await expect(toast).toBeVisible({ timeout: 5_000 });
}

/** エラーメッセージの表示検証 */
export async function expectErrorMessage(page: Page, text: string) {
  const error = page.getByText(text);
  await expect(error).toBeVisible({ timeout: 5_000 });
}
