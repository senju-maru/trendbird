# TrendBird Frontend

X(Twitter)トレンド検知 + AI投稿文自動生成 OSS「TrendBird」のフロントエンドアプリケーション。

## 技術スタック

| カテゴリ | 技術 |
|---------|------|
| フレームワーク | Next.js 16 (App Router) |
| 言語 | TypeScript 5 |
| スタイリング | Tailwind CSS v4 |
| 状態管理 | Zustand |
| API通信 | Connect RPC (Protobuf) |
| フォーム | React Hook Form + Zod |
| アニメーション | Framer Motion |
| チャート | D3.js |
| 認証 | NextAuth.js |
| アイコン | Lucide React |
| テスト | Vitest + Testing Library |

## セットアップ

```bash
# 依存パッケージのインストール
npm install

# 環境変数の設定
cp .env.local.example .env.local
```

### 環境変数

| 変数名 | 説明 | デフォルト |
|--------|------|-----------|
| `NEXT_PUBLIC_API_URL` | バックエンドAPIのURL | — |
| `NEXTAUTH_SECRET` | NextAuth用シークレット | — |
| `NEXTAUTH_URL` | アプリケーションのURL | `http://localhost:3000` |

## 開発サーバー起動

```bash
npm run dev
```

http://localhost:3000 で開発サーバーが起動します。

## ディレクトリ構成

```
src/
├── app/            # Next.js App Router ページ
│   ├── (auth)/     #   認証系（ログイン・登録）
│   ├── (app)/      #   メインアプリ
│   ├── (dashboard)/  # ダッシュボード
│   └── api/        #   APIルート
├── components/     # Reactコンポーネント
│   ├── ui/         #   UIプリミティブ（Button, Card, Badge等）
│   ├── layout/     #   レイアウトコンポーネント
│   ├── dashboard/  #   ダッシュボード固有
│   ├── topics/     #   トピック管理
│   ├── posts/      #   投稿管理
│   ├── charts/     #   D3チャート
│   └── ...         #   その他機能別コンポーネント
├── gen/            # Protobufから自動生成されたコード（手動編集禁止）
│   └── trendbird/v1/
├── hooks/          # カスタムReact Hooks
├── lib/            # ユーティリティ・設定
│   ├── transport.ts    # Connect RPCトランスポート
│   ├── proto-converters.ts  # Proto型 → フロントエンド型の変換
│   ├── connect.ts      # Connect RPCクライアント生成
│   ├── connect-error.ts # エラーハンドリング
│   ├── constants.ts    # 定数
│   └── utils.ts        # 汎用ユーティリティ
├── stores/         # Zustandストア
├── test/           # テスト設定
└── types/          # TypeScript型定義
```

## API通信アーキテクチャ

データは以下のフローで流れます：

```
Proto定義 (.proto)
  ↓  buf generate
生成コード (gen/trendbird/v1/*_pb.ts)
  ↓
Proto Converters (lib/proto-converters.ts)  ← Proto型をフロントエンド型に変換
  ↓
Custom Hooks (hooks/)                       ← API呼び出しとデータ取得
  ↓
Zustand Stores (stores/)                    ← グローバル状態管理
  ↓
Components (components/)                    ← UI描画
```

## Protobuf コード生成

プロジェクトルート（`../`）の `buf.gen.yaml` に基づき、Proto定義からTypeScriptコードを生成します。

```bash
npm run buf:generate
```

生成されたコードは `src/gen/` に出力されます。このディレクトリのファイルは手動で編集しないでください。

## テスト

```bash
# テスト実行
npx vitest

# ウォッチモード
npx vitest --watch
```

## 主要スクリプト一覧

| コマンド | 説明 |
|---------|------|
| `npm run dev` | 開発サーバー起動 |
| `npm run build` | プロダクションビルド |
| `npm run start` | プロダクションサーバー起動 |
| `npm run lint` | ESLintによる静的解析 |
| `npm run buf:generate` | Protobufコード生成 |
