import { test, expect, authenticateContext } from '../fixtures/test-base';
import { create, toJson } from '@bufbuild/protobuf';
import {
  AnalyticsSummarySchema,
  DailyAnalyticsSchema,
  PostAnalyticsSchema,
  GrowthInsightSchema,
  GetAnalyticsSummaryResponseSchema,
  ListPostAnalyticsResponseSchema,
  GetGrowthInsightsResponseSchema,
} from '../../src/gen/trendbird/v1/analytics_pb';

test.describe('分析ページ', () => {
  test.beforeEach(async ({ context }) => {
    await authenticateContext(context);
  });

  test('空状態を表示', async ({ page, apiMock }) => {
    await apiMock.setupDefaults();
    await page.goto('/analytics');
    await expect(page.getByText('アナリティクスデータがありません')).toBeVisible();
  });

  test('概要タブにKPIカードを表示', async ({ page, apiMock }) => {
    const dailyData = [
      create(DailyAnalyticsSchema, { id: '1', date: '2026-03-25', impressions: 1000, likes: 10, engagements: 50, newFollows: 5, unfollows: 1 }),
      create(DailyAnalyticsSchema, { id: '2', date: '2026-03-26', impressions: 2000, likes: 20, engagements: 100, newFollows: 8, unfollows: 2 }),
    ];
    const summary = create(AnalyticsSummarySchema, {
      startDate: '2026-03-25',
      endDate: '2026-03-26',
      totalImpressions: 3000n,
      totalLikes: 30n,
      totalEngagements: 150n,
      totalNewFollows: 13n,
      totalUnfollows: 3n,
      daysCount: 2,
      postsCount: 5,
      dailyData,
    });

    await apiMock.setupDefaults();
    // Override after setupDefaults (last registered route wins)
    await apiMock.mockRPC('AnalyticsService', 'GetAnalyticsSummary',
      toJson(GetAnalyticsSummaryResponseSchema,
        create(GetAnalyticsSummaryResponseSchema, { summary })));

    await page.goto('/analytics');
    await expect(page.getByText('3,000')).toBeVisible();
    await expect(page.getByText('+10')).toBeVisible();
  });

  test('投稿タブで投稿一覧を表示', async ({ page, apiMock }) => {
    const posts = [
      create(PostAnalyticsSchema, {
        id: '1',
        postId: '111',
        postedAt: '2026-03-25T10:00:00Z',
        postText: 'テスト投稿です',
        impressions: 500,
        likes: 10,
        engagements: 20,
      }),
    ];

    await apiMock.setupDefaults();
    await apiMock.mockRPC('AnalyticsService', 'ListPostAnalytics',
      toJson(ListPostAnalyticsResponseSchema,
        create(ListPostAnalyticsResponseSchema, { posts, total: 1 })));

    await page.goto('/analytics');
    await page.getByRole('button', { name: '投稿' }).click();
    await expect(page.getByText('テスト投稿です')).toBeVisible();
    await expect(page.getByText('500')).toBeVisible();
  });

  test('インサイトタブでインサイトを表示', async ({ page, apiMock }) => {
    const insights = [
      create(GrowthInsightSchema, {
        category: 'engagement',
        insight: 'エンゲージメント率が良好です。',
        action: 'この調子で投稿を続けましょう。',
      }),
    ];
    const summary = create(AnalyticsSummarySchema, {
      startDate: '2026-03-01',
      endDate: '2026-03-31',
      daysCount: 1,
    });

    await apiMock.setupDefaults();
    await apiMock.mockRPC('AnalyticsService', 'GetGrowthInsights',
      toJson(GetGrowthInsightsResponseSchema,
        create(GetGrowthInsightsResponseSchema, { insights, summary })));

    await page.goto('/analytics');
    await page.getByRole('button', { name: 'インサイト' }).click();
    await page.getByRole('button', { name: 'インサイトを更新' }).click();
    await expect(page.getByText('エンゲージメント率が良好です。')).toBeVisible();
    await expect(page.getByText('この調子で投稿を続けましょう。')).toBeVisible();
  });
});
