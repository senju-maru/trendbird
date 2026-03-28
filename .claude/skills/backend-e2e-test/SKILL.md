---
name: backend-e2e-test
description: バックエンド E2E テストスキル。Go + Connect RPC + 実DB構成のE2Eテスト設計・実装パターン。バックエンドのE2Eテスト追加・修正時に自動参照される。
---

# バックエンド E2E テストスキル v1.0

バックエンド E2E テストの設計・実装パターンを体系化したスキル。

---

## セクション 1: コンセプト

- **E2E ファースト方針**: `backend-architecture` セクション 6.1 と同期。単体テストよりも E2E テストを優先する
- **統合テスト範囲**: Connect client → handler → usecase → repository → PostgreSQL（実 DB）
- **Docker/testcontainers 不使用**: CLAUDE.md ルール準拠。ローカル PostgreSQL を使用
- **テスト用 DSN**: `postgres://localhost:5432/trendbird_test?sslmode=disable`
- **環境変数 `TEST_DATABASE_URL`** で上書き可能（GitHub Actions では Docker postgres Service Container を使用）

---

## セクション 2: ディレクトリ構成

```
backend/internal/e2etest/
├── testmain_test.go          # TestMain: DB接続 + マイグレーション
├── testhelper_test.go         # testEnv + seed + assertion + auth ヘルパー
├── mock_gateway_test.go       # 手書きモックゲートウェイ
├── auth_test.go               # AuthService テスト
├── topic_test.go              # TopicService テスト
├── dashboard_test.go          # DashboardService テスト
├── post_test.go               # PostService テスト
├── auto_dm_test.go            # AutoDMService テスト
└── scheduled_publish_test.go  # バッチジョブテスト
```

---

## セクション 3: テストインフラ（testEnv + setupTest）

### testEnv 構造体

```go
type testEnv struct {
    server *httptest.Server
    db     *gorm.DB

    // Connect RPC clients (認証なし)
    authClient         trendbirdv1connect.AuthServiceClient
    dashboardClient    trendbirdv1connect.DashboardServiceClient
    topicClient        trendbirdv1connect.TopicServiceClient
    postClient         trendbirdv1connect.PostServiceClient
    autoDMClient       trendbirdv1connect.AutoDMServiceClient
    // ... 他のサービスクライアント

    // Mocks
    mockTwitter *mockTwitterGateway
    mockAI      *mockAIGateway
}
```

### setupTest(t)

1. `truncateAll(t, testDB)` — 全テーブル CASCADE TRUNCATE
2. モックゲートウェイ生成（デフォルト成功レスポンス）
3. `di.NewContainerForTest(deps)` — DI コンテナ構築
4. `router.New(container)` → `httptest.NewServer(handler)` — テストサーバー起動
5. 認証なし RPC クライアント群を生成

### truncateAll

FK 依存順で全テーブルを TRUNCATE CASCADE。テーブルリストはマイグレーション追加時に更新が必要。

---

## セクション 4: Seed ヘルパーパターン（Options Pattern）

### 命名規則

- **関数名**: `seed<Entity>(t, db, [親ID,] opts ...<entity>Option) *model.<Entity>`
- **オプション型**: `type <entity>Option func(*model.<Entity>)`
- **オプション名**: `with<Field>(value)` → `func(*model.Entity)` を返す
- **ユニーク値**: `nextSeq()` atomic カウンターで `fmt.Sprintf("Value %d", n)`

### デフォルト値

各 seed 関数が妥当なデフォルトを提供する。テストごとの最小設定で動くように。

### 既存 seed 一覧

| 関数 | 親ID引数 | 主要オプション |
|------|----------|--------------|
| `seedUser` | - | `withEmail`, `withTwitterID`, `withTwitterHandle`, `withName`, `withImage` |
| `seedTopic` | - | `withTopicName`, `withTopicStatus`, `withTopicGenre`, `withTopicKeywords`, `withTopicZScore`, `withTopicCurrentVolume`, `withTopicBaselineVolume`, `withTopicChangePercent`, `withTopicContext`, `withTopicContextSummary`, `withTopicSpikeStartedAt` |
| `seedUserTopic` | userID, topicID | `withNotificationEnabled`, `withIsCreator` |
| `seedTopicVolume` | topicID | `withTopicVolumeTimestamp`, `withTopicVolumeValue` |
| `seedSourcePost` | topicID | `withSourcePostTweetID`, `withSourcePostContent`, `withSourcePostLikes`, `withSourcePostRetweets`, `withSourcePostViews`, `withSourcePostAuthorHandle`, `withSourcePostAuthorName`, `withSourcePostReplies`, `withSourcePostCollectedAt` |
| `seedActivity` | userID | `withActivityType`, `withActivityDescription` |
| `seedSpikeHistory` | topicID | `withSpikePeakZScore`, `withSpikeStatus`, `withSpikeSummary`, `withSpikeTimestamp`, `withSpikeDurationMinutes` |
| `seedAIGenerationLog` | userID | `withAIGenStyle`, `withAIGenCount` |
| `seedPost` | userID | `withPostStatus`, `withPostContent`, `withPostTopicID`, `withPostTopicName`, `withPostPublishedAt`, `withPostScheduledAt`, `withPostTweetURL` |
| `seedNotification` | userID | `withNotificationType`, `withNotificationRead` |
| `seedNotificationSetting` | userID | `withSpikeEnabled`, `withRisingEnabled`, `withEmailEnabled` |
| `seedTwitterConnection` | userID | `withTwitterConnStatus`, `withAccessToken`, `withTokenExpiresAt` |
| `seedAutoDMRule` | userID | `withRuleEnabled`, `withRuleTriggerKeywords`, `withRuleTemplateMessage` |
| `seedDMSentLog` | userID, ruleID | `withDMLogRuleID`, `withDMLogRecipientID`, `withDMLogTriggerKeyword` |
| `ensureGenre` | slug | - |
| `seedUserGenre` | userID, genreSlug | - |

### bool フィールドの罠

GORM Create は `false` をゼロ値としてスキップする。`seedNotificationSetting` のように raw SQL INSERT が必要なケースがある。

---

## セクション 5: 認証パターン

```go
// 認証付きクライアント生成（ジェネリクス）
client := connectClient(t, env, userID, trendbirdv1connect.NewXxxServiceClient)

// カスタムトークンでエッジケーステスト
token := generateCustomToken(t, signingKey, claims, signingMethod)

// Cookie フォールバック認証
client := trendbirdv1connect.NewAuthServiceClient(
    env.server.Client(), env.server.URL,
    connect.WithInterceptors(cookieInterceptor("tb_jwt="+token)),
)
```

---

## セクション 6: アサーションパターン

```go
// Connect RPC エラーコード
assertConnectCode(t, err, connect.CodeNotFound)

// フィールド比較
if got.GetField() != want {
    t.Errorf("field: want %q, got %q", want, got.GetField())
}

// Proto 全体比較
if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
    t.Errorf("mismatch (-want +got):\n%s", diff)
}

// DB 検証
var count int64
db.Model(&model.Xxx{}).Where("user_id = ?", userID).Count(&count)

// Cookie 検証
findSetCookie(header, name)
assertSetCookieCleared(t, header, name)

// 時刻 RFC3339 パース
if _, err := time.Parse(time.RFC3339, got.GetTimestamp()); err != nil { ... }
```

---

## セクション 7: テストケース設計チェックリスト

| # | カテゴリ | 必須 | 説明 |
|---|---------|------|------|
| 1 | success | YES | 正常系。レスポンス全フィールド + DB 副作用検証 |
| 2 | unauthenticated | YES | `env.xxxClient`（認証なし）→ CodeUnauthenticated |
| 3 | not_found | YES | 存在しない ID → CodeNotFound |
| 4 | permission_denied | YES | 他ユーザーのリソース → CodeNotFound（情報漏洩防止） |
| 5 | cross_user_isolation | YES | userA/B のデータが混在しないこと |
| 6 | empty | YES | データ0件時に空配列（nil ではない）を返すこと |
| 7 | limit/pagination | 該当時 | 上限件数到達時の切り捨て、ページネーション境界 |
| 9 | ordering | 該当時 | DESC/ASC 順序の検証（timestamp, engagement 等） |
| 10 | all_fields_validation | 該当時 | Proto 全フィールドの値一致を1テストで網羅 |
| 11 | optional fields | 該当時 | nil/未設定の optional フィールドのハンドリング |
| 12 | boundary values | 該当時 | 数値の境界値、月初/月末、max int32 等 |
| 13 | shared resources | 該当時 | 共有トピック等、複数ユーザーが同一リソースを参照 |
| 14 | mock 差し替え | 該当時 | 外部 API 失敗時の挙動 |
| 15 | DB 副作用 | 該当時 | mutation 後の DB 状態検証（Create/Update/Delete） |

---

## セクション 8: 時間関連テストパターン

```go
// タイムスタンプ順序保証（seed ループ内）
time.Sleep(10 * time.Millisecond)

// 月境界テスト用 raw SQL
env.db.Exec("UPDATE spike_histories SET created_at = ? WHERE id = ?", prevMonth, id)

// RFC3339 パース + 許容差
ts, _ := time.Parse(time.RFC3339, got.GetTimestamp())
if ts.Sub(want).Abs() > 2*time.Second { ... }

// 時間範囲テスト
time.Now().AddDate(0, 0, -7) // 7日前
```

---

## セクション 9: モックゲートウェイパターン

### 構造

```go
type mockXxxGateway struct {
    FunctionFn func(ctx context.Context, ...) (..., error)  // 差し替え可能な関数
    Calls      atomic.Int64                                  // 呼び出しカウンター
    mu         sync.Mutex                                    // 入力キャプチャ用
    LastInput  *SomeInput                                    // 最後の入力
}
```

### パターン

- **デフォルト**: 成功レスポンスを返す `newMockXxxGateway()`
- **テスト内差し替え**: `env.mockXxx.FunctionFn = func(...) { ... }`
- **呼び出し回数**: `Calls.Load()` 前後差分で検証
- **入力キャプチャ**: `mu.Lock()` + `LastInput` フィールド

### 4つのモック

| モック | インターフェース | 主要メソッド |
|--------|----------------|------------|
| `mockTwitterGateway` | `gateway.TwitterGateway` | BuildAuthorizationURL, ExchangeCode, GetUserInfo, PostTweet, etc. |
| `mockAIGateway` | `gateway.AIGateway` | GeneratePosts |

---

## セクション 10: テスト命名規約

```go
// トップレベル: TestServiceName_MethodName
func TestTopicService_GetTopic(t *testing.T) {
    // サブテスト: snake_case_scenario
    t.Run("success", func(t *testing.T) { ... })
    t.Run("spike_status_with_all_fields", func(t *testing.T) { ... })
    t.Run("unauthenticated", func(t *testing.T) { ... })
}
```

**サブテスト順序**: success → success variants → error codes → edge cases → unauthenticated

---

## セクション 11: コード例テンプレート

### 読み取り RPC テスト（GetXxx / ListXxx）

```go
t.Run("success", func(t *testing.T) {
    env := setupTest(t)
    user := seedUser(t, env.db)
    // seed test data
    entity := seedXxx(t, env.db, user.ID, withXxxField("value"))

    client := connectClient(t, env, user.ID, trendbirdv1connect.NewXxxServiceClient)
    resp, err := client.GetXxx(context.Background(), connect.NewRequest(&trendbirdv1.GetXxxRequest{
        Id: entity.ID,
    }))
    if err != nil {
        t.Fatalf("GetXxx: %v", err)
    }

    got := resp.Msg.GetXxx()
    if got.GetField() != "value" {
        t.Errorf("field: want %q, got %q", "value", got.GetField())
    }
})
```

### 書き込み RPC テスト（Create/Update/Delete + DB 副作用検証）

```go
t.Run("success", func(t *testing.T) {
    env := setupTest(t)
    user := seedUser(t, env.db)

    client := connectClient(t, env, user.ID, trendbirdv1connect.NewXxxServiceClient)
    resp, err := client.CreateXxx(context.Background(), connect.NewRequest(&trendbirdv1.CreateXxxRequest{
        Name: "Test",
    }))
    if err != nil {
        t.Fatalf("CreateXxx: %v", err)
    }

    // レスポンス検証
    if resp.Msg.GetXxx().GetName() != "Test" {
        t.Errorf("name: want Test, got %s", resp.Msg.GetXxx().GetName())
    }

    // DB 副作用検証
    var dbEntity model.Xxx
    if err := env.db.First(&dbEntity, "id = ?", resp.Msg.GetXxx().GetId()).Error; err != nil {
        t.Fatalf("DB fetch: %v", err)
    }
})
```


---

## セクション 12: seed ヘルパー追加手順

1. `model.Xxx` の全フィールドを確認
2. FK 依存がある場合は親 entity の ID を引数に含める
3. `nextSeq()` でユニーク値を生成
4. `with<Field>` オプション関数を定義
5. `db.Create` のエラーは `t.Fatalf` で即座に失敗
6. **bool フィールドのゼロ値問題を確認**（必要なら raw SQL）

---

## セクション 13: テスト実行コマンド

```bash
# 全 E2E テスト
cd backend && go test ./internal/e2etest/... -v -race -count=1

# 特定サービス
cd backend && go test ./internal/e2etest/... -run TestDashboardService -v -race -count=1

# 特定メソッド
cd backend && go test ./internal/e2etest/... -run TestDashboardService_GetStats -v -race -count=1
```

---

## セクション 14: 他スキルとの関連

| 関心事 | 本スキル | 委譲先 |
|---|---|---|
| クリーンアーキテクチャ方針 | テスト対象の理解 | `/backend-architecture` |
| ドメインルール（z-score） | テストシナリオ設計 | ソースコード参照 |
| Proto 定義・コメント | テストの期待値設計 | `/protobuf-style` |
| フロントエンド E2E | 役割分担（バックエンド=実DB、FE=モック API） | `/frontend-e2e-test` |
