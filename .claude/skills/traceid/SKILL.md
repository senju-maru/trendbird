---
name: traceid
description: TraceID スキル。フロントエンド・バックエンド横断のリクエスト追跡仕様ガイド。新しいエンドポイント追加、ログ出力、エラーハンドリング変更時に自動参照される。
---

# TraceID スキル — リクエスト追跡仕様ガイド

TrendBird の全エンドポイントに適用される、フロントエンド・バックエンド横断のリクエスト追跡機構。

---

## 1. 仕様

| 項目 | 値 |
|------|-----|
| ヘッダー名 | `X-Trace-Id` |
| FE UUID バージョン | v4（`crypto.randomUUID()`） |
| BE UUID バージョン | v7（`github.com/gofrs/uuid/v5`、タイムスタンプベース） |
| 最大長 | 128 文字 |
| 禁止文字 | 制御文字（`unicode.IsControl(r)` が true のもの） |
| バリデーション失敗時 | BE 側で UUID v7 を自動生成して置き換え |

---

## 2. フロー全体図

```
[フロントエンド]
  generateTraceId() で UUID v4 生成
    ↓
  req.header.set('X-Trace-Id', traceId)  — Connect RPC interceptor
    ↓
[バックエンド]
  RecoveryInterceptor  ← panic 時はヘッダーから直接取得してログ出力
    ↓
  TraceIDInterceptor
    - ヘッダーから X-Trace-Id 読み取り
    - isValidTraceID() でバリデーション（長さ・制御文字）
    - 無効なら UUID v7 を新規生成
    - context に SetTraceID(ctx, traceID)
    ↓
  LoggingInterceptor
    - GetTraceID(ctx) で "trace_id" フィールドをログ出力
    ↓
  AuthInterceptor → Handler 処理
    ↓
  レスポンス:
    - 正常: resp.Header().Set("X-Trace-Id", traceID)
    - エラー: connectErr.Meta().Set("X-Trace-Id", traceID)
    ↓
[フロントエンド]
  エラー時: err.metadata.get('X-Trace-Id') → console.error でログ出力
```

---

## 3. 実装パターン: Connect RPC インターセプタ vs HTTP ミドルウェア

| パターン | 用途 | 適用方法 |
|---------|------|---------|
| **Connect RPC インターセプタ** (`NewTraceIDInterceptor`) | RPC エンドポイント | `connect.WithInterceptors()` で自動適用 |
| **HTTP ミドルウェア** (`TraceIDMiddleware`) | 非 RPC エンドポイント（healthz, OAuth, Webhook） | `router.go` でハンドラを個別ラップ |

**インターセプタの順序**（外側 → 内側）:
1. `RecoveryInterceptor` — panic をキャッチ（TraceID はヘッダーから直接取得）
2. `TraceIDInterceptor` — TraceID の検証・生成・context 格納
3. `LoggingInterceptor` — リクエストログに trace_id を付与
4. `AuthInterceptor` — JWT 認証

---

## 4. 関連ファイル一覧

### バックエンド

| ファイル | 役割 |
|---------|------|
| `backend/internal/adapter/middleware/context.go` | `SetTraceID()` / `GetTraceID()` — context への格納・取得 |
| `backend/internal/adapter/middleware/traceid.go` | `NewTraceIDInterceptor()` — Connect RPC 用インターセプタ、`isValidTraceID()` バリデーション |
| `backend/internal/adapter/middleware/traceid_http.go` | `TraceIDMiddleware()` — 非 RPC エンドポイント用 HTTP ミドルウェア |
| `backend/internal/adapter/middleware/traceid_test.go` | TraceID のユニットテスト（生成・伝播・バリデーション・エラーレスポンス） |
| `backend/internal/adapter/middleware/logging.go` | `NewLoggingInterceptor()` — ログに `trace_id` フィールドを含める |
| `backend/internal/adapter/middleware/recovery.go` | `NewRecoveryInterceptor()` — panic 時にヘッダーから trace_id を取得してログ出力 |
| `backend/internal/adapter/router/router.go` | インターセプタチェーン定義、HTTP ミドルウェアの個別適用 |
| `backend/internal/infrastructure/server/server.go` | CORS 設定（`AllowedHeaders` / `ExposedHeaders` に `X-Trace-Id`） |
| `backend/internal/di/container.go` | `TraceIDInterceptor` の DI 登録 |

### フロントエンド

| ファイル | 役割 |
|---------|------|
| `frontend/src/lib/trace.ts` | `TRACE_ID_HEADER` 定数、`generateTraceId()` 関数 |
| `frontend/src/lib/transport.ts` | `traceIdInterceptor` — 全リクエストに TraceID ヘッダーを付与 |
| `frontend/src/lib/connect-error.ts` | `connectErrorToMessage()` — エラーメタデータから TraceID を取得してコンソール出力 |

---

## 5. 今後の実装時チェックリスト

### 新しい RPC エンドポイント追加時
- **対応不要**。`router.go` の `connect.WithInterceptors()` でインターセプタチェーンが自動適用される。

### 新しい非 RPC エンドポイント追加時（Webhook, OAuth, ヘルスチェック等）
- `router.go` で `middleware.TraceIDMiddleware(handler)` でハンドラをラップすること。
- 例: `mux.Handle("POST /webhooks/new", middleware.TraceIDMiddleware(newHandler))`

### バックエンドでログ出力を追加する時
- `slog` の属性に `"trace_id", middleware.GetTraceID(ctx)` を含めること。
- ハンドラ・ユースケース・リポジトリのいずれの層でも同様。

### 新しい外部 API クライアント追加時（X API, OpenAI 等）
- エラーログに `"trace_id", middleware.GetTraceID(ctx)` を含めること。
- どのリクエストで外部 API エラーが発生したかを追跡可能にする。

### エラーハンドリング変更時
- Connect Error を返す箇所で、`connectErr.Meta().Set("X-Trace-Id", traceID)` が設定される仕組みを壊さないこと。
- TraceID の付与は `TraceIDInterceptor` が自動で行うため、ハンドラ側での手動設定は不要。

### フロントエンドでエラーハンドリング追加時
- `connectErrorToMessage()` を経由すれば TraceID が自動でコンソールにログされる。
- 独自のエラーハンドリングを追加する場合は、`err.metadata.get(TRACE_ID_HEADER)` で TraceID を取得・ログすること。

### Streaming RPC 追加時
- 現在の `NewTraceIDInterceptor` は `UnaryInterceptorFunc` のみ対応。
- Streaming RPC を追加する場合は `StreamingHandlerInterceptorFunc` 版の TraceID インターセプタを実装すること。

### `server.go` の CORS 設定変更時
- `AllowedHeaders` に `"X-Trace-Id"` を維持すること（FE → BE のヘッダー送信に必要）。
- `ExposedHeaders` に `"X-Trace-Id"` を維持すること（BE → FE のヘッダー読み取りに必要）。

---

## 6. Do's / Don'ts

### Do's
- **Do**: 新しい非 RPC エンドポイントには必ず `TraceIDMiddleware` を適用する
- **Do**: ログ出力時は `"trace_id"` フィールド名で統一する（`"traceId"` や `"trace-id"` は使わない）
- **Do**: エラーレスポンスにも TraceID を含める（ユーザーからの問い合わせ時に追跡可能にするため）
- **Do**: バリデーションに失敗した TraceID は黙って UUID v7 で置き換える（エラーを返さない）

### Don'ts
- **Don't**: ハンドラ内で TraceID を手動生成・設定しない（インターセプタ / ミドルウェアが自動で行う）
- **Don't**: TraceID をユーザーに見えるエラーメッセージに含めない（コンソールログのみ）
- **Don't**: TraceID をデータベースに保存しない（ログ追跡専用）
- **Don't**: `isValidTraceID()` のバリデーションルールを緩めない（ログインジェクション対策）
- **Don't**: インターセプタの順序を変えない（Recovery → TraceID → Logging → Auth の順序が重要）
