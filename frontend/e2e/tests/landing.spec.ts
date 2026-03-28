import { test, expect } from '../fixtures/test-base';

test.describe('ランディングページ', () => {
  test('Hero セクションと CTA ボタンが表示される', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Hero テキスト
    await expect(page.getByText('トレンド、まだ手動で追ってませんか？')).toBeVisible();
    await expect(page.getByText('AIが検知から投稿文の作成まで自動で')).toBeVisible();

    // CTA ボタン
    await expect(page.getByRole('link', { name: '無料で始める' }).first()).toBeVisible();
    await expect(page.getByText('パスワード共有なし').first()).toBeVisible();
  });

  test('CTA ボタンが X 認証 URL を指している', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    const href = await page.getByRole('link', { name: '無料で始める' }).first().getAttribute('href');
    expect(href).toContain('/auth/x');
  });

  test('FAQ アコーディオンが展開/折りたたみできる', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // FAQ セクション
    await expect(page.getByRole('heading', { name: 'Q&A' })).toBeVisible();

    // 最初のFAQ質問
    const question = page.getByText('TrendBirdで何ができますか？');
    await expect(question).toBeVisible();

    // 回答は最初は非表示
    const answer = page.getByText('トレンド監視、AI投稿文生成、予約投稿など、すべての機能を無料で利用できます。');
    await expect(answer).not.toBeVisible();

    // クリックで展開
    await question.click();
    await expect(answer).toBeVisible();

    // 再クリックで折りたたみ
    await question.click();
    await expect(answer).not.toBeVisible();
  });

  test('FAQ複数項目が同時に展開できる', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    const q1 = page.getByText('TrendBirdで何ができますか？');
    const q2 = page.getByText('AI投稿文はどのように生成されますか？');

    // Q1 展開
    await q1.click();
    const a1 = page.getByText('トレンド監視、AI投稿文生成、予約投稿など、すべての機能を無料で利用できます。');
    await expect(a1).toBeVisible();

    // Q2 展開（Q1 はまだ開いている）
    await q2.click();
    const a2 = page.getByText('トレンド検知時に、トピックの文脈と最新の関連ポストを分析し');
    await expect(a2).toBeVisible();

    // Q1 の回答がまだ表示されていることを確認
    await expect(a1).toBeVisible();
  });

  test('Featuresセクションが表示される', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Features セクションまでスクロール（viewport アニメーション発火）
    await page.evaluate(() => window.scrollTo(0, 800));
    await page.waitForTimeout(500);

    await expect(page.getByText('トレンド検知').first()).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText('AI投稿文生成').first()).toBeVisible();
  });

  test('HowItWorksセクションが表示される', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // HowItWorks セクションまでスクロール
    await page.evaluate(() => window.scrollTo(0, 1600));
    await page.waitForTimeout(500);

    await expect(page.getByText('ジャンルとトピックを選ぶ')).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText('トレンドを自動検知').first()).toBeVisible();
  });

  test('FinalCtaセクションが表示される', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // ページ最下部までスクロール
    await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));
    await page.waitForTimeout(500);

    await expect(page.getByText('今すぐ始めましょう')).toBeVisible({ timeout: 10_000 });
  });

  test('フッターにコピーライトが表示される', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    await expect(page.getByText('TrendBird. All rights reserved.')).toBeVisible();
  });
});
