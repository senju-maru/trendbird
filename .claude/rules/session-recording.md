---
description: 即時記録プロトコル — CORRECTION/DECISION/PHILOSOPHYイベントをpending-learning.txtに記録するルール
---

会話中に以下の3種のイベントが発生したら、**そのターン内に** `.claude/retro/pending-learning.txt` へ追記すること。

## 記録対象と記録タイミング
| 種別 | トリガー | 例 |
|------|---------|-----|
| **CORRECTION** | ユーザーが Claude の提案を修正・却下した | 「autoPost じゃなくて canPublish を使って」 |
| **DECISION** | 設計・アーキテクチャ・ビジネス上の重要な判断が確定した | 「通知クールダウンは Free:1440分にしよう」 |
| **PHILOSOPHY** | ユーザーがプロダクトビジョン・方針を言語化した | 「開発者の当事者視点で書くべき」 |

## 記録フォーマット（pending-learning.txt）
```
## CORRECTION: [タイトル]
- proposed: [Claude が提案した内容]
- decided: [ユーザーの最終判断]
- category: UX / Architecture / Business / Scope / Other

## DECISION: [タイトル]
- content: [決定内容の要約]

## PHILOSOPHY
- content: [ユーザーが言語化した方針]
```

## 追加アクション（CORRECTION 発生時のみ）
1. MEMORY.md の該当セクションを同ターン内に更新する（承認不要）
2. スキルファイル更新が必要な場合はユーザーに提案する（**承認必要**）

## 運用ルール
- 記録は次のターンに持ち越さない。イベント発生ターン内で完了する
- 些細な修正（typo 指摘等）は記録不要。**判断・方針に関わるもの**のみ記録する
- `pending-learning.txt` は定期的に確認・クリアすること
