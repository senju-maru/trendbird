# TrendBird

X (Twitter) のトレンドデータを収集・分析し、AI による投稿文の自動生成やトピック監視を行う Web アプリケーションです。

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | Next.js 16, React 19, TypeScript, Tailwind CSS v4 |
| Backend | Go 1.25, Connect-RPC, GORM |
| Database | PostgreSQL 16 |
| API Schema | Protocol Buffers v3 (Buf) |
| External APIs | X API v2, Claude API (Anthropic) |

## Prerequisites

以下のツールをインストールしてください。

### 1. Go (1.25+)

```bash
# macOS (Homebrew)
brew install go

# バージョン確認
go version
```

> 他の OS: https://go.dev/doc/install

### 2. Node.js (20+) & npm

```bash
# macOS (Homebrew)
brew install node

# または nvm を使う場合
nvm install 20 && nvm use 20

# バージョン確認
node -v && npm -v
```

### 3. PostgreSQL 16

Docker を使う方法とローカルインストールの 2 通りがあります。

**方法 A: Docker (推奨)**

```bash
# Docker Desktop をインストール
# https://www.docker.com/products/docker-desktop/
docker --version
```

**方法 B: ローカルインストール**

```bash
# macOS
brew install postgresql@16
brew services start postgresql@16

# データベースを作成
createdb trendbird
```

## Quick Start

### 1. リポジトリのクローン

```bash
git clone https://github.com/trendbird/trendbird.git
cd trendbird
```

### 2. セットアップ (初回のみ)

```bash
make setup
```

このコマンドは以下を自動で実行します:

1. フロントエンドの依存関係インストール (`npm install`)
2. 環境変数テンプレートのコピー (`.env.example` → `.env`)
3. PostgreSQL の起動 (Docker Compose)
4. データベースマイグレーション
5. サンプルデータの投入

### 3. 環境変数の設定

外部 API を利用する機能には API キーが必要です。

**backend/.env** を編集:

```bash
# データベース (setup で自動設定済み)
DATABASE_URL=postgres://postgres:postgres@localhost:5432/trendbird?sslmode=disable

# JWT シークレット (任意のランダム文字列)
JWT_SECRET=your-secret-key-here

# X (Twitter) OAuth 2.0 — ログイン・投稿機能に必要
# https://developer.x.com/en/portal/dashboard で取得
X_CLIENT_ID=your-x-client-id
X_CLIENT_SECRET=your-x-client-secret
X_REDIRECT_URI=http://localhost:3000/callback

# Claude API — AI 投稿文生成に必要
# https://console.anthropic.com/ で取得
ANTHROPIC_API_KEY=sk-ant-xxx
```

> **Note:** すべての API キーがなくてもアプリは起動します。各機能を使うときに対応する API キーを設定してください。

### 4. 起動

```bash
make start
```

以下のサーバーが起動します:

| サービス | URL |
|---------|-----|
| Frontend | http://localhost:3000 |
| Backend API | http://localhost:8080 |
| Health Check | http://localhost:8080/health |

停止するには `Ctrl+C` を押してください。

### 5. データ収集スケジューラ (オプション)

トレンドデータの定期収集、通知送信、予約投稿の自動公開を行うスケジューラを起動できます。

```bash
make scheduler
```

スケジューラは以下のジョブを cron スケジュールで自動実行します:

| ジョブ | デフォルト間隔 | 必要な API キー | 説明 |
|-------|-------------|---------------|------|
| `trend-fetch` | 12 時間ごと | `X_BEARER_TOKEN` | X のツイート数を収集し、z-score でスパイク検出 |
| `topic-research` | trend-fetch 後 | `ANTHROPIC_API_KEY` | Claude でトピックの Web 調査 |
| `spike-notification` | 5 分ごと | — | スパイク検出時にアプリ内通知 |
| `rising-notification` | 10 分ごと | — | ライジングトレンド通知 |
| `scheduled-publish` | 毎時 0 分 | X OAuth | 予約投稿の自動公開 |
| `reply-dm-batch` | 毎時 0 分 | X OAuth | 自動 DM 送信 |

不要なジョブは環境変数で無効化できます:

```bash
# .env に追記
SCHEDULER_DISABLED_JOBS=daily-summary,weekly-summary,reply-dm-batch
```

スケジュールのカスタマイズも可能です:

```bash
# .env に追記（cron 式）
SCHEDULE_TREND_FETCH=0 */6 * * *    # 6 時間ごとに変更
SCHEDULER_TIMEZONE=America/New_York  # タイムゾーン変更
```

個別ジョブの手動実行:

```bash
make batch-run JOB=trend-fetch
```

---

## Development with Claude Code

このプロジェクトは [Claude Code](https://docs.anthropic.com/en/docs/claude-code) での開発に最適化されています。プロジェクト固有のルール・スキルが `.claude/` ディレクトリに定義されており、Claude Code が自動的に読み込みます。

### インストール

```bash
npm install -g @anthropic-ai/claude-code
```

### 使い方

```bash
# プロジェクトルートで起動
cd trendbird
claude
```

### できること

```
> このファイルは何をしている？
> 認証フローを説明して
> 新しい API エンドポイントを追加して
> トピック検索機能を実装して
> バックエンドの E2E テストを実行して
> 変更をコミットして
```

### プロジェクト固有スキル (Slash Commands)

Claude Code 内で使えるコマンドです。プロジェクトのルール・設計方針に沿った回答やコード生成を行います。

| コマンド | 説明 |
|---------|------|
| `/backend-architecture` | バックエンドのクリーンアーキテクチャ設計ガイド |
| `/design-system` | フロントエンド UI のデザインシステム |
| `/protobuf-style` | Protobuf 定義のスタイルガイド |
| `/backend-e2e-test` | バックエンド E2E テストパターン |
| `/frontend-e2e-test` | フロントエンド E2E テストガイド |
| `/implementation-plan` | 実装計画書の作成ガイド |
| `/commit-push` | 変更のコミット & プッシュ |

---

## Project Structure

```
trendbird/
├── backend/                 # Go バックエンド
│   ├── cmd/
│   │   ├── server/          #   API サーバーのエントリポイント
│   │   ├── scheduler/       #   ローカルスケジューラのエントリポイント
│   │   ├── batch/           #   バッチジョブ手動実行用
│   │   └── seed/            #   サンプルデータ投入
│   ├── gen/                 #   Protobuf 生成コード (自動生成)
│   ├── internal/
│   │   ├── adapter/         #   ハンドラ・ルーター・ミドルウェア
│   │   ├── domain/          #   エンティティ・リポジトリIF
│   │   ├── infrastructure/  #   DB・外部API・認証
│   │   ├── usecase/         #   ビジネスロジック
│   │   ├── di/              #   依存性注入
│   │   └── e2etest/         #   E2E テスト
│   ├── migrations/          #   SQL マイグレーション
│   └── .env.example         #   環境変数テンプレート
├── frontend/                # Next.js フロントエンド
│   ├── src/
│   │   ├── app/             #   App Router ページ
│   │   ├── components/      #   React コンポーネント
│   │   ├── hooks/           #   カスタムフック
│   │   ├── lib/             #   ユーティリティ
│   │   ├── stores/          #   Zustand ストア
│   │   └── gen/             #   Protobuf 生成コード (自動生成)
│   ├── e2e/                 #   Playwright E2E テスト
│   └── .env.local.example   #   環境変数テンプレート
├── proto/                   # Protobuf 定義ファイル
│   └── trendbird/v1/        #   API スキーマ
├── plans.json               # プラン設定 (Single Source of Truth)
├── docker-compose.yml       # PostgreSQL コンテナ定義
├── Makefile                 # 開発コマンド
├── CLAUDE.md                # Claude Code プロジェクト設定
└── .claude/                 # Claude Code スキル・ルール定義
```

## Make Commands

```bash
make help            # 全コマンドの一覧を表示

# セットアップ & 起動
make setup           # 初回セットアップ (依存関係 + DB + マイグレーション + シード)
make start           # フロントエンド + バックエンドを起動
make scheduler       # データ収集スケジューラを起動

# 個別起動
make frontend-dev    # フロントエンドのみ起動 (Next.js dev server)
make backend-dev     # バックエンドのみ起動 (air でホットリロード)

# データベース
make db-up           # PostgreSQL コンテナを起動
make db-down         # PostgreSQL コンテナを停止
make migrate         # マイグレーションを実行
make seed            # サンプルデータを投入

# バッチジョブ
make batch-run JOB=trend-fetch  # 個別ジョブを手動実行

# ユーティリティ
make kill-dev        # 開発ポート (3000, 8080) のプロセスを停止
```

## Testing

```bash
# バックエンド E2E テスト
cd backend && go test ./internal/e2etest/...

# バックエンド全テスト
cd backend && go test ./...

# フロントエンド E2E テスト
cd frontend && npm run e2e
```

## Architecture

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

## License

MIT
