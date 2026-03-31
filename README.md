# TrendBird

**Xの「今バズっていること」を自動で見つけて、AIが投稿文まで作ってくれる。完全無料。**

---

## 月額0円で、ここまでできる

**トレンドを見つける → 投稿文を作る → 投稿する。この全部を自動でやります。**

- **バズの兆しを自動検知** — 気になるキーワードを登録するだけ。急上昇トレンドを5分ごとに検知して通知
- **AIが投稿文を自動生成** — トレンドの背景をWeb検索で調べ、カジュアル・速報・分析の3パターンで下書きを作成
- **予約投稿** — 作った投稿をベストなタイミングで自動公開
- **自動リプライ・自動DM** — 特定キーワードに反応して自動で返信・DM送信。キャンペーン運用にも対応
- **投稿の反応を記録** — いいね・リポスト・リプライ数を自動追跡
- **Claude Code から自然言語で操作** — 「明日の12時に投稿予約して」と話しかけるだけ

他のサービスは月額980円〜22,000円。TrendBirdはオープンソースなので **ずっと無料** です。

## 他サービスとの比較

|  | SocialDog | XTEP | TrendBird |
|---|:---:|:---:|:---:|
| **月額** | 980円〜 | 22,000円 | **0円** |
| トレンド自動検知 | ❌ | ❌ | ✅ |
| AI投稿文生成 | ❌ | ❌ | ✅ |
| 予約投稿 | ✅ | ✅ | ✅ |
| 自動リプライ | ❌ | ✅ | ✅ |
| 自動DM | ❌ | ✅ | ✅ |
| キーワード監視 | 一部 | ❌ | ✅ |
| スパイク通知 | ❌ | ❌ | ✅ |
| 投稿履歴・分析 | ✅ | ❌ | ✅ |
| 検知→生成→投稿の自動化 | ❌ | ❌ | ✅ |
| セルフホスト | ❌ | ❌ | ✅ |
| MCP対応（AI操作） | ❌ | ❌ | ✅ |
| API公開 | ❌ | ❌ | **全RPC** |
| CLI / 自然言語操作 | ❌ | ❌ | ✅ |
| ソースコード | 非公開 | 非公開 | **OSS** |

---

## 主な機能

### トレンド検知・監視
- 気になるキーワードを登録すると、5〜10分ごとに自動でチェック
- 「急にバズり始めた」を統計的に検知してリアルタイム通知
- 見逃さず、最速でトレンドに乗れる

### AI投稿文生成
- トレンドの背景をWebから自動調査
- カジュアル / 速報 / 分析の3スタイルで投稿文を生成
- ワンクリックで下書き保存・即時投稿・予約投稿

### 自動リプライ・自動DM
- 特定のポストへのリプライを監視し、キーワードに反応して自動返信
- フォローやリプライに対して自動DM送信
- 送信履歴を記録し、同じ人に二重送信しない

### 投稿管理
- 下書き・予約投稿・即時投稿を一画面で管理
- 投稿ごとの反応（いいね・リポスト・リプライ・閲覧数）を記録
- 最適な投稿タイミングを提案

---

## はじめかた

### 必要なもの

- Docker（[インストール](https://www.docker.com/products/docker-desktop/)）
- X Developer アカウント（[developer.x.com](https://developer.x.com/) で無料取得）

### 最短で動かす（Docker Compose）

```bash
git clone https://github.com/trendbird/trendbird.git
cd trendbird
cp .env.example .env
```

`.env` を開いて、X API のキーを設定:

```bash
JWT_SECRET=any-random-string-here
X_CLIENT_ID=your-x-client-id
X_CLIENT_SECRET=your-x-client-secret
```

起動:

```bash
docker compose up -d --build
```

ブラウザで http://localhost:3000 を開いてXアカウントでログインしてください。

> **Claude Code で操作するには:** ログイン完了後に Claude Code を起動（または再起動）すると、自然言語での操作が使えるようになります。

### ローカル開発で動かす

<details>
<summary>詳細手順を表示</summary>

#### 追加で必要なもの

- Go 1.25+
- Node.js 20+

#### セットアップ

```bash
make setup
```

自動で以下を実行します:
1. フロントエンド依存パッケージのインストール
2. 環境変数ファイルのコピー
3. PostgreSQL の起動
4. DBマイグレーション
5. サンプルデータの投入

#### 環境変数の設定

`backend/.env` を編集:

**必須:**

| 変数 | 説明 |
|---|---|
| `JWT_SECRET` | JWT署名用のランダム文字列 |
| `X_CLIENT_ID` | X OAuth 2.0 クライアントID |
| `X_CLIENT_SECRET` | X OAuth 2.0 シークレット |
| `X_REDIRECT_URI` | コールバックURL（デフォルト: `http://localhost:3000/callback`） |

**オプション:**

| 変数 | 説明 |
|---|---|
| `X_BEARER_TOKEN` | X API Bearer Token（ツイート数カウント・検索用） |
| `ANTHROPIC_API_KEY` | Claude API キー（AI投稿生成用。未設定時はスキップ） |

> すべてのAPIキーがなくてもアプリは起動します。

#### 起動

```bash
make start
```

| サービス | URL |
|---|---|
| フロントエンド | http://localhost:3000 |
| バックエンド API | http://localhost:8080 |

#### スケジューラ起動（オプション）

トレンド自動取得・通知・予約投稿の自動実行に必要です:

```bash
make scheduler
```

| ジョブ | 周期 | 内容 |
|---|---|---|
| trend-fetch | 12時間ごと | トレンドデータ取得・分析 |
| spike-notification | 5分ごと | スパイク検知 → 通知 |
| rising-notification | 10分ごと | 上昇トレンド検知 → 通知 |
| scheduled-publish | 毎時00分 | 予約投稿の自動公開 |
| reply-dm-batch | 毎時00分 | 自動DM送信 |
| auto-reply-batch | 毎時00分 | 自動リプライ送信 |
| topic-research | trend-fetch後 | AIでトピック背景調査 |

ジョブを個別に手動実行:

```bash
make batch-run JOB=trend-fetch
```

スケジュールのカスタマイズ:

```bash
# backend/.env
SCHEDULE_TREND_FETCH=0 */6 * * *     # 6時間ごとに変更
SCHEDULER_TIMEZONE=America/New_York   # タイムゾーン変更
SCHEDULER_DISABLED_JOBS=reply-dm-batch # ジョブ無効化
```

</details>

### Makeコマンド一覧

| コマンド | 説明 |
|---|---|
| `make setup` | 初期セットアップ |
| `make start` | フロントエンド + バックエンド起動 |
| `make dev` | 開発モード起動（ホットリロード） |
| `make scheduler` | スケジューラ起動 |
| `make batch-run JOB=xxx` | バッチジョブ単発実行 |
| `make migrate` | DBマイグレーション |
| `make seed` | サンプルデータ投入 |
| `make db-up` / `db-down` | PostgreSQL起動 / 停止 |
| `make docker-up` / `docker-down` | Docker全サービス起動 / 停止 |

---

## 技術スタック

| レイヤー | 技術 |
|---|---|
| フロントエンド | Next.js 16 / React 19 / TypeScript / Tailwind CSS |
| バックエンド | Go 1.25 / Connect RPC / GORM |
| データベース | PostgreSQL 16 |
| AI | Claude API (Anthropic) |
| API定義 | Protocol Buffers / Buf |
| テスト | Playwright (E2E) / Go testing |

## テスト

```bash
# バックエンド E2E テスト
cd backend && go test ./internal/e2etest/...

# フロントエンド E2E テスト
cd frontend && npm run e2e
```

## アーキテクチャ

```
┌──────────────┐     Connect-RPC      ┌──────────────┐
│   Next.js    │ ◄──── (HTTP) ─────► │   Go Server   │
│  Frontend    │                      │   Backend     │
│  :3000       │                      │   :8080       │
└──────────────┘                      └───────┬───────┘
                                              │
┌──────────────┐                              │
│  Scheduler   │──(cron)──┐                   │
│  (optional)  │          │                   │
└──────────────┘          │                   │
                          ▼                   │
                  ┌──────────────┐            │
                  │              │            │
             ┌────▼────┐   ┌────▼────┐  ┌────▼────┐
             │ Postgres │   │ X API v2 │  │ Claude  │
             │  :5432   │   │(Twitter) │  │   API   │
             └──────────┘   └─────────┘  └─────────┘
```

## プロジェクト構成

```
trendbird/
├── backend/
│   ├── cmd/                 # エントリポイント (server / scheduler / batch / seed)
│   ├── internal/
│   │   ├── domain/          # エンティティ・リポジトリIF
│   │   ├── usecase/         # ビジネスロジック
│   │   ├── adapter/         # ハンドラ・ルーター
│   │   ├── infrastructure/  # DB実装・外部API
│   │   └── di/              # 依存性注入
│   ├── gen/                 # Protobuf生成コード (Go)
│   └── migrations/          # SQLマイグレーション
├── frontend/
│   ├── src/
│   │   ├── app/             # Next.js App Router
│   │   ├── components/      # UIコンポーネント
│   │   ├── hooks/           # カスタムフック
│   │   └── gen/             # Protobuf生成コード (TypeScript)
│   └── e2e/                 # Playwright E2Eテスト
├── proto/                   # Protobuf API定義
├── docker-compose.yml
├── Makefile
└── CLAUDE.md                # Claude Code設定
```

## Claude Code から操作する

このプロジェクトは [Claude Code](https://docs.anthropic.com/en/docs/claude-code) に対応しており、**自然言語で全機能を操作**できます。

```bash
npm install -g @anthropic-ai/claude-code
cd trendbird
claude
```

起動したら、そのまま話しかけるだけ:

```
> 監視中のトピック一覧を見せて
> 「AI × 働き方改革」というトピックを追加して
> AIでトピックの投稿文を生成して
> 明日の12時に投稿予約して
> 自動リプライルールを作成して
> 通知を確認して
> 下書き一覧を見せて
```

### 使えるツール一覧

| ツール | できること |
|---|---|
| `list_topics` | 監視中のトピック一覧 |
| `create_topic` | トピック作成（監視開始） |
| `get_topic` | トピック詳細 |
| `delete_topic` | トピック削除 |
| `generate_posts` | AIで投稿文を3パターン生成 |
| `create_draft` | 下書き作成 |
| `list_drafts` | 下書き・予約投稿一覧 |
| `schedule_post` | 投稿予約 |
| `create_and_schedule_post` | 下書き作成 + 予約を一発で |
| `publish_post` | 今すぐXに投稿 |
| `list_post_history` | 投稿履歴（反応数つき） |
| `list_notifications` | 通知一覧 |
| `mark_all_notifications_read` | 全通知を既読 |
| `list_auto_reply_rules` | 自動リプライルール一覧 |
| `create_auto_reply_rule` | 自動リプライルール作成 |
| `delete_auto_reply_rule` | 自動リプライルール削除 |
| `list_auto_dm_rules` | 自動DMルール一覧 |
| `create_auto_dm_rule` | 自動DMルール作成 |
| `delete_auto_dm_rule` | 自動DMルール削除 |

## ライセンス

MIT
