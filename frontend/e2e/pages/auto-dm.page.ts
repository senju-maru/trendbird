import { Page, Locator } from '@playwright/test';

export class AutoDMPage {
  constructor(private page: Page) {}

  async goto() {
    // ダッシュボードで先にユーザー状態を確立（layout の getCurrentUser 完了待ち）
    await this.page.goto('/dashboard');
    await this.page.waitForLoadState('networkidle');
    // Next.js Link でクライアントサイド遷移（Zustand auth state を維持）
    await this.page.locator('a[href="/auto-dm"]').click();
    await this.page.getByRole('heading', { name: '自動DM' }).waitFor({ timeout: 10_000 });
  }

  get heading() {
    return this.page.getByRole('heading', { name: '自動DM' });
  }

  get settingsTab() {
    return this.page.getByRole('button', { name: '設定' });
  }

  get historyTab() {
    return this.page.getByRole('button', { name: /送信履歴/ });
  }

  get addRuleButton() {
    return this.page.getByRole('button', { name: /ルールを追加/ });
  }

  get emptyState() {
    return this.page.getByText('自動DMルールがありません');
  }

  /** ルールカード内のToggle（button要素）。data-testid で特定 */
  toggleButton(index: number): Locator {
    return this.page.locator(`[data-testid="auto-dm-toggle-${index}"]`);
  }

  saveButton(label: '作成する' | '保存する' = '保存する'): Locator {
    return this.page.getByRole('button', { name: label });
  }

  get keywordInput(): Locator {
    return this.page.getByPlaceholder('キーワードを入力してEnter');
  }

  get templateInput(): Locator {
    return this.page.getByPlaceholder('自動送信するDMのテンプレートを入力');
  }

  get deleteButton(): Locator {
    return this.page.getByTitle('削除');
  }
}
