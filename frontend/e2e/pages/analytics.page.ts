import { Page } from '@playwright/test';

export class AnalyticsPage {
  constructor(private page: Page) {}

  async goto() {
    // ダッシュボードで先にユーザー状態を確立（layout の getCurrentUser 完了待ち）
    await this.page.goto('/dashboard');
    await this.page.waitForLoadState('networkidle');
    // Next.js Link でクライアントサイド遷移（Zustand auth state を維持）
    await this.page.locator('a[href="/analytics"]').click();
    await this.page.getByRole('heading', { name: '分析' }).waitFor({ timeout: 10_000 });
  }

  get emptyState() {
    return this.page.getByText('アナリティクスデータがありません');
  }

  tabButton(name: string) {
    return this.page.getByRole('button', { name });
  }
}
