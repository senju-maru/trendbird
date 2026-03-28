---
name: tutorial-onboarding
description: チュートリアル・オンボーディングガイド。driver.jsベースのマルチページオンボーディングチュートリアル。チュートリアル機能の追加・修正時に自動参照される。
---

# チュートリアル・オンボーディングガイド

## 1. 概要

driver.js ベースのマルチページオンボーディングチュートリアル。新規ユーザーが初回ログイン後に5つのフェーズ（Welcome → トピック選択 → ダッシュボード説明 → カードクリック体験 → 詳細画面説明）を通じてプロダクトの主要機能を学べるようガイドする。

### フェーズ設計

| Phase | ページ | 内容 | ステップ数 |
|-------|--------|------|-----------|
| A: Welcome | `/dashboard`（オーバーレイ） | プロダクト紹介 | 1（要素なしセンターポップオーバー） |
| B: TopicSetup | `/topics` | ジャンル選択誘導 | 1 + ユーザー操作待ち + フローティングCTA |
| C: Dashboard | `/dashboard` | ダミーデータでUI説明 | 3（status-bar, genre-tabs, topic-card） |
| D: Navigate | `/dashboard` | カードクリック誘導 | 1（ユーザーの実クリック待ち） |
| E: Detail | `/dashboard/tutorial-dummy` | ダミーデータで詳細画面説明 | 3（detail-hero, detail-context, detail-ai） |

### ページ遷移フロー

```
[OAuth ログイン]
  → /dashboard（Phase A: Welcome オーバーレイ）
  → 「次へ」→ router.push('/topics')
  → /topics（Phase B: ジャンル選択誘導）
  → ユーザーがジャンル＋トピック追加
  → フローティングCTA「ダッシュボードを見てみよう →」クリック
  → /dashboard（Phase C: ダミーデータでUI説明 → Phase D: カードクリック誘導）
  → ユーザーがカードをクリック
  → /dashboard/tutorial-dummy（Phase E: ダミーデータで詳細画面説明）
  → 「完了」→ チュートリアル終了、/dashboard にリダイレクト
```

## 2. isNewUser フラグのフロー

```
Backend (auth.go:98)
  time.Since(user.CreatedAt) < 5*time.Second → isNewUser 判定
    ↓
Proto (auth.proto:53)
  XAuthResponse.is_new_user フィールド
    ↓
Frontend (useAuth.ts)
  xAuth(): res.isNewUser → authStore.setState({ isNewUser })
           + sessionStorage.setItem('tb_tutorial_pending', '1')
    ↓
Frontend (authStore.ts)
  isNewUser: boolean（Zustand ストア）
    ↓
Frontend (OnboardingTutorial.tsx)
  useAuthStore((s) => s.isNewUser) を監視 → useTutorialState.start() → フェーズ開始
```

### セッション復元時

`GetCurrentUserResponse` には `isNewUser` が含まれないため、`getCurrentUser()` 内で `sessionStorage` の `tb_tutorial_pending` を確認し、存在すれば `isNewUser: true` を再設定してツアーを継続する。

## 3. ストレージキー

| キー | 保存先 | 用途 | 設定タイミング |
|------|--------|------|---------------|
| `tb_tutorial_completed` | localStorage | ツアー完了済みフラグ | Phase E 完了時 |
| `tb_tutorial_pending` | sessionStorage | ツアー未完了フラグ | xAuth() 成功時 |
| `tb_tutorial_state` | sessionStorage | フェーズ進捗管理（JSON） | 各フェーズ遷移時 |

- `tb_tutorial_pending` はツアー完了時に削除
- `tb_tutorial_completed` はツアー完了時に `'1'` を設定
- `tb_tutorial_state` は `{ phase, topicId? }` 形式の JSON

## 4. data-tutorial 属性（セレクタ規約）

ツアーステップは `data-tutorial` 属性をセレクタとして使用する。新しいステップを追加する場合もこの規約に従うこと。

| Phase | セレクタ | 対象 | 設定ファイル |
|-------|---------|------|------------|
| B | `data-tutorial="genre-select-cta"` | 「ジャンルを選ぶ」ボタン | `topics/page.tsx` |
| C | `data-tutorial="status-bar"` | モニタリング状況バー | `dashboard/page.tsx` |
| C | `data-tutorial="genre-tabs"` | ジャンルフィルタータブ | `dashboard/page.tsx` |
| C/D | `data-tutorial="topic-card"` | 第1トピックカード | `dashboard/page.tsx` → TopicCard props |
| E | `data-tutorial="detail-hero"` | ヒーローカード（盛り上がり度） | `dashboard/[id]/page.tsx` |
| E | `data-tutorial="detail-context"` | コンテキストカード（バズ理由） | `dashboard/[id]/page.tsx` |
| E | `data-tutorial="detail-ai"` | AI投稿文生成セクション | `dashboard/[id]/page.tsx` |
| - | `data-tutorial="sidebar-nav"` | サイドバーナビゲーション（未使用） | `Sidebar.tsx` |

### 注意事項

- `topic-card` は第1カードのみに付与（`i === 0` の条件分岐）
- 各フェーズは対応する要素が DOM に出現するまで最大10秒待機
- 要素描画完了後、500ms の待機を経てツアーを開始

## 5. ダミーデータモード

Phase C/D/E ではチュートリアル用のダミーデータ（固定値）を使用し、APIアクセスを行わない。

### 仕組み

- `useTutorialState` フックの `isTutorialMode` + `phase` でダッシュボード・詳細画面のデータソースを切り替え
- `dashboard/page.tsx`: `isDummyMode ? TUTORIAL_DASHBOARD : apiData` パターン
- `dashboard/[id]/page.tsx`: `isDummyMode ? TUTORIAL_DETAIL : realTopic` パターン
- 詳細画面のダミーデータは**スパイク状態**で固定（最もリッチな画面を見せるため）

### ダミーデータファイル

`frontend/src/components/tutorial/tutorialDummyData.ts` に定義:
- `TUTORIAL_DUMMY_TOPIC`: ダッシュボード・詳細画面共用のトピックデータ
- `TUTORIAL_DASHBOARD`: ステータスカウント + トピック一覧
- `TUTORIAL_AI_POSTS`: AI生成投稿文のサンプル
- `TUTORIAL_TRENDING_POSTS`: トレンド投稿のサンプル

### ダミーURL

- `/dashboard/tutorial-dummy` — チュートリアルモード中のみアクセス可能
- チュートリアル外でアクセスすると `/dashboard` にリダイレクト

## 6. エッジケース対応

| ケース | 対応 |
|--------|------|
| 途中でブラウザを閉じる | sessionStorage 消失 → 次回ログインでチュートリアル再実行されない |
| 途中でリロード | sessionStorage 維持 → 現在のフェーズから再開 |
| フェーズとパスの不一致 | ツアー非起動、正しいページに戻った時に再開 |
| `/dashboard/tutorial-dummy` に直アクセス | isTutorialMode チェック → 非チュートリアル時は `/dashboard` にリダイレクト |

## 7. テスト方法

`useAuth.ts` の `xAuth()` 内で `isNewUser` をハードコードすると、毎回チュートリアルが起動する:

```typescript
// useAuth.ts の xAuth() 内（テスト用に一時変更）
useAuthStore.setState({ aiGenerationUsed: res.aiGenerationUsed, isNewUser: true });
//                                                               ^^^^^^^^^ res.isNewUser → true
```

**テスト後は必ず `res.isNewUser` に戻すこと。**

また、localStorage の `tb_tutorial_completed` を削除すれば、完了済みユーザーでも再度ツアーを確認できる:

```javascript
// ブラウザの DevTools Console で実行
localStorage.removeItem('tb_tutorial_completed');
sessionStorage.removeItem('tb_tutorial_state');
```

## 8. スタイリング

`tutorial.css` でニューモルフィズムテーマを適用。driver.js デフォルトスタイルを上書きしている。

- オーバーレイ: 半透明 `rgba(190, 202, 214, 0.6)`
- ポップオーバー: `#e4eaf1` 背景 + ボーダーラディウス 20px
- ボタン: グラデーション青 (`#41b1e1` → `#7acbee`)
- フローティングCTA: グラデーション青 + ボックスシャドウ + スライドアップアニメーション
- `prefers-reduced-motion` 対応済み

## 9. 関連ファイル一覧

### バックエンド
- `backend/internal/usecase/auth.go` - isNewUser 判定ロジック
- `backend/internal/adapter/handler/auth.go` - IsNewUser をレスポンスに設定

### Proto
- `proto/trendbird/v1/auth.proto` - `is_new_user` フィールド定義

### フロントエンド（チュートリアルコア）
- `frontend/src/components/tutorial/OnboardingTutorial.tsx` - フェーズオーケストレータ
- `frontend/src/components/tutorial/OnboardingTutorialLoader.tsx` - 動的インポートラッパー（SSR無効）
- `frontend/src/components/tutorial/useTutorialState.ts` - sessionStorage ベース状態管理フック
- `frontend/src/components/tutorial/tutorialDummyData.ts` - ダミーデータ定義
- `frontend/src/components/tutorial/TutorialFloatingCta.tsx` - Phase B→C のフローティングCTA
- `frontend/src/components/tutorial/tutorial.css` - ニューモルフィズムテーマ + フローティングCTA

### フロントエンド（フェーズ定義）
- `frontend/src/components/tutorial/phases/welcomePhase.ts` - Phase A
- `frontend/src/components/tutorial/phases/topicSetupPhase.ts` - Phase B
- `frontend/src/components/tutorial/phases/dashboardPhase.ts` - Phase C
- `frontend/src/components/tutorial/phases/navigatePhase.ts` - Phase D
- `frontend/src/components/tutorial/phases/detailPhase.ts` - Phase E

### フロントエンド（修正対象ページ）
- `frontend/src/hooks/useAuth.ts` - isNewUser ストア設定・sessionStorage 管理
- `frontend/src/stores/authStore.ts` - isNewUser 状態管理（Zustand）
- `frontend/src/app/(app)/layout.tsx` - OnboardingTutorialLoader のマウント
- `frontend/src/app/(app)/dashboard/page.tsx` - ダミーデータモード対応 + data-tutorial 属性
- `frontend/src/app/(app)/dashboard/[id]/page.tsx` - ダミーデータモード対応 + data-tutorial 属性
- `frontend/src/app/(app)/topics/page.tsx` - data-tutorial="genre-select-cta" + フローティングCTA
- `frontend/src/components/layout/Sidebar.tsx` - data-tutorial="sidebar-nav"
