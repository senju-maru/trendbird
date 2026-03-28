---
name: x-api
description: TrendBirdバックエンドにおけるX(Twitter) API v2連携ガイド。認証・投稿・検索・トレンド取得・レート制限の仕様リファレンス。バックエンドでX API連携コードを書く際に自動参照される。
---

# TrendBird X(Twitter) API v2 連携ガイド v1.0

**Go + net/http + OAuth 2.0**
最終更新: 2026-02-17

---

## 1. 概要・目的

TrendBird は X(Twitter) のトレンド検知 + AI 投稿文自動生成の SaaS であり、バックエンドから X API v2 を利用して以下を実現する:

1. **トレンド取得** — 特定地域（日本 WOEID: `23424856`）のトレンドを定期的に取得
2. **ツイート検索** — トレンドに関連するツイートを検索し、文脈を分析
3. **ツイートカウント** — トレンドの盛り上がり推移を時系列で取得
4. **ツイート投稿** — AI が生成した投稿文をユーザーの代理で投稿
5. **ユーザー情報取得** — 連携ユーザーのプロフィール・タイムラインを取得

**推奨ワークフロー:**

```
トレンド取得 → ツイート検索（文脈分析） → AI 投稿文生成 → ツイート投稿
     ↑                                                        ↓
   定期実行（cron）                               ユーザー承認後に実行
```

**ベース URL:** `https://api.x.com`

**Clean Architecture における位置付け:**

```
domain/gateway/          twitter.go（TwitterClient interface）
usecase/                 twitter.go（トレンド取得・投稿のユースケース）
infrastructure/external/ twitter.go（TwitterClient 実装）
```

---

## 2. 認証方式

### 認証キー一覧と使い分け

X Developer Console で取得できる 3 種類の認証キーの比較:

| 認証方式 | キー | 用途 | ユーザーコンテキスト | トークン有効期限 |
|---|---|---|---|---|
| OAuth 1.0a | Consumer Key/Secret + Access Token/Secret | ユーザー代理操作（メディアアップロード等のレガシー API で必須） | あり | 無期限 |
| OAuth 2.0 PKCE | Client ID / Client Secret | ユーザー代理操作（**TrendBird 推奨**） | あり | Access Token: 2 時間（Refresh Token で更新可） |
| Bearer Token | Bearer Token | 公開データ読み取り専用 | なし | 無期限 |

**エンドポイント別の認証方式対応:**

| エンドポイント | OAuth 2.0 PKCE | OAuth 1.0a | Bearer Token |
|---|---|---|---|
| `POST /2/tweets`（投稿） | ✅ | ✅ | ❌ |
| `DELETE /2/tweets/{id}` | ✅ | ✅ | ❌ |
| `GET /2/tweets/search/recent` | ✅ | ✅ | ✅ |
| `GET /2/tweets/counts/recent` | ❌ | ❌ | ✅ |
| `GET /2/trends/by/woeid/{woeid}` | ❌ | ❌ | ✅ |
| `GET /2/users/personalized_trends` | ✅ | ❌ | ❌ |
| `GET /2/users/me` | ✅ | ✅ | ❌ |
| `GET /2/users/{id}` | ✅ | ✅ | ✅ |
| `POST /2/media/upload`（メディア） | ❌ | ✅ | ❌ |

> **TrendBird の方針:** 基本は **OAuth 2.0 PKCE** を使用。メディア付き投稿が必要な場合のみ OAuth 1.0a を併用する。公開データの読み取り（トレンド取得・ツイートカウント等）は **Bearer Token** を使用。

### 2.1 OAuth 2.0 Authorization Code Flow with PKCE（ユーザーコンテキスト）

ユーザーの代理でツイート投稿・削除を行う場合に使用する。TrendBird の主要認証方式。

**認可フロー:**

```
User          Frontend         Backend                    X API
 │              │                │                          │
 │  X連携開始   │                │                          │
 │─────────────▶│  GET /auth/x   │                          │
 │              │───────────────▶│  code_verifier 生成      │
 │              │                │  code_challenge = S256(verifier)
 │              │  ← redirect    │                          │
 │              │◀───────────────│                          │
 │  リダイレクト │                │                          │
 │─────────────▶│═══ X 認可画面 ═══════════════════════════▶│
 │              │  ユーザーが許可                           │
 │              │◀═══ callback?code=xxx&state=yyy ══════════│
 │              │  POST /auth/x/callback                   │
 │              │───────────────▶│  POST /2/oauth2/token    │
 │              │                │─────────────────────────▶│
 │              │                │  ← access_token          │
 │              │                │    refresh_token         │
 │              │                │  → DB に保存             │
 │              │  ← 認証完了    │                          │
 │              │◀───────────────│                          │
```

**認可 URL:** `https://x.com/i/oauth2/authorize`

| パラメータ | 値 |
|------------|-----|
| `response_type` | `code` |
| `client_id` | アプリの Client ID |
| `redirect_uri` | 登録済みのコールバック URL と完全一致 |
| `scope` | スペース区切りのスコープ |
| `state` | CSRF 対策用のランダム文字列 |
| `code_challenge` | `code_verifier` の SHA-256 ハッシュ（Base64URL エンコード） |
| `code_challenge_method` | `S256` |

**トークン交換:** `POST https://api.x.com/2/oauth2/token`

- Content-Type: `application/x-www-form-urlencoded`
- パラメータ: `code`, `grant_type=authorization_code`, `client_id`, `redirect_uri`, `code_verifier`

**トークン有効期限:**

| トークン | 有効期限 |
|----------|----------|
| Authorization code | **30 秒**（即座に交換すること） |
| Access token | **2 時間** |
| Refresh token | 無期限（`offline.access` スコープ必須） |

**リフレッシュ:** `POST https://api.x.com/2/oauth2/token` に `grant_type=refresh_token` と `refresh_token` を送信。

### 2.2 App-only Bearer Token（アプリコンテキスト）

公開データの読み取り専用。ユーザーコンテキスト不要のエンドポイントで使用。

- 取得: `POST https://api.x.com/oauth2/token` に `grant_type=client_credentials`
- 用途: 検索、ユーザー情報取得、トレンド取得、ツイートカウント

### 2.3 TrendBird で必要なスコープ

| スコープ | 用途 |
|----------|------|
| `tweet.read` | ツイート検索・取得 |
| `tweet.write` | ツイート投稿・削除 |
| `users.read` | ユーザー情報取得 |
| `users.email` | ユーザーメールアドレス取得（アカウント登録・通知用） |
| `offline.access` | リフレッシュトークン発行（長期アクセス） |

**全スコープ一覧（参考）:**

| スコープ | 説明 |
|----------|------|
| `tweet.read` | ツイートの読み取り |
| `tweet.write` | ツイートの投稿・削除 |
| `tweet.moderate.write` | リプライの非表示 |
| `users.read` | ユーザープロフィールの読み取り |
| `users.email` | ユーザーメールの読み取り |
| `follows.read` | フォロー関係の読み取り |
| `follows.write` | フォロー・アンフォロー |
| `offline.access` | リフレッシュトークンの発行 |
| `like.read` / `like.write` | いいねの読み取り・実行 |
| `bookmark.read` / `bookmark.write` | ブックマークの読み取り・作成 |
| `block.read` / `block.write` | ブロックの読み取り・実行 |
| `mute.read` / `mute.write` | ミュートの読み取り・実行 |
| `list.read` / `list.write` | リストの読み取り・管理 |
| `dm.read` / `dm.write` | DM の読み取り・送信 |
| `media.write` | メディアのアップロード |
| `space.read` | スペースの読み取り |

### 2.4 メールアドレス取得

TrendBird ではユーザー登録時に X アカウントのメールアドレスを取得し、アカウント識別に使用する。

#### Developer Portal での必須設定

メールアドレスを取得するには、X Developer Portal でアプリに以下の設定が必要:

1. **Privacy Policy URL** — 公開済みのプライバシーポリシーページ URL
2. **Terms of Service URL** — 公開済みの利用規約ページ URL
3. **Website URL** — アプリの公式サイト URL
4. **Request email from users** チェックボックス — 有効化が必須

> **重要:** これらの設定が不足していると、`users.email` スコープを指定しても認可画面にメール許可のチェックボックスが表示されず、メールアドレスを取得できない。

#### 取得方法（OAuth 2.0 PKCE）

1. 認可リクエストのスコープに `users.email` を含める
2. ユーザーが認可画面でメールアドレスの共有を許可する
3. アクセストークン取得後、以下のエンドポイントでメールアドレスを取得する

**エンドポイント:** `GET /2/users/me?user.fields=confirmed_email`

**レスポンス例:**

```json
{
  "data": {
    "id": "1234567890",
    "name": "Example User",
    "username": "example_user",
    "confirmed_email": "user@example.com"
  }
}
```

> **注意:** `confirmed_email` フィールドはメールアドレスが確認済みの場合のみ返却される。未確認のメールアドレスは返却されない。

#### X Developer Policy におけるデータ保持義務

- ユーザーがアプリ連携を解除した場合、取得したメールアドレスを含むユーザーデータを **速やかに削除** する義務がある
- メールアドレスは認証・通知の目的以外に使用してはならない
- データ保持期間・削除ポリシーはプライバシーポリシーに明記すること

### 2.5 Developer Portal アプリ設定チェックリスト

メールアドレス取得を含む OAuth 2.0 PKCE 連携のために、Developer Portal で以下を設定する:

- [ ] **App permissions**: Read and write（投稿機能に必要）
- [ ] **Type of App**: Web App（Confidential client）
- [ ] **App info > Website URL**: `https://trendbird.app`（本番 URL）
- [ ] **App info > Terms of Service URL**: `https://trendbird.app/terms`
- [ ] **App info > Privacy Policy URL**: `https://trendbird.app/privacy`
- [ ] **App info > Request email from users**: ✅ 有効化
- [ ] **OAuth 2.0 settings > Callback URL**: `https://api.trendbird.app/auth/x/callback`（本番）、`http://localhost:8080/auth/x/callback`（開発）
- [ ] **OAuth 2.0 settings > Client ID**: 環境変数 `X_CLIENT_ID` に設定
- [ ] **OAuth 2.0 settings > Client Secret**: 環境変数 `X_CLIENT_SECRET` に設定（Secret Manager 推奨）
- [ ] **Keys and tokens > Bearer Token**: 環境変数 `X_BEARER_TOKEN` に設定

---

## 3. エンドポイントリファレンス

### 3.1 トレンド取得（WOEID）

#### `GET /2/trends/by/woeid/{woeid}`

特定地域のトレンドを取得する。

| 項目 | 値 |
|------|-----|
| 認証 | Bearer Token（App-only） |
| レート制限 | 後述のレート制限表を参照 |

**パスパラメータ:**

| パラメータ | 型 | 必須 | 説明 |
|------------|------|------|------|
| `woeid` | integer | Yes | Where On Earth ID。日本: `23424856`, 全世界: `1` |

**クエリパラメータ:**

| パラメータ | 型 | デフォルト | 説明 |
|------------|------|-----------|------|
| `max_trends` | integer | 20 | 最大件数（1-50） |
| `trend.fields` | string | — | `trend_name`, `tweet_count`（カンマ区切り） |

**レスポンス例:**

```json
{
  "data": [
    {
      "trend_name": "#AI",
      "tweet_count": 125000
    },
    {
      "trend_name": "ChatGPT",
      "tweet_count": 89000
    }
  ]
}
```

> **注意:** `tweet_count` フィールドは `trend.fields` で明示的に指定しても返却されない場合がある（コミュニティで報告されている既知の問題）。

**主要 WOEID 一覧:**

| 地域 | WOEID |
|------|-------|
| 全世界 | `1` |
| 日本 | `23424856` |
| 東京 | `1118370` |
| 大阪 | `15015370` |
| アメリカ | `23424977` |

### 3.2 パーソナライズドトレンド

#### `GET /2/users/personalized_trends`

認証ユーザーの位置情報・興味関心に基づくパーソナライズされたトレンドを取得する。

| 項目 | 値 |
|------|-----|
| 認証 | OAuth 2.0 User Context（`tweet.read`, `users.read`） |
| レート制限 | User: 1 / 15 分（Basic プラン時） |

**クエリパラメータ:**

| パラメータ | 型 | 必須 | 説明 |
|------------|------|------|------|
| `personalized_trend.fields` | string | No | `category`, `post_count`, `trend_name`, `trending_since`（カンマ区切り）|

**レスポンス例:**

```json
{
  "data": [
    {
      "trend_name": "#AI",
      "category": "Technology",
      "post_count": "125000",
      "trending_since": "2026-02-17T08:00:00.000Z"
    },
    {
      "trend_name": "ChatGPT",
      "category": "Technology",
      "post_count": "89000",
      "trending_since": "2026-02-17T06:30:00.000Z"
    }
  ]
}
```

> **注意:** X Premium 未加入ユーザーの場合、`category` が `"Unknown"`、`trending_since` が空になる等、限定的なレスポンスになる。

**WOEID トレンドとの使い分け:**

| 用途 | 推奨エンドポイント |
|------|-------------------|
| 地域別のトレンド一覧（定期バッチ） | `GET /2/trends/by/woeid/{woeid}` |
| ユーザーごとのパーソナライズドトレンド | `GET /2/users/personalized_trends` |

### 3.3 ツイート検索（Recent）

#### `GET /2/tweets/search/recent`

過去 **7 日間** のツイートを検索する。

| 項目 | 値 |
|------|-----|
| 認証 | Bearer Token / OAuth 2.0 / OAuth 1.0a |
| レート制限 | App: 450 / 15 分, User: 300 / 15 分 |

**クエリパラメータ:**

| パラメータ | 型 | 必須 | 説明 |
|------------|------|------|------|
| `query` | string | Yes | 検索クエリ（1-512 文字、Enterprise: 4,096 文字） |
| `start_time` | ISO 8601 | No | 検索開始日時（inclusive） |
| `end_time` | ISO 8601 | No | 検索終了日時（exclusive） |
| `since_id` | string | No | この ID より新しいツイートのみ |
| `until_id` | string | No | この ID より古いツイートのみ |
| `max_results` | integer | No | 10-100（デフォルト: 10） |
| `next_token` | string | No | ページネーショントークン |
| `sort_order` | enum | No | `recency`（新しい順）/ `relevancy`（関連度順） |
| `tweet.fields` | string | No | `created_at,author_id,public_metrics,lang,conversation_id` |
| `expansions` | string | No | `author_id,attachments.media_keys,referenced_tweets.id` |
| `user.fields` | string | No | `username,verified,public_metrics,profile_image_url` |
| `media.fields` | string | No | `type,height,width,url,preview_image_url` |

**レスポンス例:**

```json
{
  "data": [
    {
      "id": "1234567890",
      "text": "AIの最新トレンドについて...",
      "created_at": "2026-02-17T10:00:00.000Z",
      "author_id": "9876543210",
      "public_metrics": {
        "retweet_count": 150,
        "reply_count": 30,
        "like_count": 500,
        "quote_count": 20,
        "impression_count": 10000
      },
      "lang": "ja"
    }
  ],
  "includes": {
    "users": [
      {
        "id": "9876543210",
        "name": "Example User",
        "username": "example_user"
      }
    ]
  },
  "meta": {
    "result_count": 10,
    "newest_id": "1234567890",
    "oldest_id": "1234567880",
    "next_token": "b26v89c19zqg8o3fpzbhqwlatyem"
  }
}
```

### 3.4 ツイート検索（Full-Archive）

#### `GET /2/tweets/search/all`

**2006 年 3 月以降** の全ツイートを検索する。

| 項目 | 値 |
|------|-----|
| 認証 | Bearer Token（App-only のみ） |
| レート制限 | 300 / 15 分, 1 リクエスト/秒 |
| アクセス | Pro / Enterprise / Pay-Per-Use のみ（Free・Basic は不可） |
| 最大結果数 | 500 / リクエスト |
| クエリ長 | Self-serve: 1,024 文字, Enterprise: 4,096 文字 |

パラメータは `/search/recent` と同一。

### 3.5 ツイートカウント（Recent）

#### `GET /2/tweets/counts/recent`

過去 **7 日間** のツイート数を時系列で取得する。

| 項目 | 値 |
|------|-----|
| 認証 | Bearer Token（App-only） |
| アクセス | Basic 以上 |

**クエリパラメータ:**

| パラメータ | 型 | 必須 | 説明 |
|------------|------|------|------|
| `query` | string | Yes | 検索クエリ |
| `start_time` / `end_time` | ISO 8601 | No | 期間指定 |
| `granularity` | enum | No | `minute` / `hour`（デフォルト）/ `day` |
| `next_token` | string | No | ページネーション |

**レスポンス例:**

```json
{
  "data": [
    {
      "start": "2026-02-17T00:00:00.000Z",
      "end": "2026-02-17T01:00:00.000Z",
      "tweet_count": 42
    },
    {
      "start": "2026-02-17T01:00:00.000Z",
      "end": "2026-02-17T02:00:00.000Z",
      "tweet_count": 67
    }
  ],
  "meta": {
    "total_tweet_count": 1234,
    "next_token": "..."
  }
}
```

### 3.6 ツイートカウント（Full-Archive）

#### `GET /2/tweets/counts/all`

2006 年 3 月以降の全期間のツイート数を取得する。

| 項目 | 値 |
|------|-----|
| アクセス | Pro / Enterprise / Pay-Per-Use のみ |
| デフォルト粒度 | `hour`（長期間は `day` を推奨） |

### 3.7 ツイート投稿

#### `POST /2/tweets`

ユーザーの代理でツイートを投稿する。

| 項目 | 値 |
|------|-----|
| 認証 | OAuth 2.0 User Context（`tweet.read`, `tweet.write`, `users.read`）|
| レート制限 | User: 100 / 15 分, App: 10,000 / 24 時間 |

**リクエストボディ:**

```json
{
  "text": "AIが分析した最新トレンド: #AI が急上昇中！"
}
```

**全フィールド:**

| フィールド | 型 | 説明 |
|------------|------|------|
| `text` | string | 必須。ツイート本文 |
| `media.media_ids` | string[] | メディア ID（最大 4 件）。`poll` と排他 |
| `poll.options` | string[] | 投票選択肢（2-4 件） |
| `poll.duration_minutes` | integer | 投票期間（5-10,080 分） |
| `reply.in_reply_to_tweet_id` | string | リプライ先ツイート ID |
| `quote_tweet_id` | string | 引用元ツイート ID |
| `reply_settings` | string | `"following"` / `"mentionedUsers"` |

**レスポンス（201 Created）:**

```json
{
  "data": {
    "id": "1234567890123456789",
    "text": "AIが分析した最新トレンド: #AI が急上昇中！"
  }
}
```

### 3.8 ツイート削除

#### `DELETE /2/tweets/{id}`

ツイートを削除する。認証ユーザーが所有するツイートのみ削除可能。

| 項目 | 値 |
|------|-----|
| 認証 | OAuth 2.0 User Context（`tweet.read`, `tweet.write`, `users.read`）|

**レスポンス（200 OK）:**

```json
{
  "data": {
    "deleted": true
  }
}
```

### 3.9 ユーザー検索

#### `GET /2/users/by/username/{username}`

ユーザー名からユーザー情報を取得する。

| 項目 | 値 |
|------|-----|
| 認証 | Bearer Token / OAuth 2.0 / OAuth 1.0a |

**デフォルトレスポンスフィールド:** `id`, `name`, `username`

**オプション `user.fields`:** `created_at`, `description`, `profile_image_url`, `public_metrics`, `verified`, `url`, `location`, `protected`

**関連エンドポイント:**

| エンドポイント | 説明 |
|----------------|------|
| `GET /2/users/{id}` | ID でユーザー取得 |
| `GET /2/users/me` | 認証ユーザー自身（User Context のみ） |
| `GET /2/users?ids=...` | ID バッチ取得（最大 100 件） |
| `GET /2/users/by?usernames=...` | ユーザー名バッチ取得（最大 100 件） |

### 3.10 ユーザータイムライン

#### `GET /2/users/{id}/tweets`

特定ユーザーのツイート一覧を取得する。ページネーションで最大 3,200 件取得可能。

| 項目 | 値 |
|------|-----|
| 認証 | Bearer Token / OAuth 2.0 / OAuth 1.0a |
| レート制限 | App: 1,500 / 15 分, User: 900 / 15 分 |

**クエリパラメータ:**

| パラメータ | 型 | 説明 |
|------------|------|------|
| `max_results` | integer | 5-100（デフォルト: 10） |
| `pagination_token` | string | ページネーション |
| `start_time` / `end_time` | ISO 8601 | 期間指定 |
| `exclude` | string | `retweets`, `replies`（カンマ区切り） |
| `tweet.fields` | string | 取得フィールド指定 |

---

## 4. 検索クエリ演算子

### 4.1 スタンドアロン演算子（単独使用可）

| 演算子 | 説明 | 例 |
|--------|------|-----|
| `keyword` | トークン化キーワード | `AI トレンド` |
| `"exact phrase"` | 完全一致フレーズ | `"人工知能"` |
| `#hashtag` | ハッシュタグ | `#AI` |
| `@mention` | メンション | `@username` |
| `$cashtag` | キャッシュタグ | `$AAPL` |
| `from:` | 特定ユーザーのツイート | `from:TwitterDev` |
| `to:` | 特定ユーザーへのリプライ | `to:TwitterDev` |
| `url:` | URL を含むツイート | `url:example.com` |
| `conversation_id:` | スレッド内の全ツイート | `conversation_id:123` |
| `context:` | コンテキストアノテーション | `context:123.456` |
| `entity:` | エンティティ文字列マッチ | `entity:"Bitcoin"` |
| `place:` | 場所タグ付きツイート | `place:"Tokyo"` |
| `place_country:` | 国コード | `place_country:JP` |
| `lang:` | 言語フィルター（BCP 47） | `lang:ja` |

### 4.2 結合必須演算子（スタンドアロン演算子と組み合わせて使用）

| 演算子 | 説明 | 例 |
|--------|------|-----|
| `is:retweet` | リツイートのみ | `from:user is:retweet` |
| `is:reply` | リプライのみ | `from:user is:reply` |
| `is:quote` | 引用ツイートのみ | `from:user is:quote` |
| `is:verified` | 認証済みアカウント | `#AI is:verified` |
| `has:hashtags` | ハッシュタグ付き | `from:user has:hashtags` |
| `has:links` | URL 付き | `from:user has:links` |
| `has:mentions` | メンション付き | `from:user has:mentions` |
| `has:media` | メディア付き | `#photo has:media` |
| `has:images` | 画像付き | `from:user has:images` |
| `has:video_link` | 動画付き | `from:user has:video_link` |

### 4.3 論理演算子

| 構文 | 説明 | 例 |
|------|------|-----|
| （スペース） | AND（暗黙） | `cat dog`（両方含む） |
| `OR` | 論理 OR | `cat OR dog` |
| `-` | 否定（除外） | `-is:retweet` |
| `()` | グルーピング | `(cat OR dog) from:user` |

### 4.4 TrendBird での実用クエリ例

```
// 日本語のオリジナルツイート（RT・リプライ除外）でリンク付き
(#AI OR #MachineLearning) lang:ja -is:retweet -is:reply has:links

// 特定トレンドのツイート数を集計（RT 除外）
"トレンドワード" lang:ja -is:retweet

// 特定ユーザーの最近の投稿を検索
from:username -is:retweet -is:reply
```

---

## 5. レート制限

### 5.1 エンドポイント別レート制限（15 分ウィンドウ）

| エンドポイント | App（Bearer） | User（OAuth） |
|----------------|---------------|---------------|
| `GET /2/trends/by/woeid/{woeid}` | 75 / 15 分 | — |
| `GET /2/users/personalized_trends` | — | 1 / 15 分 |
| `GET /2/tweets/search/recent` | 450 / 15 分 | 300 / 15 分 |
| `GET /2/tweets/search/all` | 300 / 15 分, 1/秒 | — |
| `GET /2/tweets/counts/recent` | 300 / 15 分 | — |
| `GET /2/tweets/counts/all` | 300 / 15 分 | — |
| `POST /2/tweets` | 10,000 / 24 時間 | 100 / 15 分 |
| `DELETE /2/tweets/{id}` | — | 50 / 15 分 |
| `GET /2/users/{id}/tweets` | 1,500 / 15 分 | 900 / 15 分 |
| `GET /2/users/by/username/{username}` | 300 / 15 分 | 300 / 15 分 |
| `GET /2/users/me` | — | 75 / 15 分 |

### 5.2 レスポンスヘッダー

| ヘッダー | 説明 |
|----------|------|
| `x-rate-limit-limit` | ウィンドウ内の最大リクエスト数 |
| `x-rate-limit-remaining` | 残りリクエスト数 |
| `x-rate-limit-reset` | ウィンドウリセットの Unix タイムスタンプ |

### 5.3 429 Too Many Requests 対応

レート制限超過時は HTTP `429` が返される。`x-rate-limit-reset` ヘッダーの値を参照し、リセットまで待機する。

**バックオフ戦略:**

1. `x-rate-limit-remaining` を監視し、残り少なくなったらリクエスト間隔を広げる
2. `429` を受けたら `x-rate-limit-reset` までスリープ
3. リトライ時は指数バックオフを適用（1s → 2s → 4s → ...）
4. 最大リトライ回数を設定し、超過時はエラーを返す

---

## 6. 料金体系

### 6.1 料金体系の概要

- **2026 年 2 月 6 日に Pay-Per-Use（従量課金）が正式ローンチ**
- 固定月額プラン（Free / Basic / Pro / Enterprise）は段階的廃止中（新規登録不可）
- 既存契約者は一時的に維持可能だが、今後 Pay-Per-Use への移行が推奨される
- **TrendBird は Pay-Per-Use を前提として設計する**

### 6.2 Pay-Per-Use の仕組み

| 項目 | 内容 |
|------|------|
| 課金モデル | クレジット事前購入制（[console.x.com](https://console.x.com)） |
| 契約 | 不要。クレジット購入後すぐに利用可能 |
| 消費タイミング | リアルタイム消費（成功レスポンスのみ課金） |
| 自動チャージ | 閾値 + チャージ額を設定可能 |
| 月間支出上限 | 設定可能。上限到達で API リクエストがブロックされる |
| 重複排除 | 24 時間 UTC 日単位。同一リソース（ポスト/ユーザー）の再取得は非課金 |
| 月間ポスト読み取り上限 | **200 万件/月** |

### 6.3 エンドポイント別クレジットコスト

| エンドポイント | コスト（概算） | 備考 |
|---|---|---|
| Post read（lookup / search） | ~$0.005/件 | 取得ポストごとに課金 |
| Post create | ~$0.010/件 | 投稿ごとに課金 |
| User read | ~$0.010/件 | ユーザーごとに課金 |
| DM read | ~$0.010/件 | リクエストごとに課金 |
| Trends / Bookmarks 等 | Developer Console 確認 | 公式未公開 |

> **注意:** 上記は概算値。正確な最新価格は [Developer Console](https://console.x.com) で確認すること。エンドポイントごとのコストは予告なく変更される可能性がある。

### 6.4 xAI クレジット報酬プログラム

X API の月間累計支出に応じて、xAI API（Grok）のクレジットが還元される。

| 月間累計支出 | 還元率 |
|---|---|
| $0〜$199 | 0% |
| $200〜$499 | 10% |
| $500〜$999 | 15% |
| $1,000+ | 20% |

### 6.5 TrendBird 向けコスト試算

| ユースケース | 推奨 | 概算月額コスト |
|---|---|---|
| MVP / 開発・テスト | Pay-Per-Use（支出上限 $50 設定） | $10〜50 |
| 本番（小規模 〜50 ユーザー） | Pay-Per-Use（支出上限 $200 設定） | $50〜150 |
| 本番（中規模 〜200 ユーザー） | Pay-Per-Use（支出上限 $500 設定） | $150〜400 |
| 大規模展開（1,000+ ユーザー） | Pay-Per-Use + xAI 還元活用 | $500〜2,000 |

**TrendBird の主要 API 消費パターン:**

- トレンド取得（5 分間隔バッチ）: 月 ~8,640 リクエスト
- ツイート検索（トピック分析）: ~$0.005/件 × 検索結果数
- ツイート投稿（AI 生成文）: ~$0.010/件 × ユーザー投稿数

---

## 7. Go 実装パターン

### 7.1 HTTPクライアント構成

```go
package external

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://api.x.com"

// TwitterClient は X API v2 のクライアント実装。
// domain/gateway/twitter.go の TwitterClient interface を実装する。
type TwitterClient struct {
	httpClient   *http.Client
	bearerToken  string
}

// NewTwitterClient は新しい TwitterClient を生成する。
func NewTwitterClient(bearerToken string) *TwitterClient {
	return &TwitterClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		bearerToken: bearerToken,
	}
}
```

### 7.2 認証ヘッダー設定

```go
// Bearer Token（App-only）認証のリクエスト送信
func (c *TwitterClient) doAppRequest(ctx context.Context, method, path string, params url.Values) (*http.Response, error) {
	u := baseURL + path
	if params != nil {
		u += "?" + params.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, u, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.bearerToken)

	return c.httpClient.Do(req)
}

// OAuth 2.0 User Context 認証のリクエスト送信
func (c *TwitterClient) doUserRequest(ctx context.Context, method, path string, accessToken string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.httpClient.Do(req)
}
```

### 7.3 レート制限ハンドリング

```go
import (
	"strconv"
	"time"
)

// RateLimitInfo はレート制限の状態を保持する。
type RateLimitInfo struct {
	Limit     int
	Remaining int
	ResetAt   time.Time
}

// parseRateLimit はレスポンスヘッダーからレート制限情報を取得する。
func parseRateLimit(resp *http.Response) RateLimitInfo {
	info := RateLimitInfo{}

	if v := resp.Header.Get("x-rate-limit-limit"); v != "" {
		info.Limit, _ = strconv.Atoi(v)
	}
	if v := resp.Header.Get("x-rate-limit-remaining"); v != "" {
		info.Remaining, _ = strconv.Atoi(v)
	}
	if v := resp.Header.Get("x-rate-limit-reset"); v != "" {
		ts, _ := strconv.ParseInt(v, 10, 64)
		info.ResetAt = time.Unix(ts, 0)
	}

	return info
}

// waitForRateLimit は 429 レスポンス時にリセットまで待機する。
func waitForRateLimit(ctx context.Context, info RateLimitInfo) error {
	waitDuration := time.Until(info.ResetAt)
	if waitDuration <= 0 {
		return nil
	}

	select {
	case <-time.After(waitDuration):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
```

### 7.4 エラーハンドリング

```go
// APIError は X API のエラーレスポンスを表す。
type APIError struct {
	StatusCode int
	Title      string `json:"title"`
	Detail     string `json:"detail"`
	Type       string `json:"type"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("x api error %d: %s - %s", e.StatusCode, e.Title, e.Detail)
}

// handleResponse はレスポンスのステータスコードを検証する。
func handleResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)

	apiErr := &APIError{StatusCode: resp.StatusCode}
	if err := json.Unmarshal(body, apiErr); err != nil {
		apiErr.Detail = string(body)
	}

	return apiErr
}
```

### 7.5 トレンド取得の実装例

```go
// Trend はトレンド情報を表す。
type Trend struct {
	TrendName  string `json:"trend_name"`
	TweetCount int    `json:"tweet_count"`
}

type trendsResponse struct {
	Data []Trend `json:"data"`
}

// GetTrends は指定 WOEID のトレンドを取得する。
func (c *TwitterClient) GetTrends(ctx context.Context, woeid int) ([]Trend, error) {
	params := url.Values{}
	params.Set("max_trends", "50")
	params.Set("trend.fields", "trend_name,tweet_count")

	path := fmt.Sprintf("/2/trends/by/woeid/%d", woeid)
	resp, err := c.doAppRequest(ctx, http.MethodGet, path, params)
	if err != nil {
		return nil, fmt.Errorf("get trends: %w", err)
	}
	defer resp.Body.Close()

	if err := handleResponse(resp); err != nil {
		return nil, err
	}

	var result trendsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode trends: %w", err)
	}

	return result.Data, nil
}
```

### 7.6 パーソナライズドトレンド取得の実装例

```go
// PersonalizedTrend はパーソナライズされたトレンド情報を表す。
type PersonalizedTrend struct {
	TrendName     string `json:"trend_name"`
	Category      string `json:"category"`
	PostCount     string `json:"post_count"`
	TrendingSince string `json:"trending_since"`
}

type personalizedTrendsResponse struct {
	Data []PersonalizedTrend `json:"data"`
}

// GetPersonalizedTrends は認証ユーザーのパーソナライズドトレンドを取得する。
// WOEID 版と異なり、OAuth 2.0 User Context（accessToken）が必要。
func (c *TwitterClient) GetPersonalizedTrends(ctx context.Context, accessToken string) ([]PersonalizedTrend, error) {
	params := url.Values{}
	params.Set("personalized_trend.fields", "category,post_count,trend_name,trending_since")

	path := "/2/users/personalized_trends?" + params.Encode()
	resp, err := c.doUserRequest(ctx, http.MethodGet, path, accessToken, nil)
	if err != nil {
		return nil, fmt.Errorf("get personalized trends: %w", err)
	}
	defer resp.Body.Close()

	if err := handleResponse(resp); err != nil {
		return nil, err
	}

	var result personalizedTrendsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode personalized trends: %w", err)
	}

	return result.Data, nil
}
```

### 7.7 ツイート投稿の実装例

```go
import (
	"bytes"
)

type createTweetRequest struct {
	Text string `json:"text"`
}

type createTweetResponse struct {
	Data struct {
		ID   string `json:"id"`
		Text string `json:"text"`
	} `json:"data"`
}

// PostTweet はユーザーの代理でツイートを投稿する。
func (c *TwitterClient) PostTweet(ctx context.Context, accessToken, text string) (string, error) {
	body, err := json.Marshal(createTweetRequest{Text: text})
	if err != nil {
		return "", fmt.Errorf("marshal tweet: %w", err)
	}

	resp, err := c.doUserRequest(ctx, http.MethodPost, "/2/tweets", accessToken, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("post tweet: %w", err)
	}
	defer resp.Body.Close()

	if err := handleResponse(resp); err != nil {
		return "", err
	}

	var result createTweetResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode tweet response: %w", err)
	}

	return result.Data.ID, nil
}
```

### 7.8 ツイート検索の実装例

```go
type searchResponse struct {
	Data []struct {
		ID            string `json:"id"`
		Text          string `json:"text"`
		AuthorID      string `json:"author_id"`
		CreatedAt     string `json:"created_at"`
		Lang          string `json:"lang"`
		PublicMetrics struct {
			RetweetCount    int `json:"retweet_count"`
			ReplyCount      int `json:"reply_count"`
			LikeCount       int `json:"like_count"`
			QuoteCount      int `json:"quote_count"`
			ImpressionCount int `json:"impression_count"`
		} `json:"public_metrics"`
	} `json:"data"`
	Meta struct {
		ResultCount int    `json:"result_count"`
		NewestID    string `json:"newest_id"`
		OldestID    string `json:"oldest_id"`
		NextToken   string `json:"next_token"`
	} `json:"meta"`
}

// SearchRecentTweets は過去 7 日間のツイートを検索する。
func (c *TwitterClient) SearchRecentTweets(ctx context.Context, query string, maxResults int) (*searchResponse, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("max_results", strconv.Itoa(maxResults))
	params.Set("tweet.fields", "created_at,author_id,public_metrics,lang")
	params.Set("sort_order", "relevancy")

	resp, err := c.doAppRequest(ctx, http.MethodGet, "/2/tweets/search/recent", params)
	if err != nil {
		return nil, fmt.Errorf("search tweets: %w", err)
	}
	defer resp.Body.Close()

	if err := handleResponse(resp); err != nil {
		return nil, err
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode search response: %w", err)
	}

	return &result, nil
}
```

### 7.9 OAuth 2.0 PKCE トークン管理

```go
import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// generateCodeVerifier は PKCE 用の code_verifier を生成する。
func generateCodeVerifier() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// generateCodeChallenge は code_verifier から S256 チャレンジを生成する。
func generateCodeChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// buildAuthorizationURL は X の認可画面 URL を構築する。
func buildAuthorizationURL(clientID, redirectURI, codeChallenge, state string) string {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", clientID)
	params.Set("redirect_uri", redirectURI)
	params.Set("scope", "tweet.read tweet.write users.read users.email offline.access")
	params.Set("state", state)
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "S256")

	return "https://x.com/i/oauth2/authorize?" + params.Encode()
}
```
