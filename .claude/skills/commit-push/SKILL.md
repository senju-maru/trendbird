---
name: commit-push
description: 変更ファイルをサービスごとにカテゴリ分けし、カテゴリ別にコミットして main に push する。「コミットして」「push して」「変更を反映して」と言われた際に使用する。
---

# /commit-push コマンド

変更ファイルをサービスごとにカテゴリ分けし、カテゴリ別にコミットして main に push する。

## 手順

1. `git status` で変更ファイル（staged / unstaged / untracked）を検出する
2. 変更がなければ「コミットする変更がありません」と伝えて終了する
3. 変更ファイルをディレクトリプレフィックスで以下の5カテゴリに分類する:
   - **backend/** → プレフィックス `backend:`
   - **frontend/** → プレフィックス `frontend:`
   - **proto/** → プレフィックス `proto:`
   - **infra/** → プレフィックス `infra:`
   - **その他**（上記に該当しないファイル: docs/, .github/, ルート設定ファイル等）→ プレフィックス `chore:`
4. カテゴリごとに以下を実行する（該当ファイルがないカテゴリはスキップ）:
   a. 該当ファイルを `git add` でステージングする
   b. 変更内容を分析してコミットメッセージを自動生成する
   c. `git commit` を実行する
   d. `git push origin main` を実行する（**カテゴリごとに即 push**）
5. コミットメッセージの形式:
   ```
   <prefix>: <変更内容の要約>

   Co-Authored-By: Claude Opus 4.6 <noreply@anthropic.com>
   ```
   - `$ARGUMENTS` が指定されている場合は、要約の補足情報として活用する
6. 完了したコミットの一覧を表示する

## カテゴリの処理順序

proto → backend → frontend → infra → その他

## 重要な注意事項

- 各カテゴリのコミットは独立して行う（1カテゴリ = 1コミット）
- コミットメッセージは日本語ではなく英語で書く
- `.env` やクレデンシャル系ファイルが含まれていたら警告してスキップする
- push 前にユーザーに確認を取る
- **カテゴリごとに commit → push する（まとめて push しない）。** GitHub Actions の paths フィルタは push 単位で評価されるため、まとめて push すると無関係なワークフローが発火する
