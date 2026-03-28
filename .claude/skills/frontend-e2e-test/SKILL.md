---
name: frontend-e2e-test
description: フロントエンド E2E テストスキル。Playwright + page.route() ベースのAPIモック戦略。フロントエンドのE2Eテスト追加・修正時に自動参照される。
---

# フロントエンド E2E テストスキル v3.0

Playwright + `page.route()` + Proto Schema ベースモック。**このファイルだけ読めばテストが書ける。**

---

## 1. Quick Start テンプレート

```typescript
import { test, expect, authenticateContext } from '../fixtures/test-base';
import { resetSeq } from '../fixtures/factories';

test.describe('Feature', () => {
  test.beforeEach(async ({ context, page, apiMock }) => {
    resetSeq();
    await authenticateContext(context);       // tb_jwt Cookie 設定
    await apiMock.setupDefaults();
    await page.addInitScript(() => {
      sessionStorage.removeItem('tb_tutorial_pending');
    });
  });

  test('要素が表示される', async ({ page }) => {
    // ...
  });
});
```

**実行コマンド:**
```bash
npm run e2e                          # 全テスト
npm run e2e -- --grep "ログイン"      # フィルタ実行
```

**ApiMock 3メソッド:**

| メソッド | 用途 | シグネチャ |
|---|---|---|
| `mockRPC` | 正常レスポンス | `(service, method, body, status=200, headers?)` |
| `mockRPCError` | エラーレスポンス | `(service, method, code, message)` — code: `not_found`, `unauthenticated`, `permission_denied`, `resource_exhausted`, `invalid_argument`, `internal`, `already_exists`, `failed_precondition` |
| `clearMock` | setupDefaults のオーバーライド解除 | `(service, method)` — `clearMock` → `mockRPC` の順で再設定 |

**Response 構築3パターン:**

```typescript
// パターン1: Flat — toJson + create
import { create, toJson } from '@bufbuild/protobuf';
import { ListTopicsResponseSchema } from '../../src/gen/trendbird/v1/topic_pb';
toJson(ListTopicsResponseSchema, create(ListTopicsResponseSchema, { topics: [...] }))

// パターン2: Wrapped — ネストした create
import { GetStatsResponseSchema, DashboardStatsSchema } from '../../src/gen/trendbird/v1/dashboard_pb';
toJson(GetStatsResponseSchema, create(GetStatsResponseSchema, {
  stats: create(DashboardStatsSchema, { totalTopics: 5 }),
}))

// パターン3: Set-Cookie 付き（ログアウト・退会）
await apiMock.mockRPC('AuthService', 'Logout', {}, 200, {
  'Set-Cookie': 'tb_jwt=; HttpOnly; SameSite=Lax; Path=/; Max-Age=0',
});
```

---

## 2. setupDefaults モック一覧テーブル

`setupDefaults()` が登録する **14 RPC**。テスト中に `clearMock` → `mockRPC` でオーバーライド可能。

| # | Service | Method | Response Field(s) | 構造 | Default 値の要点 |
|---|---|---|---|---|---|
| 1 | AuthService | GetCurrentUser | `user`, `aiGenerationUsed` | Wrapped | user: `{id:'default-user', name:'Default User'}`, aiGenerationUsed: `0` |
| 2 | TopicService | ListTopics | `topics` | Array | 2件: SPIKE(id:`default-topic-1`, zScore:3.5) + STABLE(id:`default-topic-2`, zScore:1.2) |
| 3 | TopicService | ListGenres | `genres` | Array | 2件: technology + business (GenreSchema: slug/label) |
| 4 | TopicService | ListUserGenres | `genres` | **string[]** | `['technology']` — **Genre[] ではなく string[]** |
| 5 | PostService | ListDrafts | `drafts` | Array | 空 (`{}`) — フィールド名は `drafts`（`posts` ではない） |
| 6 | PostService | ListPostHistory | (empty) | Array | 空 (`{}`) |
| 7 | PostService | GetPostStats | `stats` | Wrapped | `stats: create(PostStatsSchema, {})` — 全フィールドゼロ |
| 8 | DashboardService | GetActivities | `activities` | Array | 空 (`{}`) |
| 9 | DashboardService | GetStats | `stats` | Wrapped | `stats: create(DashboardStatsSchema, {})` — 全フィールドゼロ |
| 10 | NotificationService | ListNotifications | `notifications` | Array | 空 (`{}`) |
| 11 | SettingsService | GetNotificationSettings | `settings` | Wrapped | `settings: {spikeEnabled:true, risingEnabled:true}` |
| 13 | TwitterService | GetConnectionInfo | `info` | Wrapped | `info: {status: DISCONNECTED}` |
| 14 | AutoDMService | ListAutoDMRules | `rules` | Array | 空 (`{}`) |
| 15 | AutoDMService | GetDMSentLogs | `logs` | Array | 空 (`{}`) |

### setupDefaults にない頻出 RPC

テスト内で `mockRPC` / `mockRPCError` で個別に設定する RPC。

| Service | Method | Response Field(s) | 典型的な使い方 |
|---|---|---|---|
| AuthService | XAuth | `{}` + Set-Cookie `tb_jwt=xxx` | OAuth コールバックテスト |
| AuthService | Logout | `{}` + Set-Cookie `tb_jwt=; Max-Age=0` | ログアウト（JWT クリア必須） |
| AuthService | DeleteAccount | `{}` + Set-Cookie `tb_jwt=; Max-Age=0` | 退会（JWT クリア必須） |
| TopicService | SuggestTopics | `suggestions` (TopicSuggestion[]) | ジャンル選択後のおすすめ |
| TopicService | CreateTopic | `{ topic }` | トピック追加 |
| TopicService | DeleteTopic | `{}` | トピック削除 |
| TopicService | AddGenre | `{}` | ジャンル追加 |
| TopicService | RemoveGenre | `{}` | ジャンル削除 |
| TopicService | UpdateTopicNotification | `{}` | 通知トグル |
| PostService | GeneratePosts | `posts` (GeneratedPost[]) | AI 生成 |
| PostService | CreateDraft | `{ draft }` | 下書き保存 |
| PostService | UpdateDraft | `{ draft }` | 下書き編集 |
| PostService | DeleteDraft | `{}` | 下書き削除 |
| PostService | PublishPost | `{}` | 即時投稿 |
| PostService | SchedulePost | `{}` | 予約投稿 |
| SettingsService | UpdateProfile | `{}` | プロフィール保存 |
| SettingsService | UpdateNotifications | `{ updated: true }` | 通知設定保存 |
| TwitterService | DisconnectTwitter | `{}` | X 連携解除 |
| DashboardService | GetTrendingPosts | `posts` (TrendingPost[]) | トピック詳細の注目ポスト |
| NotificationService | MarkAsRead | `{}` | 個別既読 |
| NotificationService | MarkAllAsRead | `{}` | 一括既読 |
| AutoDMService | CreateAutoDMRule | `{ rule }` | ルール作成 |
| AutoDMService | UpdateAutoDMRule | `{ rule }` | ルール更新（トグル含む） |
| AutoDMService | DeleteAutoDMRule | `{}` | ルール削除 |

---

## 3. ファクトリリファレンステーブル

`factories.ts` の全ビルダー。すべて `toJson(Schema, create(Schema, { defaults + overrides }))` パターン（※印は `create()` のみで `toJson()` しない）。

| 関数 | Proto Schema | Override Fields | Key Defaults |
|---|---|---|---|
| `buildUser` | UserSchema | id, name, email, image, twitterHandle, createdAt | id: `user-{n}` |
| `buildTopic` | TopicSchema | id, name, keywords, genre, status, changePercent, zScore, currentVolume, baselineVolume, context, contextSummary, spikeStartedAt, notificationEnabled, createdAt | status: `STABLE`, genre: `'technology'`, zScore: `1.2` |
| `buildGenre` | GenreSchema | (positional: slug, label) | sortOrder: `0` |
| `buildAutoDMRule` | AutoDMRuleSchema | id, enabled, triggerKeywords, templateMessage, createdAt, updatedAt | enabled: `true` |
| `buildDMSentLog` | DMSentLogSchema | id, recipientTwitterId, replyTweetId, triggerKeyword, dmText, sentAt | — |
| `buildTrendingPost` | TrendingPostSchema | id, authorHandle, authorName, content, likes, retweets, replies, views | likes: `100*n`, views: `10000*n` |
| `buildGeneratedPost` | GeneratedPostSchema | id, style, styleLabel, styleIcon, content, topicId, sourcePostIds | style: `CASUAL`, topicId: `'default-topic-1'` |
| `buildDraft` | ScheduledPostSchema | id, content, topicId, topicName, status, scheduledAt, createdAt, updatedAt, characterCount, errorMessage, failedAt | status: `DRAFT`, characterCount: auto |
| `buildScheduledPost` | ScheduledPostSchema | id, content, topicId, topicName, scheduledAt, createdAt, updatedAt, characterCount | status: `SCHEDULED` |
| `buildPostHistory` | PostHistorySchema | id, content, topicId, topicName, publishedAt, likes, retweets, replies, views, tweetUrl | — |
| `buildTopicSuggestion` | TopicSuggestionSchema | id, name, keywords, genre, genreLabel, similarityScore | similarityScore: `0.85` |
| `buildNotification` | NotificationSchema | id, type, title, message, timestamp, isRead, topicId, topicName, topicStatus, actionUrl, actionLabel | type: `TREND`, isRead: `false` |
| `buildActivity` | ActivitySchema | id, type, topicName, description, timestamp | type: `SPIKE` |
| `buildSpikeHistoryEntry` ※ | SpikeHistoryEntrySchema | id, timestamp, peakZScore, status, summary, durationMinutes | peakZScore: `4.2`, **returns `create()` not `toJson()`** |
| `buildSparklineDataPoint` ※ | SparklineDataPointSchema | (positional: timestamp, value) | **returns `create()` not `toJson()`** |
| `buildPostingTips` ※ | PostingTipsSchema | peakDays, peakHoursStart, peakHoursEnd, nextSuggestedTime | peakDays: `['月','水','金']`, **returns `create()` not `toJson()`** |

**Re-exported Enums:**

| Enum | 値 | 元モジュール |
|---|---|---|
| `TopicStatus` | STABLE=0, RISING=1, SPIKE=2 | topic_pb |
| `ActivityType` | SPIKE=0, RISING=1, AI_GENERATION=2, POST_PUBLISHED=3 | dashboard_pb |
| `PostStyle` | CASUAL=0, PROFESSIONAL=1, HUMOROUS=2 | post_pb |
| `PostStatus` | DRAFT=0, SCHEDULED=1, PUBLISHED=2, FAILED=3 | post_pb |
| `NotificationType` | TREND=0, SYSTEM=1 | notification_pb |

---

## 4. Response Schema 構築パターン

### Flat（配列レスポンス）
```typescript
import { create, toJson } from '@bufbuild/protobuf';
import { ListTopicsResponseSchema } from '../../src/gen/trendbird/v1/topic_pb';
import { buildTopic } from '../fixtures/factories';

const topic = buildTopic({ status: TopicStatus.SPIKE, zScore: 5.0 });
await apiMock.mockRPC('TopicService', 'ListTopics', { topics: [topic] });
```

### Wrapped（ネストオブジェクト）
```typescript
import { GetStatsResponseSchema, DashboardStatsSchema } from '../../src/gen/trendbird/v1/dashboard_pb';

await apiMock.mockRPC('DashboardService', 'GetStats',
  toJson(GetStatsResponseSchema, create(GetStatsResponseSchema, {
    stats: create(DashboardStatsSchema, { totalTopics: 5, spikeCount: 2 }),
  })));
```

### Set-Cookie 付き（JWT クリア）
```typescript
const CLEAR_JWT = { 'Set-Cookie': 'tb_jwt=; HttpOnly; SameSite=Lax; Path=/; Max-Age=0' };
await apiMock.mockRPC('AuthService', 'Logout', {}, 200, CLEAR_JWT);
await apiMock.mockRPC('AuthService', 'DeleteAccount', {}, 200, CLEAR_JWT);
```

### ファクトリ出力をそのまま使う（最も頻出）
```typescript
// buildXxx() は toJson() 済みの plain object を返す → そのまま渡せる
const rule = buildAutoDMRule({ enabled: true });
await apiMock.mockRPC('AutoDMService', 'ListAutoDMRules', { rules: [rule] });

// ※印ファクトリ（create()のみ）は Response Schema に埋め込んで toJson する
const entry = buildSpikeHistoryEntry({ peakZScore: 6.0 });
const topic = buildTopic(); // ← 注: spikeHistory フィールドは TopicSchema 内なので別途構築
```

---

## 5. POM クイックリファレンステーブル

| クラス | import | goto() | 遷移方式 | 主要ロケータ/メソッド |
|---|---|---|---|---|
| `LoginPage` | `login.page` | `/login` 直接 | Direct | `heading`, `xLoginButton`, `loggedOutMessage`, `guidanceText`, `subtitle` |
| `DashboardPage` | `dashboard.page` | `/dashboard` 直接 | Direct | `statusBar`, `spikeCountText`, `risingCountText`, `genreTabs`, `genreTab(label)`, `topicCardByName(name)`, `gotoDetail(topicName)`, `backButton`, `topicHeading`, `detailHero`, `contextCard`, `aiSection`, `aiGenerateButton`, `pricingLink`, `csvDownloadButton`, `notificationToggle`, `topicNotFound`, `calmText`, `generatingText`, `downloadError` |
| `TopicsPage` | `topics.page` | `/topics` 直接 | Direct | `addGenreButton`, `genreChip(label)`, `genreModal`, `genreOption(label)`, `browseByGenreTab`, `searchTopicsTab`, `topicSearchInput`, `topicLimitInput`, `recommendedSection`, `selectedTopicsSection`, `genreProgress(text)`, `emptyGenreState`, `noSuggestions`, `noSearchResults` |
| `PostsPage` | `posts.page` | `/posts` 直接 | Direct | `topicCard(name)`, `statusFilterButton(label)`, `genreFilterButton(label)`, `generateButton`, `generatingButton`, `upgradeButton`, `aiResultsHeader`, `remainingGenerations`, `composerTextarea`, `charCounter`, `saveDraftButton`, `scheduleButton`, `publishButton`, `xDisconnectedLabel`, `managementHeader`, `draftsTab`, `scheduledTab`, `historyTab`, `noDrafts`, `noScheduled`, `noHistory`, `selectTopicPrompt`, `editModal`, `scheduleModal`, `deleteDialog`, `publishDialog` |
| `NotificationsPage` | `notifications.page` | `/notifications` 直接 | Direct | `heading`, `markAllReadButton`, `allTab`, `trendTab`, `systemTab`, `tab(label)`, `notificationByTitle(title)`, `emptyState`, `errorState`, `retryButton` |
| `SettingsPage` | `settings.page` | `/settings?tab=xxx` 直接 | Direct | `switchTab(tab)`, `heading`, `saveButton`, `notificationToggle(label)`, `connectButton`, `reconnectButton`, `disconnectButton`, `logoutButton`, `deleteAccountButton` |
| `AutoDMPage` | `auto-dm.page` | **Dashboard 経由** | Via navigation | `heading`, `settingsTab`, `historyTab`, `addRuleButton`, `emptyState`, `toggleButton(index)`, `saveButton(label)`, `keywordInput`, `templateInput`, `deleteButton` |

**Dashboard 経由ページの遷移ルール:**
- `AutoDMPage.goto()` は `/dashboard` → `a[href="/auto-dm"]` クリックで遷移（Zustand auth state を維持）
- `page.goto('/auto-dm')` 直接はNG — Zustand store がリセットされるため

---

## 6. パターン & 注意事項

### Proto3 JSON のデフォルト値省略
```typescript
// false/0/""/"[]" は toJson() で省略される → undefined で返る
const body = JSON.parse(req.postData() ?? '{}');
expect(body.enabled ?? false).toBe(false);  // undefined or false を許容
expect(body.enabled).toBe(true);            // true は省略されないので直接チェック
```

### チュートリアル干渉防止
`beforeEach` で必ず `sessionStorage.removeItem('tb_tutorial_pending')` を実行。

### authRedirectInterceptor
`transport.ts` で `Code.Unauthenticated` → `/login` 自動リダイレクト。**setupDefaults にない RPC が呼ばれると localhost:8080 に実リクエスト → 予期せぬ挙動が発生。** 新画面追加時は全 RPC がモック済みか確認。

### clearMock → mockRPC でオーバーライド
setupDefaults のレスポンスを変えたい場合:
```typescript
await apiMock.clearMock('TopicService', 'ListTopics');
await apiMock.mockRPC('TopicService', 'ListTopics', { topics: [customTopic] });
```

### カスタム UI コンポーネント
- `TabsTrigger` は `<button>` で `role="tab"` がない → `getByRole('button', { name: /テキスト/ })` を使用
- `Toggle` は `data-testid` で特定 → `page.locator('[data-testid="xxx"]').locator('button')`
- nav と main に同名ボタンがある場合 → `getByRole('main').getByRole('button', ...)` でスコープを絞る

### data-testid 追加戦略
テスト作成時にインクリメンタルに追加。命名規則: `{page}-{element}` (例: `auto-dm-toggle-0`)

---

## 7. テスト追加チェックリスト

1. **RPC モック漏れ:** 画面が呼ぶ全 RPC が setupDefaults にあるか §2 テーブルで確認。なければ `mockRPC` で追加
2. **Dashboard 経由ページ:** §5 の遷移方式が「Via navigation」なら dashboard 経由。`page.goto()` 直接は不可
3. **Proto3 JSON:** リクエストボディのアサーションでデフォルト値省略を考慮（§6）
4. **Set-Cookie:** ログアウト・退会テストでは JWT クリアヘッダー必須（§4）
5. **fixtures/POM を変更したら、このスキルのテーブル（§2, §3, §5）も更新する**
