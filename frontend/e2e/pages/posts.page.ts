import { Page, Locator } from '@playwright/test';

export class PostsPage {
  constructor(private page: Page) {}

  // ── Navigation ──

  async goto() {
    await this.page.goto('/posts');
    await this.page.waitForLoadState('networkidle');
  }

  // ── Left pane: Topic selection ──

  topicCard(name: string): Locator {
    return this.page.getByText(name, { exact: true });
  }

  statusFilterButton(label: string): Locator {
    return this.page.getByRole('button', { name: label, exact: true });
  }

  genreFilterButton(label: string): Locator {
    return this.page.getByRole('button', { name: label });
  }

  get noTopicsMessage(): Locator {
    return this.page.getByText('該当するトピックがありません');
  }

  // ── Right pane: AI generation ──

  get generateButton(): Locator {
    return this.page.getByRole('button', { name: 'AI生成' });
  }

  get generatingButton(): Locator {
    return this.page.getByRole('button', { name: 'AI生成中...' });
  }

  get upgradeButton(): Locator {
    return this.page.getByRole('button', { name: 'アップグレード' });
  }

  get aiResultsHeader(): Locator {
    return this.page.getByText('生成結果');
  }

  get remainingGenerations(): Locator {
    return this.page.getByText(/残り生成回数/);
  }

  // ── Composer ──

  get composerTextarea(): Locator {
    return this.page.locator('textarea');
  }

  get charCounter(): Locator {
    return this.page.getByText(/\/280/);
  }

  get saveDraftButton(): Locator {
    return this.page.getByRole('button', { name: '下書き保存' });
  }

  get scheduleButton(): Locator {
    return this.page.getByRole('button', { name: '予約', exact: true });
  }

  get publishButton(): Locator {
    return this.page.getByRole('button', { name: '今すぐ投稿' });
  }

  get xDisconnectedLabel(): Locator {
    return this.page.getByText('X未連携');
  }

  // ── Management section ──

  get managementHeader(): Locator {
    return this.page.getByText('投稿管理');
  }

  mgmtTab(label: string): Locator {
    // タブ名にはカウントバッジが含まれる（例: "下書き 1"）ため部分一致
    return this.page.getByRole('button', { name: new RegExp(`^${label}\\s`) });
  }

  get draftsTab(): Locator {
    return this.mgmtTab('下書き');
  }

  get scheduledTab(): Locator {
    return this.mgmtTab('予約');
  }

  get historyTab(): Locator {
    return this.mgmtTab('履歴');
  }

  // ── Empty states ──

  get noDrafts(): Locator {
    return this.page.getByText('下書きはありません');
  }

  get noScheduled(): Locator {
    return this.page.getByText('予約中の投稿はありません');
  }

  get noHistory(): Locator {
    return this.page.getByText('投稿履歴はありません');
  }

  // ── No topic selected ──

  get selectTopicPrompt(): Locator {
    return this.page.getByText('トピックを選択してAI投稿文を生成');
  }

  // ── Modals ──

  get editModal(): Locator {
    return this.page.getByText('投稿を編集');
  }

  get scheduleModal(): Locator {
    return this.page.getByText('投稿を予約');
  }

  get deleteDialog(): Locator {
    return this.page.getByRole('heading', { name: '下書きを削除' });
  }

  get publishDialog(): Locator {
    return this.page.getByText('投稿の確認');
  }
}
