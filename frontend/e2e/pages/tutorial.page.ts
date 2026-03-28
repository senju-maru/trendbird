import { type Page, type Locator, expect } from '@playwright/test';

export class TutorialPage {
  constructor(private page: Page) {}

  // ── driver.js ポップオーバー ──
  get popover(): Locator { return this.page.locator('.driver-popover'); }
  get popoverTitle(): Locator { return this.page.locator('.driver-popover-title'); }
  get popoverDescription(): Locator { return this.page.locator('.driver-popover-description'); }
  get nextButton(): Locator { return this.page.locator('.driver-popover-next-btn'); }
  get doneButton(): Locator { return this.page.locator('.driver-popover-done-btn'); }

  // ── data-tutorial 要素 ──
  get genreSelectCta(): Locator { return this.page.locator('[data-tutorial="genre-select-cta"]'); }
  get tutorialGenreRow(): Locator { return this.page.locator('[data-tutorial="tutorial-genre-row"]'); }
  get tutorialSuggestRow(): Locator { return this.page.locator('[data-tutorial="tutorial-suggest-row"]'); }
  get statusBar(): Locator { return this.page.locator('[data-tutorial="status-bar"]'); }
  get topicCard(): Locator { return this.page.locator('[data-tutorial="topic-card"]'); }
  get detailHero(): Locator { return this.page.locator('[data-tutorial="detail-hero"]'); }
  get detailContext(): Locator { return this.page.locator('[data-tutorial="detail-context"]'); }
  get detailAi(): Locator { return this.page.locator('[data-tutorial="detail-ai"]'); }
  get sidebarDashboard(): Locator { return this.page.locator('[data-tutorial="sidebar-dashboard"]'); }
  get sidebarTopics(): Locator { return this.page.locator('[data-tutorial="sidebar-topics"]'); }

  // ── ヘルパー ──
  async waitForPopover(title?: string) {
    if (title) {
      // Wait for popover with matching title (avoids stale popover from previous phase)
      const popoverWithTitle = this.page.locator('.driver-popover').filter({
        has: this.page.locator('.driver-popover-title', { hasText: title }),
      });
      await popoverWithTitle.waitFor({ state: 'visible', timeout: 15_000 });
    } else {
      await this.popover.waitFor({ state: 'visible', timeout: 15_000 });
    }
  }

  async clickNext() {
    const btn = this.page.locator('.driver-popover-next-btn, .driver-popover-done-btn').first();
    await btn.waitFor({ state: 'visible' });
    await btn.click();
  }

  async waitForPopoverHidden() {
    await this.popover.waitFor({ state: 'hidden', timeout: 10_000 });
  }
}
