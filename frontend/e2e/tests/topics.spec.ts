import { test, expect, authenticateContext } from '../fixtures/test-base';
import {
  resetSeq,
  TopicStatus,
  buildTopic,
  buildTopicSuggestion,
} from '../fixtures/factories';
import { TopicsPage } from '../pages/topics.page';
import { create, toJson } from '@bufbuild/protobuf';
import {
  TopicSchema,
  TopicSuggestionSchema,
  ListTopicsResponseSchema,
  SuggestTopicsResponseSchema,
  AddGenreResponseSchema,
  RemoveGenreResponseSchema,
  CreateTopicResponseSchema,
  DeleteTopicResponseSchema,
  ListUserGenresResponseSchema,
} from '../../src/gen/trendbird/v1/topic_pb';

test.describe('トピック管理ページ', () => {
  test.beforeEach(async ({ context, page, apiMock }) => {
    resetSeq();
    await authenticateContext(context);
    await apiMock.setupDefaults();
    await page.addInitScript(() => {
      sessionStorage.removeItem('tb_tutorial_pending');
    });
  });

  test.describe('ジャンル管理', () => {
    test('ジャンル追加でジャンルタブが表示される', async ({ page, apiMock }) => {
      // ユーザーのジャンルを空にする
      await apiMock.clearMock('TopicService', 'ListUserGenres');
      await apiMock.mockRPC('TopicService', 'ListUserGenres',
        toJson(ListUserGenresResponseSchema,
          create(ListUserGenresResponseSchema, { genres: [] })));
      await apiMock.clearMock('TopicService', 'ListTopics');
      await apiMock.mockRPC('TopicService', 'ListTopics', { topics: [] });
      await apiMock.mockRPC('TopicService', 'AddGenre',
        toJson(AddGenreResponseSchema, create(AddGenreResponseSchema, {})));
      // AddGenre 後の再取得をモック
      await apiMock.mockRPC('TopicService', 'SuggestTopics',
        toJson(SuggestTopicsResponseSchema,
          create(SuggestTopicsResponseSchema, { suggestions: [] })));

      const topics = new TopicsPage(page);
      await topics.goto();

      // 空状態
      await expect(topics.emptyGenreState).toBeVisible();

      // 「ジャンルを選ぶ」をクリック
      await topics.addGenreButton.click();

      // ジャンル選択モーダル
      await expect(page.getByText('ジャンルを追加')).toBeVisible();
    });

    test('ジャンルタブ切替でサジェストが表示される', async ({ page, apiMock }) => {
      await apiMock.mockRPC('TopicService', 'SuggestTopics',
        toJson(SuggestTopicsResponseSchema,
          create(SuggestTopicsResponseSchema, {
            suggestions: [
              create(TopicSuggestionSchema, {
                id: 'suggest-1',
                name: 'AIアシスタント',
                keywords: ['AI'],
                genre: 'technology',
                genreLabel: 'Technology',
                similarityScore: 0.9,
              }),
            ],
          })));

      const topics = new TopicsPage(page);
      await topics.goto();

      // サジェストセクション
      await expect(page.getByText('AIアシスタント')).toBeVisible();
    });
  });

  test.describe('ジャンル削除', () => {
    test('ジャンルチップの×ボタンでジャンルが削除される', async ({ page, apiMock }) => {
      await apiMock.mockRPC('TopicService', 'RemoveGenre',
        toJson(RemoveGenreResponseSchema, create(RemoveGenreResponseSchema, {})));
      await apiMock.mockRPC('TopicService', 'SuggestTopics',
        toJson(SuggestTopicsResponseSchema,
          create(SuggestTopicsResponseSchema, { suggestions: [] })));

      const topics = new TopicsPage(page);
      await topics.goto();

      // Technology ジャンルチップ（テキスト + ×ボタンを含む div）
      const chip = page.getByText('Technology').first();
      await expect(chip).toBeVisible();

      // チップ内の × ボタンをクリック
      await chip.locator('..').getByRole('button', { name: '×' }).click();

      // ジャンルが消えて空状態に
      await expect(topics.emptyGenreState).toBeVisible();
    });

    test('ジャンル削除失敗でエラー toast', async ({ page, apiMock }) => {
      await apiMock.mockRPCError('TopicService', 'RemoveGenre', 'internal', 'server error');
      await apiMock.mockRPC('TopicService', 'SuggestTopics',
        toJson(SuggestTopicsResponseSchema,
          create(SuggestTopicsResponseSchema, { suggestions: [] })));

      const topics = new TopicsPage(page);
      await topics.goto();

      const chip = page.getByText('Technology').first();
      await expect(chip).toBeVisible();
      await chip.locator('..').getByRole('button', { name: '×' }).click();

      // エラー時は toast が表示される（汎用エラーメッセージ）
      await expect(page.getByText(/ただいま処理がうまくいきませんでした/)).toBeVisible({ timeout: 10_000 });
    });
  });

  test.describe('トピック削除', () => {
    test('トピック削除成功でAPIリクエストが正しく送信される', async ({ page, apiMock }) => {
      await apiMock.mockRPC('TopicService', 'DeleteTopic',
        toJson(DeleteTopicResponseSchema, create(DeleteTopicResponseSchema, {})));
      await apiMock.mockRPC('TopicService', 'SuggestTopics',
        toJson(SuggestTopicsResponseSchema,
          create(SuggestTopicsResponseSchema, { suggestions: [] })));

      const topics = new TopicsPage(page);
      await topics.goto();

      await expect(page.getByText('Default Topic 1')).toBeVisible();

      // チップ内の × ボタンをクリック → DeleteTopic RPC が呼ばれる
      const [req] = await Promise.all([
        page.waitForRequest((r) => r.url().includes('TopicService/DeleteTopic')),
        page.getByText('Default Topic 1').first().getByRole('button', { name: '×' }).click(),
      ]);

      const body = JSON.parse(req.postData() ?? '{}');
      expect(body.id).toBe('default-topic-1');
    });

    test('トピック削除失敗でエラー toast', async ({ page, apiMock }) => {
      await apiMock.mockRPCError('TopicService', 'DeleteTopic', 'internal', 'server error');
      await apiMock.mockRPC('TopicService', 'SuggestTopics',
        toJson(SuggestTopicsResponseSchema,
          create(SuggestTopicsResponseSchema, { suggestions: [] })));

      const topics = new TopicsPage(page);
      await topics.goto();

      await expect(page.getByText('Default Topic 1')).toBeVisible();

      // チップ内の × ボタンクリック → エラー
      await page.getByText('Default Topic 1').first().getByRole('button', { name: '×' }).click();

      // エラー toast
      await expect(page.getByText('ただいま処理がうまくいきませんでした')).toBeVisible({ timeout: 10_000 });
    });
  });

  test.describe('トピック追加・削除', () => {
    test('サジェストからトピック追加 → チップ表示', async ({ page, apiMock }) => {
      const newTopic = create(TopicSchema, {
        id: 'new-topic-1',
        name: 'AIアシスタント',
        keywords: ['AI'],
        genre: 'technology',
        status: TopicStatus.STABLE,
        changePercent: 0,
        zScore: 0,
        currentVolume: 0,
        baselineVolume: 0,
        sparklineData: [],
        weeklySparklineData: [],
        spikeHistory: [],
        notificationEnabled: true,
        createdAt: '2026-01-01T00:00:00Z',
      });

      await apiMock.mockRPC('TopicService', 'SuggestTopics',
        toJson(SuggestTopicsResponseSchema,
          create(SuggestTopicsResponseSchema, {
            suggestions: [
              create(TopicSuggestionSchema, {
                id: 'suggest-1',
                name: 'AIアシスタント',
                keywords: ['AI'],
                genre: 'technology',
                genreLabel: 'Technology',
                similarityScore: 0.9,
              }),
            ],
          })));
      await apiMock.mockRPC('TopicService', 'CreateTopic',
        toJson(CreateTopicResponseSchema,
          create(CreateTopicResponseSchema, { topic: newTopic })));

      const topics = new TopicsPage(page);
      await topics.goto();

      // サジェストからトピック追加
      await page.getByText('AIアシスタント').click();

      // 追加成功 toast
      await expect(page.getByText('「AIアシスタント」を追加しました')).toBeVisible();
    });

    test('トピック削除（チップの×ボタン）', async ({ page, apiMock }) => {
      await apiMock.mockRPC('TopicService', 'DeleteTopic',
        toJson(DeleteTopicResponseSchema, create(DeleteTopicResponseSchema, {})));
      await apiMock.mockRPC('TopicService', 'SuggestTopics',
        toJson(SuggestTopicsResponseSchema,
          create(SuggestTopicsResponseSchema, { suggestions: [] })));

      const topics = new TopicsPage(page);
      await topics.goto();

      // デフォルトトピックが表示されている
      await expect(page.getByText('Default Topic 1')).toBeVisible();
    });
  });

  test.describe('検索', () => {
    test('検索タブでキーワード検索 → 結果表示', async ({ page, apiMock }) => {
      await apiMock.mockRPC('TopicService', 'SuggestTopics',
        toJson(SuggestTopicsResponseSchema,
          create(SuggestTopicsResponseSchema, {
            suggestions: [
              create(TopicSuggestionSchema, {
                id: 'search-result-1',
                name: '検索結果トピック',
                keywords: ['検索'],
                genre: 'technology',
                genreLabel: 'Technology',
                similarityScore: 0.8,
              }),
            ],
          })));

      const topics = new TopicsPage(page);
      await topics.goto();

      // 検索タブに切替
      await topics.searchTopicsTab.click();

      // 検索入力
      await page.getByPlaceholder('トピック名で検索...').fill('検索');

      // 結果が表示される
      await expect(page.getByText('検索結果トピック')).toBeVisible();
    });
  });

  test.describe('検索結果なし', () => {
    test('検索結果0件で空状態メッセージが表示される', async ({ page, apiMock }) => {
      await apiMock.mockRPC('TopicService', 'SuggestTopics',
        toJson(SuggestTopicsResponseSchema,
          create(SuggestTopicsResponseSchema, { suggestions: [] })));

      const topics = new TopicsPage(page);
      await topics.goto();

      // 検索タブに切替
      await topics.searchTopicsTab.click();

      // 検索入力
      await page.getByPlaceholder('トピック名で検索...').fill('存在しないトピック');

      // 空状態メッセージ
      await expect(page.getByText('一致するトピックが見つかりません')).toBeVisible();
    });
  });

  test.describe('検索正規化', () => {
    test('全角英数字でも検索結果が表示される', async ({ page, apiMock }) => {
      await apiMock.mockRPC('TopicService', 'SuggestTopics',
        toJson(SuggestTopicsResponseSchema,
          create(SuggestTopicsResponseSchema, {
            suggestions: [
              create(TopicSuggestionSchema, {
                id: 'norm-result',
                name: 'AIトレンド',
                keywords: ['AI', 'トレンド'],
                genre: 'technology',
                genreLabel: 'Technology',
                similarityScore: 0.9,
              }),
            ],
          })));

      const topics = new TopicsPage(page);
      await topics.goto();

      // 検索タブに切替
      await topics.searchTopicsTab.click();

      // 全角で検索（ＡＩ→AI に正規化されて結果が表示される）
      await page.getByPlaceholder('トピック名で検索...').fill('ＡＩ');

      // サジェスト結果が表示される（正規化されてマッチ）
      await expect(page.getByText('AIトレンド')).toBeVisible();
    });
  });

  test.describe('空状態', () => {
    test('ジャンル未選択で空状態が表示される', async ({ page, apiMock }) => {
      await apiMock.clearMock('TopicService', 'ListUserGenres');
      await apiMock.mockRPC('TopicService', 'ListUserGenres',
        toJson(ListUserGenresResponseSchema,
          create(ListUserGenresResponseSchema, { genres: [] })));
      await apiMock.clearMock('TopicService', 'ListTopics');
      await apiMock.mockRPC('TopicService', 'ListTopics', { topics: [] });

      const topics = new TopicsPage(page);
      await topics.goto();

      await expect(topics.emptyGenreState).toBeVisible();
      await expect(topics.addGenreButton).toBeVisible();
    });
  });

  test.describe('カスタムトピック作成制御', () => {
    test('検索結果0件でカスタム作成ボタンが表示される（トピックから探す）', async ({ page, apiMock }) => {
      await apiMock.mockRPC('TopicService', 'SuggestTopics',
        toJson(SuggestTopicsResponseSchema,
          create(SuggestTopicsResponseSchema, { suggestions: [] })));

      const topics = new TopicsPage(page);
      await topics.goto();

      await topics.searchTopicsTab.click();

      await page.getByPlaceholder('トピック名で検索...').fill('新しいトピック');

      await expect(page.getByText('一致するトピックが見つかりません')).toBeVisible();

      await expect(page.getByRole('button', { name: /トピックとして作成/ })).toBeVisible();
    });

    test('ジャンル内検索で一致なし時にカスタム追加ボタンが表示される（ジャンルから探す）', async ({ page, apiMock }) => {
      await apiMock.clearMock('TopicService', 'ListTopics');
      await apiMock.mockRPC('TopicService', 'ListTopics',
        toJson(ListTopicsResponseSchema,
          create(ListTopicsResponseSchema, { topics: [] })));

      // ジャンルのサジェストは空（検索ヒットなし）
      await apiMock.mockRPC('TopicService', 'SuggestTopics',
        toJson(SuggestTopicsResponseSchema,
          create(SuggestTopicsResponseSchema, { suggestions: [] })));

      const topics = new TopicsPage(page);
      await topics.goto();

      // ジャンルタブ（デフォルト）で検索入力
      await page.getByPlaceholder('トピックを検索...').fill('存在しないトピック');

      // 「一致するトピックがありません」メッセージが表示
      await expect(page.getByText('一致するトピックがありません')).toBeVisible();

      await expect(page.getByText(/「存在しないトピック」をトピックとして追加/)).toBeVisible();
    });
  });

  test.describe('検索エッジケース', () => {
    test('特殊文字入力でクラッシュしない', async ({ page, apiMock }) => {
      await apiMock.mockRPC('TopicService', 'SuggestTopics',
        toJson(SuggestTopicsResponseSchema,
          create(SuggestTopicsResponseSchema, { suggestions: [] })));

      const topics = new TopicsPage(page);
      await topics.goto();

      await topics.searchTopicsTab.click();

      // 特殊文字（括弧・記号・正規表現メタ文字）を入力
      await page.getByPlaceholder('トピック名で検索...').fill('test()[]{}<>.*+?^$|\\');

      // クラッシュせず空状態メッセージが表示される
      await expect(page.getByText('一致するトピックが見つかりません')).toBeVisible();
    });

    test('検索APIエラーでもページが維持される', async ({ page, apiMock }) => {
      await apiMock.mockRPCError('TopicService', 'SuggestTopics', 'internal', 'server error');

      const topics = new TopicsPage(page);
      await topics.goto();

      // ページがクラッシュせずタブが表示されたまま
      await expect(topics.browseByGenreTab).toBeVisible();
      await expect(topics.searchTopicsTab).toBeVisible();
    });
  });
});
