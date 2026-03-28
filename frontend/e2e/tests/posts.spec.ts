import { test, expect, authenticateContext } from '../fixtures/test-base';
import {
  resetSeq,
  TopicStatus,
  PostStatus,
  buildTopic,
  buildDraft,
  buildScheduledPost,
  buildPostHistory,
  buildGeneratedPost,
} from '../fixtures/factories';
import { expectToast } from '../helpers/assertions';
import { PostsPage } from '../pages/posts.page';
import { create, toJson } from '@bufbuild/protobuf';
import {
  ScheduledPostSchema,
  PostHistorySchema,
  ListDraftsResponseSchema,
  ListPostHistoryResponseSchema,
  CreateDraftResponseSchema,
  DeleteDraftResponseSchema,
  PublishPostResponseSchema,
  GeneratePostsResponseSchema,
  GeneratedPostSchema,
  UpdateDraftResponseSchema,
  SchedulePostResponseSchema,
} from '../../src/gen/trendbird/v1/post_pb';
import {
  TopicSchema,
  ListTopicsResponseSchema,
} from '../../src/gen/trendbird/v1/topic_pb';
import {
  TwitterConnectionInfoSchema,
  TwitterConnectionStatus,
  GetConnectionInfoResponseSchema,
} from '../../src/gen/trendbird/v1/twitter_pb';
import {
  UserSchema,
  GetCurrentUserResponseSchema,
} from '../../src/gen/trendbird/v1/auth_pb';

test.describe('投稿ページ', () => {
  test.beforeEach(async ({ context, page, apiMock }) => {
    resetSeq();
    await authenticateContext(context);
    await apiMock.setupDefaults();
    // X連携済みにセット
    await apiMock.clearMock('TwitterService', 'GetConnectionInfo');
    await apiMock.mockRPC('TwitterService', 'GetConnectionInfo',
      toJson(GetConnectionInfoResponseSchema,
        create(GetConnectionInfoResponseSchema, {
          info: create(TwitterConnectionInfoSchema, {
            status: TwitterConnectionStatus.CONNECTED,
            connectedAt: '2026-01-15T10:00:00Z',
          }),
        })));
    await page.addInitScript(() => {
      sessionStorage.removeItem('tb_tutorial_pending');
    });
  });

  // ─── トピック選択ペイン ─────────────────────────────────────

  test.describe('トピック選択ペイン', () => {
    test('トピックカード一覧が表示される', async ({ page }) => {
      const posts = new PostsPage(page);
      await posts.goto();

      await expect(posts.topicCard('Default Topic 1')).toBeVisible();
      await expect(posts.topicCard('Default Topic 2')).toBeVisible();
    });

    test('ステータスフィルタ切替', async ({ page, apiMock }) => {
      const spikeTopic = buildTopic({ id: 'spike-1', name: 'Spike Topic', status: TopicStatus.SPIKE, zScore: 3.5 });
      const stableTopic = buildTopic({ id: 'stable-1', name: 'Stable Topic', status: TopicStatus.STABLE, zScore: 1.0 });
      await apiMock.clearMock('TopicService', 'ListTopics');
      await apiMock.mockRPC('TopicService', 'ListTopics', { topics: [spikeTopic, stableTopic] });

      const posts = new PostsPage(page);
      await posts.goto();

      // 「全て」で両方表示
      await expect(page.getByText('Spike Topic')).toBeVisible();
      await expect(page.getByText('Stable Topic')).toBeVisible();

      // 「話題沸騰」フィルタ
      await posts.statusFilterButton('話題沸騰').click();
      await expect(page.getByText('Spike Topic')).toBeVisible();
      await expect(page.getByText('Stable Topic')).not.toBeVisible();

      // 「安定」フィルタ
      await posts.statusFilterButton('安定').click();
      await expect(page.getByText('Stable Topic')).toBeVisible();
      await expect(page.getByText('Spike Topic')).not.toBeVisible();
    });

    test('トピック選択 → 右ペインに選択トピック表示', async ({ page }) => {
      const posts = new PostsPage(page);
      await posts.goto();

      // 初期状態: トピック未選択
      await expect(posts.selectTopicPrompt).toBeVisible();

      // トピック選択
      await posts.topicCard('Default Topic 1').click();

      // 選択トピックのコンテキストが表示される
      await expect(page.getByText('Default Topic 1').first()).toBeVisible();
    });

    test('0件フィルタ結果', async ({ page, apiMock }) => {
      const stableTopic = buildTopic({ id: 'stable-1', name: 'Stable Only', status: TopicStatus.STABLE });
      await apiMock.clearMock('TopicService', 'ListTopics');
      await apiMock.mockRPC('TopicService', 'ListTopics', { topics: [stableTopic] });

      const posts = new PostsPage(page);
      await posts.goto();

      // spike フィルタで0件
      await posts.statusFilterButton('話題沸騰').click();
      await expect(posts.noTopicsMessage).toBeVisible();
    });
  });

  // ─── AI生成 ────────────────────────────────────────────────

  test.describe('AI生成', () => {
    test('生成ボタンクリック → 結果表示', async ({ page, apiMock }) => {
      const genPost = buildGeneratedPost({ content: 'AIで生成されたテスト投稿', topicId: 'default-topic-1' });
      await apiMock.mockRPC('PostService', 'GeneratePosts', { posts: [genPost] });

      const posts = new PostsPage(page);
      await posts.goto();

      // トピック選択
      await posts.topicCard('Default Topic 1').click();

      // 生成ボタンクリック
      await posts.generateButton.click();

      // 結果が表示される
      await expect(page.getByText('AIで生成されたテスト投稿')).toBeVisible();
    });

  });

  // ─── コンポーザー ──────────────────────────────────────────

  test.describe('コンポーザー', () => {
    test('テキスト入力と文字カウンター', async ({ page }) => {
      const posts = new PostsPage(page);
      await posts.goto();

      // トピック選択
      await posts.topicCard('Default Topic 1').click();

      // テキスト入力
      await posts.composerTextarea.fill('テスト投稿文です');

      // 文字カウンター
      await expect(posts.charCounter).toBeVisible();
    });

    test('下書き保存成功 → toast', async ({ page, apiMock }) => {
      const draft = create(ScheduledPostSchema, {
        id: 'new-draft-1',
        content: '下書きテスト',
        status: 1, // DRAFT
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: '2026-01-01T00:00:00Z',
        characterCount: 6,
      });
      await apiMock.mockRPC('PostService', 'CreateDraft',
        toJson(CreateDraftResponseSchema,
          create(CreateDraftResponseSchema, { draft })));

      const posts = new PostsPage(page);
      await posts.goto();

      // トピック選択 + テキスト入力
      await posts.topicCard('Default Topic 1').click();
      await posts.composerTextarea.fill('下書きテスト');

      // 下書き保存
      await posts.saveDraftButton.click();

      await expect(page.getByText('下書きに保存しました')).toBeVisible();
    });

    test('X未連携: X未連携ラベル表示', async ({ page, apiMock }) => {
      // X 未連携に戻す
      await apiMock.clearMock('TwitterService', 'GetConnectionInfo');
      await apiMock.mockRPC('TwitterService', 'GetConnectionInfo',
        toJson(GetConnectionInfoResponseSchema,
          create(GetConnectionInfoResponseSchema, {
            info: create(TwitterConnectionInfoSchema, {
              status: TwitterConnectionStatus.DISCONNECTED,
            }),
          })));

      const posts = new PostsPage(page);
      await posts.goto();

      // トピック選択 + テキスト入力
      await posts.topicCard('Default Topic 1').click();
      await posts.composerTextarea.fill('テスト投稿');

      // X未連携表示
      await expect(posts.xDisconnectedLabel).toBeVisible();
    });
  });

  // ─── ジャンルフィルタ ─────────────────────────────────────

  test.describe('ジャンルフィルタ', () => {
    test('ジャンルボタンでトピックがフィルタされる', async ({ page, apiMock }) => {
      const techTopic = buildTopic({ id: 'tech-1', name: 'Tech Topic', genre: 'technology', status: TopicStatus.SPIKE, zScore: 3.0 });
      const bizTopic = buildTopic({ id: 'biz-1', name: 'Business Topic', genre: 'business', status: TopicStatus.STABLE, zScore: 1.0 });
      await apiMock.clearMock('TopicService', 'ListTopics');
      await apiMock.mockRPC('TopicService', 'ListTopics', { topics: [techTopic, bizTopic] });
      // ユーザーのジャンルに business も追加
      await apiMock.clearMock('TopicService', 'ListUserGenres');
      const { create: c, toJson: tj } = await import('@bufbuild/protobuf');
      const { ListUserGenresResponseSchema: Schema } = await import('../../src/gen/trendbird/v1/topic_pb');
      await apiMock.mockRPC('TopicService', 'ListUserGenres',
        tj(Schema, c(Schema, { genres: ['technology', 'business'] })));

      const posts = new PostsPage(page);
      await posts.goto();

      // ジャンル「全て」でデフォルト表示 → 両方見える
      await expect(page.getByText('Tech Topic')).toBeVisible();
      await expect(page.getByText('Business Topic')).toBeVisible();

      // Technology フィルタ
      await posts.genreFilterButton('Technology').click();
      await expect(page.getByText('Tech Topic')).toBeVisible();
      await expect(page.getByText('Business Topic')).not.toBeVisible();

      // Business フィルタ
      await posts.genreFilterButton('Business').click();
      await expect(page.getByText('Business Topic')).toBeVisible();
      await expect(page.getByText('Tech Topic')).not.toBeVisible();
    });

    test('ジャンルフィルタ後に0件メッセージ', async ({ page, apiMock }) => {
      const techTopic = buildTopic({ id: 'tech-only', name: 'Tech Only', genre: 'technology', status: TopicStatus.SPIKE });
      await apiMock.clearMock('TopicService', 'ListTopics');
      await apiMock.mockRPC('TopicService', 'ListTopics', { topics: [techTopic] });
      // business ジャンルも追加して business フィルタを有効にする
      await apiMock.clearMock('TopicService', 'ListUserGenres');
      const { create: c, toJson: tj } = await import('@bufbuild/protobuf');
      const { ListUserGenresResponseSchema: Schema } = await import('../../src/gen/trendbird/v1/topic_pb');
      await apiMock.mockRPC('TopicService', 'ListUserGenres',
        tj(Schema, c(Schema, { genres: ['technology', 'business'] })));

      const posts = new PostsPage(page);
      await posts.goto();

      // Business フィルタで0件
      await posts.genreFilterButton('Business').click();
      await expect(posts.noTopicsMessage).toBeVisible();
    });
  });

  // ─── 予約投稿 ──────────────────────────────────────────────

  test.describe('予約投稿', () => {
    test('Pro: 予約ボタンが表示される', async ({ page }) => {
      const posts = new PostsPage(page);
      await posts.goto();

      // トピック選択
      await posts.topicCard('Default Topic 1').click();

      // 予約ボタン（コンポーザー内、exact true）
      await expect(posts.scheduleButton).toBeVisible();
    });

  });

  // ─── 即時公開 ──────────────────────────────────────────────

  test.describe('即時公開', () => {
    test('X連携済み: 確認ダイアログ → 公開成功', async ({ page, apiMock }) => {
      const publishedPost = create(PostHistorySchema, {
        id: 'published-1',
        content: 'テスト公開投稿',
        publishedAt: '2026-01-01T12:00:00Z',
        likes: 0,
        retweets: 0,
        replies: 0,
        views: 0,
      });
      await apiMock.mockRPC('PostService', 'CreateDraft',
        toJson(CreateDraftResponseSchema,
          create(CreateDraftResponseSchema, {
            draft: create(ScheduledPostSchema, {
              id: 'temp-draft',
              content: 'テスト公開投稿',
              status: 1,
              createdAt: '2026-01-01T00:00:00Z',
              updatedAt: '2026-01-01T00:00:00Z',
              characterCount: 7,
            }),
          })));
      await apiMock.mockRPC('PostService', 'PublishPost',
        toJson(PublishPostResponseSchema,
          create(PublishPostResponseSchema, { post: publishedPost })));

      const posts = new PostsPage(page);
      await posts.goto();

      // トピック選択 + テキスト入力
      await posts.topicCard('Default Topic 1').click();
      await posts.composerTextarea.fill('テスト公開投稿');

      // 「今すぐ投稿」をクリック
      await posts.publishButton.click();

      // 確認ダイアログ
      await expect(posts.publishDialog).toBeVisible();
      await page.getByRole('button', { name: '投稿する' }).click();

      // 成功 toast
      await expect(page.getByText('投稿しました')).toBeVisible();
    });

    test('公開失敗 → エラー toast', async ({ page, apiMock }) => {
      await apiMock.mockRPC('PostService', 'CreateDraft',
        toJson(CreateDraftResponseSchema,
          create(CreateDraftResponseSchema, {
            draft: create(ScheduledPostSchema, {
              id: 'temp-draft',
              content: '失敗テスト',
              status: 1,
              createdAt: '2026-01-01T00:00:00Z',
              updatedAt: '2026-01-01T00:00:00Z',
              characterCount: 5,
            }),
          })));
      await apiMock.mockRPCError('PostService', 'PublishPost', 'internal', '投稿に失敗しました');

      const posts = new PostsPage(page);
      await posts.goto();

      await posts.topicCard('Default Topic 1').click();
      await posts.composerTextarea.fill('失敗テスト');
      await posts.publishButton.click();

      // 確認ダイアログで投稿
      await expect(posts.publishDialog).toBeVisible();
      await page.getByRole('button', { name: '投稿する' }).click();

      // エラー toast
      await expect(page.getByText('投稿がうまくいきませんでした')).toBeVisible({ timeout: 10_000 });
    });
  });

  // ─── 管理セクション ────────────────────────────────────────

  test.describe('管理セクション', () => {
    test('下書き/予約/履歴タブ切替', async ({ page, apiMock }) => {
      // 下書きと履歴にデータを入れる
      const draft = create(ScheduledPostSchema, {
        id: 'draft-1',
        content: 'テスト下書き内容',
        status: 1, // DRAFT
        topicName: 'Default Topic 1',
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: '2026-01-01T00:00:00Z',
        characterCount: 8,
      });
      const scheduled = create(ScheduledPostSchema, {
        id: 'sched-1',
        content: 'テスト予約投稿内容',
        status: 2, // SCHEDULED
        scheduledAt: '2026-03-01T12:00:00Z',
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: '2026-01-01T00:00:00Z',
        characterCount: 9,
      });
      const history = create(PostHistorySchema, {
        id: 'hist-1',
        content: 'テスト投稿履歴内容',
        publishedAt: '2026-01-15T12:00:00Z',
        likes: 10,
        retweets: 5,
        replies: 2,
        views: 1000,
      });

      await apiMock.clearMock('PostService', 'ListDrafts');
      await apiMock.mockRPC('PostService', 'ListDrafts',
        toJson(ListDraftsResponseSchema,
          create(ListDraftsResponseSchema, { drafts: [draft, scheduled] })));
      await apiMock.clearMock('PostService', 'ListPostHistory');
      await apiMock.mockRPC('PostService', 'ListPostHistory',
        toJson(ListPostHistoryResponseSchema,
          create(ListPostHistoryResponseSchema, { posts: [history] })));

      const posts = new PostsPage(page);
      await posts.goto();

      // 管理セクションはデフォルトで開いている
      // 下書きタブ（デフォルト）
      await expect(page.getByText('テスト下書き内容')).toBeVisible();

      // 予約タブ
      await posts.scheduledTab.click();
      await expect(page.getByText('テスト予約投稿内容')).toBeVisible();

      // 履歴タブ
      await posts.historyTab.click();
      await expect(page.getByText('テスト投稿履歴内容')).toBeVisible();
    });

    test('下書き削除 → 確認 → 削除', async ({ page, apiMock }) => {
      const draft = create(ScheduledPostSchema, {
        id: 'draft-del',
        content: '削除対象の下書き文章',
        status: 1,
        topicName: 'Default Topic 1',
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: '2026-01-01T00:00:00Z',
        characterCount: 10,
      });
      await apiMock.clearMock('PostService', 'ListDrafts');
      await apiMock.mockRPC('PostService', 'ListDrafts',
        toJson(ListDraftsResponseSchema,
          create(ListDraftsResponseSchema, { drafts: [draft] })));
      await apiMock.mockRPC('PostService', 'DeleteDraft',
        toJson(DeleteDraftResponseSchema, create(DeleteDraftResponseSchema, {})));

      const posts = new PostsPage(page);
      await posts.goto();

      // 管理セクションはデフォルトで開いている
      // 下書きが表示される
      await expect(page.getByText('削除対象の下書き文章')).toBeVisible();

      // 削除ボタン（aria-label）をクリック
      await page.getByRole('button', { name: '削除' }).first().click();

      // 確認ダイアログ
      await expect(posts.deleteDialog).toBeVisible();
      await page.getByRole('button', { name: '削除する' }).click();

      // 削除成功 toast
      await expect(page.getByText('削除しました')).toBeVisible();
    });
  });

  // ─── 履歴タブ: メトリクス表示 ──────────────────────────────

  test.describe('履歴タブ: メトリクス表示', () => {
    test('いいね・RT・返信・表示数が表示される', async ({ page, apiMock }) => {
      const history = create(PostHistorySchema, {
        id: 'hist-metrics',
        content: 'メトリクス付き投稿',
        publishedAt: '2026-01-15T12:00:00Z',
        likes: 150,
        retweets: 42,
        replies: 8,
        views: 5200,
        tweetUrl: 'https://x.com/user/status/123',
      });
      await apiMock.clearMock('PostService', 'ListPostHistory');
      await apiMock.mockRPC('PostService', 'ListPostHistory',
        toJson(ListPostHistoryResponseSchema,
          create(ListPostHistoryResponseSchema, { posts: [history] })));

      const posts = new PostsPage(page);
      await posts.goto();

      // 履歴タブに切替
      await posts.historyTab.click();

      // メトリクスが表示される
      await expect(page.getByText('メトリクス付き投稿')).toBeVisible();
      await expect(page.getByText('150')).toBeVisible();
      await expect(page.getByText('42')).toBeVisible();
      // views は formatNumber で 5200→5.2k に変換される
      await expect(page.getByText('5.2k')).toBeVisible();
    });
  });

  // ─── 空入力・文字数制限 ──────────────────────────────────

  test.describe('空入力・文字数制限', () => {
    test('空白入力でボタンが無効化される', async ({ page }) => {
      const posts = new PostsPage(page);
      await posts.goto();

      // トピック選択
      await posts.topicCard('Default Topic 1').click();

      // 空のまま: 下書き保存・予約・今すぐ投稿が disabled
      await expect(posts.saveDraftButton).toBeDisabled();
      await expect(posts.scheduleButton).toBeDisabled();
      await expect(posts.publishButton).toBeDisabled();

      // 空白のみ入力しても disabled のまま
      await posts.composerTextarea.fill('   ');
      await expect(posts.saveDraftButton).toBeDisabled();
    });

    test('280文字で投稿ボタンが有効、281文字で無効', async ({ page }) => {
      const posts = new PostsPage(page);
      await posts.goto();

      await posts.topicCard('Default Topic 1').click();

      // 280文字ちょうど → 投稿ボタン有効
      const text280 = 'あ'.repeat(280);
      await posts.composerTextarea.fill(text280);
      await expect(posts.publishButton).toBeEnabled();
      await expect(posts.charCounter).toContainText('280/280');

      // 281文字 → 投稿ボタン無効
      const text281 = 'あ'.repeat(281);
      await posts.composerTextarea.fill(text281);
      await expect(posts.publishButton).toBeDisabled();
    });
  });

  // ─── 失敗した下書き ──────────────────────────────────────

  test.describe('失敗した下書き', () => {
    test('失敗バッジとエラーメッセージが表示される', async ({ page, apiMock }) => {
      const failedDraft = buildDraft({
        id: 'failed-1',
        content: '失敗した投稿内容',
        status: PostStatus.FAILED,
        topicName: 'Default Topic 1',
        errorMessage: 'X API との接続に失敗しました',
      });
      await apiMock.clearMock('PostService', 'ListDrafts');
      await apiMock.mockRPC('PostService', 'ListDrafts',
        toJson(ListDraftsResponseSchema,
          create(ListDraftsResponseSchema, { drafts: [
            create(ScheduledPostSchema, {
              id: 'failed-1',
              content: '失敗した投稿内容',
              status: PostStatus.FAILED,
              topicName: 'Default Topic 1',
              errorMessage: 'X API との接続に失敗しました',
              createdAt: '2026-01-01T00:00:00Z',
              updatedAt: '2026-01-01T00:00:00Z',
              characterCount: 8,
            }),
          ] })));

      const posts = new PostsPage(page);
      await posts.goto();

      // 失敗バッジが表示（exactで「失敗」テキストのみにマッチ）
      await expect(page.getByText('失敗', { exact: true })).toBeVisible();
      // エラーメッセージが表示
      await expect(page.getByText('X API との接続に失敗しました')).toBeVisible();
    });
  });

  // ─── 管理セクション空状態 ───────────────────────────────────

  test.describe('管理セクション空状態', () => {
    test('各タブで空状態メッセージが表示される', async ({ page }) => {
      const posts = new PostsPage(page);
      await posts.goto();

      // 下書きタブ（デフォルト）
      await expect(posts.noDrafts).toBeVisible();

      // 予約タブ
      await posts.scheduledTab.click();
      await expect(posts.noScheduled).toBeVisible();

      // 履歴タブ
      await posts.historyTab.click();
      await expect(posts.noHistory).toBeVisible();
    });
  });

  // ─── 予約投稿の残り時間 ────────────────────────────────────

  test.describe('予約投稿の残り時間', () => {
    test('予約タブで残り時間が表示される', async ({ page, apiMock }) => {
      // 未来の日時で予約
      const futureDate = new Date();
      futureDate.setDate(futureDate.getDate() + 3);
      const scheduled = create(ScheduledPostSchema, {
        id: 'sched-time',
        content: '残り時間テスト投稿',
        status: 2, // SCHEDULED
        scheduledAt: futureDate.toISOString(),
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: '2026-01-01T00:00:00Z',
        characterCount: 10,
      });
      await apiMock.clearMock('PostService', 'ListDrafts');
      await apiMock.mockRPC('PostService', 'ListDrafts',
        toJson(ListDraftsResponseSchema,
          create(ListDraftsResponseSchema, { drafts: [scheduled] })));

      const posts = new PostsPage(page);
      await posts.goto();

      // 予約タブに切替
      await posts.scheduledTab.click();

      // 「あとN日」が表示される（3日後なので「あと3日」or「あと2日」の範囲）
      await expect(page.getByText(/あと\d+日/)).toBeVisible();
    });
  });

  // ─── 下書き編集モーダル ──────────────────────────────────────

  test.describe('下書き編集モーダル', () => {
    test('編集モーダルで内容を変更して保存', async ({ page, apiMock }) => {
      const draft = create(ScheduledPostSchema, {
        id: 'draft-edit-1',
        content: '編集前の下書き内容',
        status: 1,
        topicName: 'Default Topic 1',
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: '2026-01-01T00:00:00Z',
        characterCount: 9,
      });
      await apiMock.clearMock('PostService', 'ListDrafts');
      await apiMock.mockRPC('PostService', 'ListDrafts',
        toJson(ListDraftsResponseSchema,
          create(ListDraftsResponseSchema, { drafts: [draft] })));

      const updatedDraft = create(ScheduledPostSchema, {
        id: 'draft-edit-1',
        content: '編集後の下書き内容',
        status: 1,
        topicName: 'Default Topic 1',
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: '2026-01-01T12:00:00Z',
        characterCount: 9,
      });
      await apiMock.mockRPC('PostService', 'UpdateDraft',
        toJson(UpdateDraftResponseSchema,
          create(UpdateDraftResponseSchema, { draft: updatedDraft })));

      const posts = new PostsPage(page);
      await posts.goto();

      // 下書きが表示される
      await expect(page.getByText('編集前の下書き内容')).toBeVisible();

      // 編集ボタンをクリック
      await page.getByRole('button', { name: '編集' }).first().click();

      // 編集モーダルが開く
      await expect(posts.editModal).toBeVisible();

      // モーダル内のテキストエリア（placeholderで区別）に編集前の内容が入っている → 書き換え
      const modalTextarea = page.getByPlaceholder('投稿文を入力してください…');
      await modalTextarea.fill('編集後の下書き内容');

      // 保存ボタンをクリック
      await page.getByRole('button', { name: '保存する' }).click();

      // 成功 toast
      await expectToast(page, '投稿を更新しました');
    });

    test('編集保存失敗 → エラー toast', async ({ page, apiMock }) => {
      const draft = create(ScheduledPostSchema, {
        id: 'draft-edit-fail',
        content: '失敗テスト下書き',
        status: 1,
        topicName: 'Default Topic 1',
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: '2026-01-01T00:00:00Z',
        characterCount: 8,
      });
      await apiMock.clearMock('PostService', 'ListDrafts');
      await apiMock.mockRPC('PostService', 'ListDrafts',
        toJson(ListDraftsResponseSchema,
          create(ListDraftsResponseSchema, { drafts: [draft] })));
      await apiMock.mockRPCError('PostService', 'UpdateDraft', 'internal', 'server error');

      const posts = new PostsPage(page);
      await posts.goto();

      // 編集ボタンをクリック
      await page.getByRole('button', { name: '編集' }).first().click();
      await expect(posts.editModal).toBeVisible();

      // 内容を変更して保存
      const modalTextarea = page.getByPlaceholder('投稿文を入力してください…');
      await modalTextarea.fill('変更後の内容');
      await page.getByRole('button', { name: '保存する' }).click();

      // エラー toast
      await expectToast(page, '更新に失敗しました');
    });
  });

  // ─── 予約モーダル ────────────────────────────────────────────

  test.describe('予約モーダル', () => {
    test('コンポーザーから予約 → モーダル表示 → 予約成功', async ({ page, apiMock }) => {
      const tempDraft = create(ScheduledPostSchema, {
        id: 'temp-sched-draft',
        content: '予約テスト投稿',
        status: 1,
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: '2026-01-01T00:00:00Z',
        characterCount: 7,
      });
      await apiMock.mockRPC('PostService', 'CreateDraft',
        toJson(CreateDraftResponseSchema,
          create(CreateDraftResponseSchema, { draft: tempDraft })));
      await apiMock.mockRPC('PostService', 'SchedulePost',
        toJson(SchedulePostResponseSchema,
          create(SchedulePostResponseSchema, {})));

      const posts = new PostsPage(page);
      await posts.goto();

      // トピック選択 + テキスト入力
      await posts.topicCard('Default Topic 1').click();
      await posts.composerTextarea.fill('予約テスト投稿');

      // 予約ボタンをクリック
      await posts.scheduleButton.click();

      // 予約モーダルが開く
      await expect(posts.scheduleModal).toBeVisible();

      // カスタム DateTimePicker: 次月に移動 → 日付 → 時刻をクリック
      // 月ヘッダーの次月ボタン（右端のナビボタン）
      await page.getByText(/\d{4}年 \d{1,2}月/).locator('..').locator('button').last().click();
      // 日付「15」（enabled のみ選択）
      await page.locator('button:not([disabled]):text-is("15")').click();
      // 時刻「12」（「時刻を選択」セクション内）
      await page.getByText('時刻を選択').locator('..').getByRole('button', { name: '12' }).click();

      // 予約ボタンをクリック
      await page.getByRole('button', { name: '予約する' }).click();

      // 成功 toast
      await expectToast(page, '投稿を予約しました');
    });

    test('既存下書きから予約 → モーダル表示 → 予約成功', async ({ page, apiMock }) => {
      const draft = create(ScheduledPostSchema, {
        id: 'draft-sched-1',
        content: '下書きから予約テスト',
        status: 1,
        topicName: 'Default Topic 1',
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: '2026-01-01T00:00:00Z',
        characterCount: 10,
      });
      await apiMock.clearMock('PostService', 'ListDrafts');
      await apiMock.mockRPC('PostService', 'ListDrafts',
        toJson(ListDraftsResponseSchema,
          create(ListDraftsResponseSchema, { drafts: [draft] })));
      await apiMock.mockRPC('PostService', 'SchedulePost',
        toJson(SchedulePostResponseSchema,
          create(SchedulePostResponseSchema, {})));

      const posts = new PostsPage(page);
      await posts.goto();

      // 下書きが表示される
      await expect(page.getByText('下書きから予約テスト')).toBeVisible();

      // 下書きカード内の予約ボタン（IconBtn aria-label="予約"）をクリック
      // コンポーザーの予約ボタン（disabled）と区別するためカード内にスコープ
      await page.getByText('下書きから予約テスト').locator('..').getByRole('button', { name: '予約' }).click();

      // 予約モーダルが開く
      await expect(posts.scheduleModal).toBeVisible();

      // カスタム DateTimePicker: 次月移動 → 日付 → 時刻
      await page.getByText(/\d{4}年 \d{1,2}月/).locator('..').locator('button').last().click();
      await page.locator('button:not([disabled]):text-is("15")').click();
      await page.getByText('時刻を選択').locator('..').getByRole('button', { name: '15' }).click();

      // 予約ボタンをクリック
      await page.getByRole('button', { name: '予約する' }).click();

      // 成功 toast
      await expectToast(page, '投稿を予約しました');
    });

    test('予約失敗 → エラー toast', async ({ page, apiMock }) => {
      const tempDraft = create(ScheduledPostSchema, {
        id: 'temp-sched-fail',
        content: '予約失敗テスト',
        status: 1,
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: '2026-01-01T00:00:00Z',
        characterCount: 7,
      });
      await apiMock.mockRPC('PostService', 'CreateDraft',
        toJson(CreateDraftResponseSchema,
          create(CreateDraftResponseSchema, { draft: tempDraft })));
      await apiMock.mockRPCError('PostService', 'SchedulePost', 'internal', 'schedule failed');

      const posts = new PostsPage(page);
      await posts.goto();

      await posts.topicCard('Default Topic 1').click();
      await posts.composerTextarea.fill('予約失敗テスト');
      await posts.scheduleButton.click();

      await expect(posts.scheduleModal).toBeVisible();
      // カスタム DateTimePicker: 次月移動 → 日付 → 時刻
      await page.getByText(/\d{4}年 \d{1,2}月/).locator('..').locator('button').last().click();
      await page.locator('button:not([disabled]):text-is("15")').click();
      await page.getByText('時刻を選択').locator('..').getByRole('button', { name: '12' }).click();
      await page.getByRole('button', { name: '予約する' }).click();

      await expectToast(page, '予約に失敗しました');
    });
  });
});
