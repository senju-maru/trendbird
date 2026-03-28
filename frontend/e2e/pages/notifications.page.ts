import { Page, Locator } from '@playwright/test';

export class NotificationsPage {
  constructor(private page: Page) {}

  // ── Navigation ──

  async goto() {
    await this.page.goto('/notifications');
    await this.page.waitForLoadState('networkidle');
  }

  // ── Header ──

  get heading(): Locator {
    return this.page.getByRole('heading', { name: '通知' });
  }

  get markAllReadButton(): Locator {
    return this.page.getByRole('button', { name: 'すべて既読にする' });
  }

  // ── Tabs ──

  tab(label: string): Locator {
    return this.page.getByText(label, { exact: true });
  }

  get allTab(): Locator {
    return this.tab('すべて');
  }

  get trendTab(): Locator {
    return this.tab('トレンド通知');
  }

  get systemTab(): Locator {
    return this.tab('運営');
  }

  // ── Notification cards ──

  notificationByTitle(title: string): Locator {
    return this.page.getByText(title, { exact: true });
  }

  // ── States ──

  get emptyState(): Locator {
    return this.page.getByText('通知はありません');
  }

  get errorState(): Locator {
    return this.page.getByText('通知の取得に失敗しました');
  }

  get retryButton(): Locator {
    return this.page.getByRole('button', { name: '再試行' });
  }

  get spinner(): Locator {
    return this.page.locator('[role="status"]');
  }
}
