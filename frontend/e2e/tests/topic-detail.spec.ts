import { test, expect, authenticateContext } from '../fixtures/test-base';
import {
  resetSeq,
  TopicStatus,
  buildTopic,
  buildGeneratedPost,
  buildSpikeHistoryEntry,
  buildSparklineDataPoint,
  buildPostingTips,
} from '../fixtures/factories';
import { DashboardPage } from '../pages/dashboard.page';

test.describe('トピック詳細', () => {
  test.beforeEach(async ({ context, page, apiMock }) => {
    resetSeq();
    await authenticateContext(context);
    await apiMock.setupDefaults();
    await page.addInitScript(() => {
      sessionStorage.removeItem('tb_tutorial_pending');
    });
  });

  test.describe('Active状態（spike）', () => {
    test('spike トピックの Hero が表示される', async ({ page }) => {
      const dashboard = new DashboardPage(page);
      await dashboard.gotoDetail('Default Topic 1');
      await page.waitForURL('**/dashboard/default-topic-1');

      // トピック名
      await expect(dashboard.topicHeading.getByText('Default Topic 1')).toBeVisible();
      // detail-hero セクション
      await expect(dashboard.detailHero).toBeVisible();
      // z-score 表示
      await expect(page.getByText('3.5')).toBeVisible();
      // ステータスバッジ
      await expect(page.getByText('盛り上がり中')).toBeVisible();
    });

    test('contextSummary があるとコンテキストカードが表示される', async ({ page }) => {
      const dashboard = new DashboardPage(page);
      await dashboard.gotoDetail('Default Topic 1');
      await page.waitForURL('**/dashboard/default-topic-1');

      // detail-context セクション
      await expect(dashboard.contextCard).toBeVisible();
      await expect(page.getByText('最新のAI技術が注目されています')).toBeVisible();
    });

    test('AI 投稿文生成（Lite）: 生成 → 結果表示', async ({ page, apiMock }) => {
      const genPost = buildGeneratedPost({ content: 'AIで生成された投稿文です', topicId: 'default-topic-1' });
      await apiMock.mockRPC('PostService', 'GeneratePosts', { posts: [genPost] });

      const dashboard = new DashboardPage(page);
      await dashboard.gotoDetail('Default Topic 1');
      await page.waitForURL('**/dashboard/default-topic-1');

      // AI セクションが表示される
      await expect(dashboard.aiSection).toBeVisible();
      // 生成ボタンをクリック
      await dashboard.aiGenerateButton.click();
      // 生成結果が表示される
      await expect(page.getByText('AIで生成された投稿文です')).toBeVisible();
    });

    test('contextSummary がないとコンテキストカードが非表示', async ({ page, apiMock }) => {
      const { create, toJson } = await import('@bufbuild/protobuf');
      const { TopicSchema, ListTopicsResponseSchema } = await import('../../src/gen/trendbird/v1/topic_pb');

      await apiMock.clearMock('TopicService', 'ListTopics');
      await apiMock.mockRPC('TopicService', 'ListTopics',
        toJson(ListTopicsResponseSchema,
          create(ListTopicsResponseSchema, {
            topics: [
              create(TopicSchema, {
                id: 'no-context-topic',
                name: 'No Context Topic',
                keywords: ['keyword'],
                genre: 'technology',
                status: TopicStatus.SPIKE,
                zScore: 3.5,
                currentVolume: 500,
                baselineVolume: 100,
                sparklineData: [],
                weeklySparklineData: [],
                spikeHistory: [],
                // contextSummary を意図的に省略
                spikeStartedAt: '2026-01-01T00:00:00Z',
                notificationEnabled: true,
                createdAt: '2026-01-01T00:00:00Z',
              }),
            ],
          })));

      const dashboard = new DashboardPage(page);
      await dashboard.goto();
      await page.getByText('No Context Topic').click();
      await page.waitForURL('**/dashboard/no-context-topic');

      // spike トピックなので Hero は表示される
      await expect(dashboard.detailHero).toBeVisible();
      // contextSummary がないのでコンテキストカードは非表示
      await expect(dashboard.contextCard).not.toBeVisible();
    });

    test('Free プランで AI 生成ボタンが表示・有効', async ({ page, apiMock }) => {
      await apiMock.setupDefaults();

      const dashboard = new DashboardPage(page);
      await dashboard.gotoDetail('Default Topic 1');
      await page.waitForURL('**/dashboard/default-topic-1');

      // Free プランでも生成ボタンが表示・有効
      await expect(dashboard.aiGenerateButton).toBeVisible();
      await expect(dashboard.aiGenerateButton).toBeEnabled();
    });

    test('AI 投稿文生成エラー時にエラーメッセージ表示', async ({ page, apiMock }) => {
      await apiMock.mockRPCError('PostService', 'GeneratePosts', 'internal', '生成に失敗しました');

      const dashboard = new DashboardPage(page);
      await dashboard.gotoDetail('Default Topic 1');
      await page.waitForURL('**/dashboard/default-topic-1');

      await dashboard.aiGenerateButton.click();
      // エラーメッセージが表示される
      await expect(page.getByText('ただいま処理がうまくいきませんでした')).toBeVisible({ timeout: 10_000 });
    });

    test('AI 生成中のローディングフロー', async ({ page, apiMock }) => {
      // GeneratePosts を遅延させる
      await page.route('**/trendbird.v1.PostService/GeneratePosts', async (route) => {
        await new Promise((r) => setTimeout(r, 2000));
        const genPost = buildGeneratedPost({ content: '遅延後の生成結果', topicId: 'default-topic-1' });
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ posts: [genPost] }),
        });
      });

      const dashboard = new DashboardPage(page);
      await dashboard.gotoDetail('Default Topic 1');
      await page.waitForURL('**/dashboard/default-topic-1');

      await dashboard.aiGenerateButton.click();
      // ローディング表示
      await expect(dashboard.generatingText).toBeVisible();
      // 遅延後に結果表示
      await expect(page.getByText('遅延後の生成結果')).toBeVisible({ timeout: 10_000 });
    });
  });

  test.describe('Stable状態', () => {
    test('stable トピックは「現在は落ち着いています」表示', async ({ page }) => {
      const dashboard = new DashboardPage(page);
      await dashboard.gotoDetail('Default Topic 2');
      await page.waitForURL('**/dashboard/default-topic-2');

      await expect(dashboard.calmText).toBeVisible();
    });

    test('スパイク履歴がある場合に履歴が表示される', async ({ page, apiMock }) => {
      const spikeEntry1 = buildSpikeHistoryEntry({
        timestamp: '2026-01-05T10:00:00Z',
        peakZScore: 4.2,
        status: TopicStatus.SPIKE,
        summary: 'AI技術の大型発表',
      });
      const spikeEntry2 = buildSpikeHistoryEntry({
        timestamp: '2026-01-02T08:00:00Z',
        peakZScore: 3.1,
        status: TopicStatus.RISING,
        summary: '新しいライブラリ公開',
      });

      const stableTopic = buildTopic({
        id: 'default-topic-2',
        name: 'Default Topic 2',
        genre: 'technology',
        status: TopicStatus.STABLE,
        zScore: 1.2,
        currentVolume: 100,
        baselineVolume: 80,
        spikeHistory: [spikeEntry1, spikeEntry2],
      });

      await apiMock.clearMock('TopicService', 'ListTopics');
      const { create, toJson } = await import('@bufbuild/protobuf');
      const { TopicSchema, ListTopicsResponseSchema } = await import('../../src/gen/trendbird/v1/topic_pb');
      await apiMock.mockRPC('TopicService', 'ListTopics',
        toJson(ListTopicsResponseSchema,
          create(ListTopicsResponseSchema, {
            topics: [
              create(TopicSchema, {
                id: 'default-topic-1',
                name: 'Default Topic 1',
                keywords: ['keyword-1'],
                genre: 'technology',
                status: TopicStatus.SPIKE,
                zScore: 3.5,
                currentVolume: 500,
                baselineVolume: 100,
                sparklineData: [],
                weeklySparklineData: [],
                spikeHistory: [],
                context: '最新のAI技術が注目されています',
                contextSummary: '最新のAI技術が注目されています',
                spikeStartedAt: '2026-01-01T00:00:00Z',
                notificationEnabled: true,
                createdAt: '2026-01-01T00:00:00Z',
              }),
              create(TopicSchema, {
                id: 'default-topic-2',
                name: 'Default Topic 2',
                keywords: ['keyword-2'],
                genre: 'technology',
                status: TopicStatus.STABLE,
                zScore: 1.2,
                currentVolume: 100,
                baselineVolume: 80,
                sparklineData: [],
                weeklySparklineData: [],
                spikeHistory: [spikeEntry1, spikeEntry2],
                notificationEnabled: true,
                createdAt: '2026-01-01T00:00:00Z',
              }),
            ],
          })));

      const dashboard = new DashboardPage(page);
      await dashboard.gotoDetail('Default Topic 2');
      await page.waitForURL('**/dashboard/default-topic-2');

      // スパイク履歴セクション
      await expect(page.getByText('スパイク履歴')).toBeVisible();
      await expect(page.getByText('AI技術の大型発表')).toBeVisible();
      await expect(page.getByText('4.2')).toBeVisible();
    });

    test('週次スパークラインチャートが表示される', async ({ page, apiMock }) => {
      const sparklineData = Array.from({ length: 7 }, (_, i) =>
        buildSparklineDataPoint(
          new Date(2026, 0, 5 + i).toISOString(),
          50 + i * 10,
        ),
      );

      await apiMock.clearMock('TopicService', 'ListTopics');
      const { create, toJson } = await import('@bufbuild/protobuf');
      const { TopicSchema, ListTopicsResponseSchema } = await import('../../src/gen/trendbird/v1/topic_pb');
      await apiMock.mockRPC('TopicService', 'ListTopics',
        toJson(ListTopicsResponseSchema,
          create(ListTopicsResponseSchema, {
            topics: [
              create(TopicSchema, {
                id: 'default-topic-1',
                name: 'Default Topic 1',
                keywords: ['keyword-1'],
                genre: 'technology',
                status: TopicStatus.SPIKE,
                zScore: 3.5,
                currentVolume: 500,
                baselineVolume: 100,
                sparklineData: [],
                weeklySparklineData: [],
                spikeHistory: [],
                context: '最新のAI技術が注目されています',
                contextSummary: '最新のAI技術が注目されています',
                spikeStartedAt: '2026-01-01T00:00:00Z',
                notificationEnabled: true,
                createdAt: '2026-01-01T00:00:00Z',
              }),
              create(TopicSchema, {
                id: 'default-topic-2',
                name: 'Default Topic 2',
                keywords: ['keyword-2'],
                genre: 'technology',
                status: TopicStatus.STABLE,
                zScore: 1.2,
                currentVolume: 100,
                baselineVolume: 80,
                sparklineData: [],
                weeklySparklineData: sparklineData,
                spikeHistory: [],
                notificationEnabled: true,
                createdAt: '2026-01-01T00:00:00Z',
              }),
            ],
          })));

      const dashboard = new DashboardPage(page);
      await dashboard.gotoDetail('Default Topic 2');
      await page.waitForURL('**/dashboard/default-topic-2');

      // 週次チャートのヘッダー
      await expect(page.getByText('過去7日間の推移')).toBeVisible();
    });

    test('おすすめ投稿タイミングが表示される', async ({ page, apiMock }) => {
      const tips = buildPostingTips({
        peakDays: ['月', '水', '金'],
        peakHoursStart: 19,
        peakHoursEnd: 22,
        nextSuggestedTime: '2026-03-12T20:00:00Z',
      });

      await apiMock.clearMock('TopicService', 'ListTopics');
      const { create, toJson } = await import('@bufbuild/protobuf');
      const { TopicSchema, ListTopicsResponseSchema } = await import('../../src/gen/trendbird/v1/topic_pb');
      await apiMock.mockRPC('TopicService', 'ListTopics',
        toJson(ListTopicsResponseSchema,
          create(ListTopicsResponseSchema, {
            topics: [
              create(TopicSchema, {
                id: 'default-topic-1',
                name: 'Default Topic 1',
                keywords: ['keyword-1'],
                genre: 'technology',
                status: TopicStatus.SPIKE,
                zScore: 3.5,
                currentVolume: 500,
                baselineVolume: 100,
                sparklineData: [],
                weeklySparklineData: [],
                spikeHistory: [],
                context: '最新のAI技術が注目されています',
                contextSummary: '最新のAI技術が注目されています',
                spikeStartedAt: '2026-01-01T00:00:00Z',
                notificationEnabled: true,
                createdAt: '2026-01-01T00:00:00Z',
              }),
              create(TopicSchema, {
                id: 'default-topic-2',
                name: 'Default Topic 2',
                keywords: ['keyword-2'],
                genre: 'technology',
                status: TopicStatus.STABLE,
                zScore: 1.2,
                currentVolume: 100,
                baselineVolume: 80,
                sparklineData: [],
                weeklySparklineData: [],
                spikeHistory: [],
                postingTips: tips,
                notificationEnabled: true,
                createdAt: '2026-01-01T00:00:00Z',
              }),
            ],
          })));

      const dashboard = new DashboardPage(page);
      await dashboard.gotoDetail('Default Topic 2');
      await page.waitForURL('**/dashboard/default-topic-2');

      await expect(page.getByText('おすすめ投稿タイミング')).toBeVisible();
      await expect(page.getByText('月・水・金 19:00〜22:00 に盛り上がりやすい傾向')).toBeVisible();
    });

    test('スパイク履歴0件のとき「スパイク履歴」セクションが非表示', async ({ page }) => {
      // デフォルトの Default Topic 2 は spikeHistory が空
      const dashboard = new DashboardPage(page);
      await dashboard.gotoDetail('Default Topic 2');
      await page.waitForURL('**/dashboard/default-topic-2');

      await expect(page.getByText('スパイク履歴')).not.toBeVisible();
    });

    test('stable でも AI セクションは表示、コンテキストは非表示', async ({ page }) => {
      const dashboard = new DashboardPage(page);
      await dashboard.gotoDetail('Default Topic 2');
      await page.waitForURL('**/dashboard/default-topic-2');

      await expect(dashboard.aiSection).toBeVisible();
      await expect(dashboard.contextCard).not.toBeVisible();
    });

  });

  test.describe('盛り上がり通知セクション削除', () => {
    test('spike 詳細に盛り上がり通知セクションが表示されない', async ({ page }) => {
      const dashboard = new DashboardPage(page);
      await dashboard.gotoDetail('Default Topic 1');
      await page.waitForURL('**/dashboard/default-topic-1');

      await expect(dashboard.detailHero).toBeVisible();
      await expect(page.getByText('盛り上がり通知')).not.toBeVisible();
    });

    test('stable 詳細に盛り上がり通知セクションが表示されない', async ({ page }) => {
      const dashboard = new DashboardPage(page);
      await dashboard.gotoDetail('Default Topic 2');
      await page.waitForURL('**/dashboard/default-topic-2');

      await expect(dashboard.calmText).toBeVisible();
      await expect(page.getByText('盛り上がり通知')).not.toBeVisible();
    });
  });

  test.describe('共通・エッジケース', () => {
    test('戻るボタンで /dashboard に遷移する', async ({ page }) => {
      const dashboard = new DashboardPage(page);
      await dashboard.gotoDetail('Default Topic 1');
      await page.waitForURL('**/dashboard/default-topic-1');

      await dashboard.backButton.click();
      await page.waitForURL('**/dashboard');
      // 詳細ページのURLではないこと
      expect(page.url()).not.toContain('default-topic-1');
    });

    test('存在しないトピックで「トピックが見つかりません」表示', async ({ page, apiMock }) => {
      // トピックが空の状態で直接アクセス
      await apiMock.clearMock('TopicService', 'ListTopics');
      await apiMock.mockRPC('TopicService', 'ListTopics', { topics: [] });

      // ダッシュボード経由で auth state 確立
      await page.goto('/dashboard');
      await page.waitForLoadState('networkidle');
      await page.goto('/dashboard/nonexistent-id');

      await expect(page.getByText('トピックが見つかりません')).toBeVisible();
      await expect(page.getByRole('button', { name: 'ダッシュボードに戻る' })).toBeVisible();
    });

    test('関連トピックセクションが表示されない', async ({ page }) => {
      // デフォルトで同ジャンルのトピックが複数あっても関連トピックは表示されない
      const dashboard = new DashboardPage(page);
      await dashboard.gotoDetail('Default Topic 1');
      await page.waitForURL('**/dashboard/default-topic-1');

      await expect(dashboard.detailHero).toBeVisible();
      await expect(page.getByText('関連トピック')).not.toBeVisible();
    });
  });
});
