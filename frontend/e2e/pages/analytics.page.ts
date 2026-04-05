import { Page } from '@playwright/test';

export class AnalyticsPage {
  constructor(private page: Page) {}

  async goto() {
    await this.page.goto('/analytics');
    await this.page.waitForLoadState('networkidle');
    await this.page.getByRole('heading', { name: '分析' }).waitFor({ timeout: 10_000 });
  }

  get emptyState() {
    return this.page.getByText('アナリティクスデータがありません');
  }

  tabButton(name: string) {
    return this.page.getByRole('button', { name });
  }
}
