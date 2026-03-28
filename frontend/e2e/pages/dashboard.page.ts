import { Page, Locator } from '@playwright/test';

export class DashboardPage {
  constructor(private page: Page) {}

  // ── Navigation ──

  async goto() {
    await this.page.goto('/dashboard');
    await this.page.waitForLoadState('networkidle');
  }

  async gotoDetail(topicName: string) {
    await this.goto();
    await this.page.getByText(topicName).first().click();
  }

  // ── Status bar ──

  get statusBar(): Locator {
    return this.page.locator('[data-tutorial="status-bar"]');
  }

  get spikeCountText(): Locator {
    return this.page.getByText(/\d+件 盛り上がり中/);
  }

  get risingCountText(): Locator {
    return this.page.getByText(/\d+件 上昇中/);
  }

  get stableCountText(): Locator {
    return this.page.getByText(/\d+件 安定/);
  }

  // ── Genre tabs ──

  get genreTabs(): Locator {
    return this.page.locator('[data-tutorial="genre-tabs"]');
  }

  genreTab(label: string): Locator {
    return this.genreTabs.getByRole('button', { name: label });
  }

  // ── Topic cards ──

  topicCardByName(name: string): Locator {
    return this.page.getByText(name, { exact: true }).first();
  }

  // ── Loading / Error ──

  get spinner(): Locator {
    return this.page.locator('[role="status"]');
  }

  // ── Detail page elements ──

  get backButton(): Locator {
    return this.page.getByRole('button', { name: 'ダッシュボード' });
  }

  get topicHeading(): Locator {
    return this.page.getByRole('heading');
  }

  get detailHero(): Locator {
    return this.page.locator('[data-tutorial="detail-hero"]');
  }

  get contextCard(): Locator {
    return this.page.locator('[data-tutorial="detail-context"]');
  }

  get aiSection(): Locator {
    return this.page.locator('[data-tutorial="detail-ai"]');
  }

  get aiGenerateButton(): Locator {
    return this.page.getByRole('button', { name: 'AI投稿文を生成する' });
  }

  get topicNotFound(): Locator {
    return this.page.getByText('トピックが見つかりません');
  }

  get backToDashboard(): Locator {
    return this.page.getByRole('button', { name: 'ダッシュボードに戻る' });
  }

  get calmText(): Locator {
    return this.page.getByText('現在は落ち着いています');
  }

  get generatingText(): Locator {
    return this.page.getByText('生成しています…');
  }

  get downloadError(): Locator {
    return this.page.getByText('ダウンロードに失敗しました');
  }
}
