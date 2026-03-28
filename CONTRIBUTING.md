# Contributing to TrendBird

TrendBird への貢献に興味を持っていただきありがとうございます。このドキュメントでは、貢献の方法とガイドラインを説明します。

## 開発環境のセットアップ

[README.md](README.md) の Quick Start セクションに従って開発環境を構築してください。

## 貢献の流れ

1. **Issue を確認する** — 既存の Issue を確認し、取り組みたいものがあればコメントしてください
2. **Fork & Clone** — リポジトリを Fork し、ローカルにクローンします
3. **ブランチを作成** — `feature/issue-番号-簡潔な説明` の形式でブランチを作成します
4. **実装** — 変更を加えます
5. **テストを実行** — 全テストが通ることを確認します
6. **Pull Request を作成** — 変更内容を説明する PR を作成します

## テスト要件

PR をマージするには、以下のテストが全て PASS する必要があります。

```bash
# バックエンド E2E テスト
cd backend && go test ./internal/e2etest/...

# フロントエンド E2E テスト
cd frontend && npm run e2e

# ビルド確認
cd backend && go build ./...
cd frontend && npm run build
```

新しい機能を追加する場合は、対応する E2E テストも追加してください。

## コミットメッセージ

以下の形式に従ってください。

```
<type>: <簡潔な説明>

<詳細な説明（必要な場合）>
```

**type の種類:**

| type | 説明 |
|------|------|
| `feat` | 新機能 |
| `fix` | バグ修正 |
| `docs` | ドキュメントのみの変更 |
| `refactor` | リファクタリング（機能変更なし） |
| `test` | テストの追加・修正 |
| `chore` | ビルド・ツール・設定の変更 |
| `infra` | インフラ関連の変更 |

**例:**

```
feat: トピック検索にフィルター機能を追加

キーワード・カテゴリ・期間でトピックを絞り込めるようにした。
```

## コードスタイル

### バックエンド (Go)

- クリーンアーキテクチャに従う（handler → usecase → repository）
- 新しい usecase/repository 関数を作成したら、[バックエンドチェックリスト](.claude/rules/backend-checklist.md)を確認
- Connect-RPC でハンドラを実装

### フロントエンド (TypeScript / React)

- Next.js App Router を使用
- コンポーネントは `src/components/` に配置
- 状態管理は Zustand を使用

### Protobuf

- `proto/trendbird/v1/` にスキーマを定義
- 日本語コメントを必ず付与

## Pull Request のガイドライン

- **1つの PR には1つの変更** — 複数の無関係な変更を混ぜない
- **説明を書く** — 何を変更したか、なぜ変更したかを明記
- **スクリーンショット** — UI の変更がある場合はスクリーンショットを添付
- **テスト結果** — テストが全て通っていることを確認

## Issue の報告

バグ報告や機能リクエストは [Issue テンプレート](.github/ISSUE_TEMPLATE/) を使って作成してください。

## 質問・相談

実装方針に迷った場合は、Issue でディスカッションを開いてください。

## ライセンス

貢献いただいたコードは [MIT License](LICENSE) の下で公開されます。
