---
description: バックエンド実装完了チェック — usecase/リポジトリ新規作成・修正時の4項目確認
globs: backend/**
---

usecase 関数またはリポジトリ関数を**新規作成・修正した後**、コードを保存する前に必ず以下の4項目を確認すること：

1. **トランザクション確認**
   - usecase 内で 2 テーブル以上に書き込む操作がある場合 → `txManager.RunInTx` で包む
   - 「A を作成して B にリンクする」パターン（例: Topic 作成 → UserTopic 紐付け）は必ず同一 tx に入れる
   - 既存の類似 usecase がトランザクションを使っているかを確認し、一貫性を保つ

2. **並行安全性確認**
   - `Find... → if nil { Create... }` パターンがある場合 → `ON CONFLICT DO NOTHING` + 競合時の再取得パターンに変更する
   - 複数ユーザーが同じリソース（topics テーブル等）を共有する箇所の Create は必ず冪等にする

3. **best-effort の罠確認**
   - best-effort で何かを作成した後、後続の主操作が失敗した場合に「作成したもの」が残らないか確認する
   - 残ると業務影響がある場合（カウントに影響するログ等）は同一 tx に格上げする
   - 例: `aiGenLogRepo.Create`（best-effort）→ `genPostRepo.BulkCreate`（主操作）はどちらが失敗しても孤立しないよう同一 tx にする

4. **GORM CreateInBatches / Create の確認**
   - `CreateInBatches` または `Create` で DB 生成 UUID（`gen_random_uuid()`）が後続処理で必要な場合 → スライスは `[]*Model`（ポインタスライス）を使う
   - 値型スライス `[]Model` では DB 生成 UUID が entity に書き戻されない
