import { test, expect, authenticateContext } from '../fixtures/test-base';
import { resetSeq, TopicStatus, buildTopic, buildGenre } from '../fixtures/factories';
import { DashboardPage } from '../pages/dashboard.page';

test.describe('ダッシュボード一覧', () => {
  test.beforeEach(async ({ context, page, apiMock }) => {
    resetSeq();
    await authenticateContext(context);
    await apiMock.setupDefaults();
    await page.addInitScript(() => {
      sessionStorage.removeItem('tb_tutorial_pending');
    });
  });

  test.describe('トピック表示', () => {
    test('トピックカードが表示される', async ({ page }) => {
      const dashboard = new DashboardPage(page);
      await dashboard.goto();

      await expect(page.getByText('Default Topic 1')).toBeVisible();
      await expect(page.getByText('Default Topic 2')).toBeVisible();
      // spike トピックの z-score
      await expect(page.getByText('3.5')).toBeVisible();
    });

    test('spike カードが stable より先に表示される', async ({ page }) => {
      const dashboard = new DashboardPage(page);
      await dashboard.goto();

      // 最初のカード（data-tutorial="topic-card"）は spike トピック
      const firstCard = page.locator('[data-tutorial="topic-card"]');
      await expect(firstCard).toBeVisible();
      await expect(firstCard.getByText('Default Topic 1')).toBeVisible();
    });

    test('倍率が1を超えると「ふだんのN倍」が表示される', async ({ page, apiMock }) => {
      // currentVolume=500, baselineVolume=100 → mult=5
      const topic = buildTopic({
        id: 'mult-topic',
        name: 'Mult Topic',
        genre: 'technology',
        status: TopicStatus.SPIKE,
        zScore: 4.0,
        currentVolume: 500,
        baselineVolume: 100,
      });
      await apiMock.clearMock('TopicService', 'ListTopics');
      await apiMock.mockRPC('TopicService', 'ListTopics', { topics: [topic] });

      const dashboard = new DashboardPage(page);
      await dashboard.goto();

      await expect(page.getByText('ふだんの').first()).toBeVisible();
      await expect(page.getByText('5倍').first()).toBeVisible();
    });

    test('倍率が1以下のとき「前回比」が表示される', async ({ page, apiMock }) => {
      // currentVolume=90, baselineVolume=100 → mult=1（1 > 1 は false）
      const topic = buildTopic({
        id: 'change-topic',
        name: 'Change Topic',
        genre: 'technology',
        status: TopicStatus.STABLE,
        zScore: 1.0,
        currentVolume: 90,
        baselineVolume: 100,
        changePercent: -10,
      });
      await apiMock.clearMock('TopicService', 'ListTopics');
      await apiMock.mockRPC('TopicService', 'ListTopics', { topics: [topic] });

      const dashboard = new DashboardPage(page);
      await dashboard.goto();

      await expect(page.getByText('前回比').first()).toBeVisible();
    });
  });

  test.describe('ステータスバー', () => {
    test('ステータスバーに件数が表示される', async ({ page }) => {
      const dashboard = new DashboardPage(page);
      await dashboard.goto();

      await expect(dashboard.spikeCountText).toBeVisible();
      await expect(dashboard.stableCountText).toBeVisible();
    });

    test('Rising 0件のとき「上昇中」が非表示', async ({ page }) => {
      // デフォルトのトピックは SPIKE と STABLE のみ（RISING なし）
      const dashboard = new DashboardPage(page);
      await dashboard.goto();

      await expect(dashboard.spikeCountText).toBeVisible();
      await expect(dashboard.stableCountText).toBeVisible();
      await expect(dashboard.risingCountText).not.toBeVisible();
    });

    test('最終チェック時刻が表示される', async ({ page, apiMock }) => {
      // GetStats に lastCheckedAt をセット
      const { create, toJson } = await import('@bufbuild/protobuf');
      const { DashboardStatsSchema, GetStatsResponseSchema } = await import('../../src/gen/trendbird/v1/dashboard_pb');
      await apiMock.clearMock('DashboardService', 'GetStats');
      await apiMock.mockRPC('DashboardService', 'GetStats',
        toJson(GetStatsResponseSchema,
          create(GetStatsResponseSchema, {
            stats: create(DashboardStatsSchema, {
              lastCheckedAt: new Date().toISOString(),
            }),
          })));

      const dashboard = new DashboardPage(page);
      await dashboard.goto();

      await expect(page.getByText('最終チェック:')).toBeVisible();
    });
  });

  test.describe('ジャンルフィルタ', () => {
    test('ジャンルタブでフィルタリングできる', async ({ page, apiMock }) => {
      // 異なるジャンルのトピックを追加
      const techTopic = buildTopic({ id: 'tech-1', name: 'Tech Topic', genre: 'technology', status: TopicStatus.SPIKE, zScore: 3.0 });
      const bizTopic = buildTopic({ id: 'biz-1', name: 'Biz Topic', genre: 'business', status: TopicStatus.STABLE, zScore: 1.0 });

      const { create, toJson } = await import('@bufbuild/protobuf');
      const { TopicSchema, ListTopicsResponseSchema } = await import('../../src/gen/trendbird/v1/topic_pb');
      await apiMock.clearMock('TopicService', 'ListTopics');
      await apiMock.mockRPC('TopicService', 'ListTopics', { topics: [techTopic, bizTopic] });

      const dashboard = new DashboardPage(page);
      await dashboard.goto();

      // 「全て」タブでは両方表示
      await expect(page.getByText('Tech Topic')).toBeVisible();
      await expect(page.getByText('Biz Topic')).toBeVisible();

      // Business タブでフィルタ
      await dashboard.genreTab('Business').click();
      await expect(page.getByText('Biz Topic')).toBeVisible();
      await expect(page.getByText('Tech Topic')).not.toBeVisible();

      // Technology タブに戻す
      await dashboard.genreTab('Technology').click();
      await expect(page.getByText('Tech Topic')).toBeVisible();
      await expect(page.getByText('Biz Topic')).not.toBeVisible();
    });
  });

  test.describe('ナビゲーション', () => {
    test('カードクリックで詳細ページに遷移する', async ({ page }) => {
      const dashboard = new DashboardPage(page);
      await dashboard.goto();

      await page.getByText('Default Topic 1').click();
      await page.waitForURL('**/dashboard/default-topic-1');
    });
  });

  test.describe('エッジケース', () => {
    test('トピック0件で空状態CTAが表示される', async ({ page, apiMock }) => {
      await apiMock.clearMock('TopicService', 'ListTopics');
      await apiMock.mockRPC('TopicService', 'ListTopics', { topics: [] });

      const dashboard = new DashboardPage(page);
      await dashboard.goto();

      await expect(page.getByText('Default Topic 1')).not.toBeVisible();
      await expect(page.getByText('Default Topic 2')).not.toBeVisible();
      await expect(page.getByText('まずトピックを追加しましょう')).toBeVisible();
      await expect(page.getByRole('button', { name: 'トピックを設定する' })).toBeVisible();
    });

    test('API エラー時にエラーメッセージが表示される', async ({ page, apiMock }) => {
      await apiMock.clearMock('TopicService', 'ListTopics');
      await apiMock.mockRPCError('TopicService', 'ListTopics', 'internal', 'サーバーエラー');

      const dashboard = new DashboardPage(page);
      await dashboard.goto();

      // エラーメッセージが表示されること
      await expect(page.getByText('ただいま処理がうまくいきませんでした')).toBeVisible({ timeout: 10_000 });
    });
  });
});
