import { Page, Locator } from '@playwright/test';

export class TopicsPage {
  constructor(private page: Page) {}

  // ── Navigation ──

  async goto() {
    await this.page.goto('/topics');
    await this.page.waitForLoadState('networkidle');
  }

  // ── Genre ──

  get addGenreButton(): Locator {
    return this.page.getByRole('button', { name: 'ジャンルを選ぶ' });
  }

  genreChip(label: string): Locator {
    return this.page.getByText(label, { exact: true });
  }

  // ── Genre select modal ──

  get genreModal(): Locator {
    return this.page.getByText('ジャンルを追加');
  }

  genreOption(label: string): Locator {
    return this.page.getByRole('button', { name: label });
  }

  // ── Tabs ──

  get browseByGenreTab(): Locator {
    return this.page.getByText('ジャンルから探す', { exact: true });
  }

  get searchTopicsTab(): Locator {
    return this.page.getByText('トピックから探す', { exact: true });
  }

  // ── Topic search ──

  get topicSearchInput(): Locator {
    return this.page.getByPlaceholder('トピック名で検索...');
  }

  // ── Suggestions ──

  get recommendedSection(): Locator {
    return this.page.getByText('おすすめトピック');
  }

  get selectedTopicsSection(): Locator {
    return this.page.getByText('選択中のトピック');
  }

  // ── Progress ──

  genreProgress(text: string): Locator {
    return this.page.getByText(text);
  }

  topicProgress(text: string): Locator {
    return this.page.getByText(text);
  }

  // ── Empty states ──

  get emptyGenreState(): Locator {
    return this.page.getByText('監視したいトピックを設定しましょう');
  }

  get noSuggestions(): Locator {
    return this.page.getByText('このジャンルにはまだトピックがありません');
  }

  get noSearchResults(): Locator {
    return this.page.getByText('一致するトピックがありません');
  }
}
