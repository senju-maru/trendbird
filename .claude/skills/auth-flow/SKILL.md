---
name: auth-flow
description: TrendBirdの認証フロー統合ガイド。X OAuth 2.0 PKCE→バックエンドJWT→フロントエンドセッション管理の一連のフローを実装する際に自動参照される。Cookie設定・コールバック・ルートガード・セッション復元をカバー。
---

# TrendBird 認証フロー統合ガイド v1.1

**X OAuth 2.0 PKCE → JWT HttpOnly Cookie → Zustand セッション管理**
最終更新: 2026-02-28

---

## 1. 概要・目的

本ガイドは、TrendBird の認証フロー全体を **フロントエンドからバックエンドまで横断的に** カバーする。

**カバー範囲:**

- X OAuth 2.0 PKCE フロー全体のオーケストレーション
- JWT の HttpOnly Cookie 設定・取得
- フロントエンドのセッション管理（Zustand authStore）
- コールバックページの実装
- ルートガード（middleware.ts）
- ログアウト・セッション復元

**他スキルへの委譲:**

| 関心事 | 本ガイド | 委譲先 |
|--------|---------|--------|
| PKCE code_verifier/challenge 生成 | 参照リンク | `/x-api` セクション 7.9 |
| OAuth スコープ・トークン仕様 | 参照リンク | `/x-api` セクション 2 |
| JWT 発行・検証の内部実装 | 参照リンク | `/backend-architecture` セクション 10, 12.4 |
| AuthInterceptor の詳細 | 参照リンク | `/backend-architecture` セクション 12 |
| Token リフレッシュの X API 仕様 | エンドポイント参照 | `/x-api` セクション 2.1 |

---

## 1.5. リダイレクト・セッション切れ仕様（ベースライン）

**この仕様はログイン・ログアウト・セッション管理のベースラインである。**

> **⚠️ 厳守事項（Claude への指示）:**
> 以下の仕様は確定済みのベースラインであり、**ユーザーから明示的に変更指示がない限り、絶対に変更してはならない。**
> 「改善」「リファクタリング」「UX向上」等の名目であっても、ユーザーの明示的な承認なしにリダイレクト先やLP表示ロジックを変更することを禁止する。
>
> **禁止事項:**
> - ログアウト・セッション切れ時のリダイレクト先を `/login` や他のパスに変更すること（ベースラインは `/`）
> - LP のCTAボタン文言やリンク先を認証状態で条件分岐させること（常に同一表示）
> - LP に `cookies()` を追加して動的レンダリングにすること
> - ログインページから MarketingHeader を削除すること
> - ミドルウェアの `/login` → `/dashboard` リダイレクトロジックを削除すること
> - 上記に関連するコードを「不要」と判断して削除すること
>
> これらの変更が必要と判断した場合は、**必ずユーザーに理由を説明し、承認を得てから実施すること。**

### 1.5.1 リダイレクト方針: 原則 LP（`/`）に遷移

セッション切れ・ログアウト・未認証アクセス時はすべて **LP（`/`）** にリダイレクトする。`/login` に直接飛ばさない。

| シナリオ | リダイレクト先 | 実装箇所 |
|---------|-------------|---------|
| ログアウト実行 | `/` | `Sidebar.tsx`, `Header.tsx`, `settings/page.tsx` |
| セッション復元失敗（JWT期限切れ等） | `/` | `(app)/layout.tsx` |
| 未認証で保護ルートにアクセス | `/` | `middleware.ts` |
| ログイン済みで `/login` にアクセス | `/dashboard` | `middleware.ts` |

**理由:** ユーザーがセッション切れ後にいきなりログインボタンだけの画面に飛ばされるのはUXが悪い。LPに戻せばサービスの全体像を見た上で自分のタイミングでログインできる。

### 1.5.2 LP の表示: 認証状態で文言を変えない

LP（`/`）のCTAボタン（「ログイン」「無料で始める」等）は **ログイン状態に関わらず常に同じ文言を表示** する。ログイン済みユーザーが「ログイン」や「無料で始める」を押した場合は、ミドルウェアが `/login` → `/dashboard` にリダイレクトするため、シームレスにダッシュボードに遷移する。

- LP は `cookies()` を使わず **静的レンダリング（Static）** を維持する
- 認証状態による条件分岐は LP コンポーネントに入れない

### 1.5.3 ログインページ（`/login`）

- **MarketingHeader を表示** する（LPに戻れるようにするため）
- ログイン済みユーザーがアクセスした場合、ミドルウェアで `/dashboard` にリダイレクト
- `(auth)/layout.tsx` に `MarketingHeader` と `AmbientBackground` を配置

### 1.5.4 マーケティングページ（`/pricing`, `/terms` 等）

- MarketingHeader は常に「ログイン」「無料で始める」を表示（認証状態で変えない）
- ログイン済みユーザーがヘッダーの「ログイン」を押す → ミドルウェアで `/dashboard` へ

### 1.5.5 ミドルウェア（`middleware.ts`）の責務

```
1. 静的アセット・API ルート → スキップ
2. /login + tb_jwt Cookie あり → /dashboard へリダイレクト
3. 公開パス（/, /login, /callback, /pricing, /terms, /privacy, /contact） → 通過
4. 保護ルート + tb_jwt Cookie なし → / へリダイレクト
5. それ以外 → 通過
```

---

## 2. 設計判断

### 2.1 NextAuth.js を使わない理由

フロントエンドの `package.json` に NextAuth.js は含まれているが、**設定ファイルは作成せず、使用しない。**

**理由:**

1. **X 専用の単一プロバイダ**: NextAuth.js のマルチプロバイダ抽象化が過剰
2. **バックエンドが JWT 発行主体**: NextAuth.js のセッション管理とバックエンド JWT の責任範囲が競合する
3. **Connect RPC との統合**: NextAuth.js は REST API 前提であり、Connect RPC とのシームレスな統合が困難

**結論:** OAuth フローはバックエンド主導で実装し、フロントエンドは単純なリダイレクト + Cookie ベースのセッション管理とする。

### 2.2 JWT の格納場所: HttpOnly Cookie

| 方式 | XSS耐性 | CSRF耐性 | 採用 |
|------|---------|---------|------|
| HttpOnly Cookie + SameSite=Strict | ○ | ○ | **採用** |
| localStorage | × | — | — |
| sessionStorage | × | — | — |
| メモリ（Zustand） | ○ | — | トークン保持に不適 |

- `authStore` は `User` 情報のみを保持し、**JWT トークンは保持しない**
- JWT の送受信はすべてブラウザの Cookie 自動送信に委ねる

### 2.3 code_verifier の一時保存

OAuth フロー中の `code_verifier` は **暗号化した HttpOnly Cookie** に一時保存する。

| 項目 | 値 |
|------|-----|
| Cookie 名 | `tb_cv` |
| 有効期限 | 10 分 |
| HttpOnly | Yes |
| Secure | Yes（本番） |
| SameSite | Lax（OAuth リダイレクトを許可するため） |
| 暗号化 | AES-GCM（COOKIE_ENCRYPTION_KEY 使用） |

> `SameSite=Lax` を使用する理由: X の認可画面からのリダイレクト（top-level navigation）で Cookie を送信する必要があるため。`Strict` ではリダイレクト時に Cookie が送信されない。

### 2.4 X Token 管理

X の access_token / refresh_token は **バックエンドの DB にのみ保存** する。フロントエンドには一切渡さない。

```
┌──────────┐      ┌──────────────┐      ┌──────────┐
│ Frontend │      │   Backend    │      │    DB    │
│          │      │              │      │          │
│ JWT のみ │◀────▶│ JWT 検証     │      │ X tokens │
│ (Cookie) │      │ → UserID取得 │─────▶│ 暗号化保存│
│          │      │ → X token取得│◀─────│          │
└──────────┘      └──────────────┘      └──────────┘
```

---

## 3. Clean Architecture 位置付け

### バックエンド

```
adapter/handler/       auth_handler.go   ← XAuth RPC, Logout, GetCurrentUser
adapter/middleware/     auth.go           ← JWT 検証 interceptor (既存)
usecase/               auth.go           ← AuthUsecase (OAuth→Upsert→JWT)
domain/entity/         user.go           ← User エンティティ
domain/repository/     user.go           ← UserRepository interface
infrastructure/auth/   jwt.go            ← JWT 発行・検証 (既存)
infrastructure/persistence/repository/
                       user.go           ← UserRepository GORM 実装
infrastructure/persistence/model/
                       user.go           ← User GORM モデル
                       twitter_connection.go ← X token 保存
```

### フロントエンド

```
app/(auth)/login/page.tsx       ← ログインページ（X認証開始）
app/(auth)/callback/page.tsx    ← OAuthコールバック（新規作成）
stores/authStore.ts             ← セッション状態管理（既存）
lib/transport.ts                ← Connect RPC transport（credentials追加）
middleware.ts                   ← ルートガード（新規作成）
```

---

## 4. 認証フロー全体シーケンス

### 4.1 ログインフロー

```
User          Frontend              Backend                   X API
 │              │                     │                         │
 │ [1] ログイン  │                     │                         │
 │  ボタンクリック│                     │                         │
 │─────────────▶│                     │                         │
 │              │ [2] window.location  │                         │
 │              │   = /auth/x         │                         │
 │              │────────────────────▶│                         │
 │              │                     │ [3] code_verifier生成    │
 │              │                     │   code_challenge = S256  │
 │              │                     │   code_verifier → 暗号化Cookie(tb_cv)
 │              │                     │   state → Cookie(tb_state)
 │              │  ← 302 redirect     │                         │
 │              │◀────────────────────│                         │
 │              │                     │                         │
 │ [4] X 認可画面│════════════════════════════════════════════▶│
 │  ユーザーが許可│                     │                         │
 │              │◀═══ ?code=xxx&state=yyy ═══════════════════│
 │              │                     │                         │
 │              │ [5] callback page   │                         │
 │              │   code 取得         │                         │
 │              │   XAuth RPC 呼出    │                         │
 │              │────────────────────▶│                         │
 │              │                     │ [6] code_verifier復元(tb_cv)
 │              │                     │   state検証(tb_state)    │
 │              │                     │   POST /2/oauth2/token  │
 │              │                     │────────────────────────▶│
 │              │                     │  ← access_token         │
 │              │                     │    refresh_token        │
 │              │                     │                         │
 │              │                     │ [7] User upsert         │
 │              │                     │   X tokens → DB暗号化保存│
 │              │                     │   JWT発行 → Set-Cookie  │
 │              │  ← User + JWT Cookie│                         │
 │              │◀────────────────────│                         │
 │              │                     │                         │
 │              │ authStore.setUser   │                         │
 │              │ router.push('/dashboard')                     │
```

### 4.2 認証済みリクエストフロー

すべての Connect RPC リクエストに Cookie を自動送信する。

```
Frontend                    Backend
   │                          │
   │  Connect RPC request     │
   │  Cookie: tb_jwt=xxx      │
   │─────────────────────────▶│
   │                          │ AuthInterceptor:
   │                          │   Cookie から JWT 取得
   │                          │   JWT 検証 → UserID 抽出
   │                          │   ctx に UserID 設定
   │                          │
   │  ← Response              │
   │◀─────────────────────────│
```

### 4.3 セッション復元フロー（ページリロード）

```
Frontend                    Backend
   │                          │
   │ [ページロード]             │
   │ authStore.isLoading=true │
   │                          │
   │  GetCurrentUser RPC      │
   │  Cookie: tb_jwt=xxx      │
   │─────────────────────────▶│
   │                          │ JWT 検証 → UserID
   │                          │ User 取得
   │  ← User data             │
   │◀─────────────────────────│
   │                          │
   │ authStore.setUser(user)  │
   │ authStore.isLoading=false│
```

### 4.4 ログアウトフロー

```
Frontend                    Backend
   │                          │
   │  Logout RPC              │
   │  Cookie: tb_jwt=xxx      │
   │─────────────────────────▶│
   │                          │ Set-Cookie: tb_jwt=""; Max-Age=0
   │  ← Response + 削除Cookie │
   │◀─────────────────────────│
   │                          │
   │ authStore.logout()       │
   │ router.push('/')         │  ← LP に遷移（/login ではない）
```

---

## 5. バックエンド実装パターン

### 5.1 OAuth 開始: GET /auth/x

Connect RPC ではなく **標準 HTTP ハンドラ** として実装する（ブラウザリダイレクトが必要なため）。

```go
// internal/adapter/handler/auth_http.go
package handler

import (
    "net/http"

    "github.com/trendbird/backend/internal/usecase"
)

// AuthHTTPHandler は OAuth リダイレクト用の HTTP ハンドラ。
// Connect RPC とは別に net/http の ServeMux に登録する。
type AuthHTTPHandler struct {
    authUC *usecase.AuthUsecase
}

func NewAuthHTTPHandler(authUC *usecase.AuthUsecase) *AuthHTTPHandler {
    return &AuthHTTPHandler{authUC: authUC}
}

// HandleOAuthStart は X OAuth 認可画面へリダイレクトする。
// GET /auth/x
func (h *AuthHTTPHandler) HandleOAuthStart(w http.ResponseWriter, r *http.Request) {
    result, err := h.authUC.StartOAuth(r.Context())
    if err != nil {
        http.Error(w, "failed to start oauth", http.StatusInternalServerError)
        return
    }

    // code_verifier を暗号化 Cookie に保存
    http.SetCookie(w, &http.Cookie{
        Name:     "tb_cv",
        Value:    result.EncryptedCodeVerifier,
        Path:     "/",
        MaxAge:   600, // 10分
        HttpOnly: true,
        Secure:   r.TLS != nil,
        SameSite: http.SameSiteLaxMode,
    })

    // state を Cookie に保存
    http.SetCookie(w, &http.Cookie{
        Name:     "tb_state",
        Value:    result.State,
        Path:     "/",
        MaxAge:   600,
        HttpOnly: true,
        Secure:   r.TLS != nil,
        SameSite: http.SameSiteLaxMode,
    })

    http.Redirect(w, r, result.AuthorizationURL, http.StatusFound)
}
```

**main.go での登録:**

```go
// main.go
mux := http.NewServeMux()

// Connect RPC ハンドラ
mux.Handle(trendbirdv1connect.NewAuthServiceHandler(container.AuthHandler, interceptors))
// ...

// OAuth HTTP ハンドラ（Connect RPC の外）
authHTTP := handler.NewAuthHTTPHandler(container.AuthUsecase)
mux.HandleFunc("GET /auth/x", authHTTP.HandleOAuthStart)
```

### 5.2 XAuth RPC ハンドラ

Connect RPC の `XAuth` RPC でトークン交換→ユーザー Upsert → JWT 発行 → Set-Cookie を行う。

```go
// internal/adapter/handler/auth_handler.go
func (h *AuthHandler) XAuth(
    ctx context.Context,
    req *connect.Request[trendbirdv1.XAuthRequest],
) (*connect.Response[trendbirdv1.XAuthResponse], error) {
    // Cookie から code_verifier と state を復元
    codeVerifier, err := h.getCodeVerifierFromCookie(ctx)
    if err != nil {
        return nil, presenter.ToConnectError(apperror.Unauthenticated("invalid oauth state"))
    }

    result, err := h.authUC.XAuth(ctx, usecase.XAuthInput{
        OAuthCode:    req.Msg.OauthCode,
        CodeVerifier: codeVerifier,
    })
    if err != nil {
        return nil, presenter.ToConnectError(err)
    }

    // JWT を HttpOnly Cookie に設定
    resp := connect.NewResponse(converter.ToXAuthResponse(result))
    resp.Header().Set("Set-Cookie", h.buildJWTCookie(result.JWT).String())

    // code_verifier Cookie を削除
    resp.Header().Add("Set-Cookie", h.clearCookie("tb_cv").String())
    resp.Header().Add("Set-Cookie", h.clearCookie("tb_state").String())

    return resp, nil
}
```

### 5.3 AuthUsecase.XAuth() ビジネスロジック

```go
// internal/usecase/auth.go
type XAuthInput struct {
    OAuthCode    string
    CodeVerifier string
}

type XAuthOutput struct {
    User             *entity.User
    JWT              string
    AIGenerationUsed int32
}

func (u *AuthUsecase) XAuth(ctx context.Context, input XAuthInput) (*XAuthOutput, error) {
    // 1. X API でトークン交換
    tokens, err := u.twitterClient.ExchangeToken(ctx, input.OAuthCode, input.CodeVerifier)
    if err != nil {
        return nil, fmt.Errorf("exchange token: %w", err)
    }

    // 2. X API でユーザー情報取得
    xUser, err := u.twitterClient.GetMe(ctx, tokens.AccessToken)
    if err != nil {
        return nil, fmt.Errorf("get x user: %w", err)
    }

    // 3. User upsert（新規登録 or 既存更新）
    user, err := u.userRepo.UpsertByTwitterID(ctx, entity.UpsertUserInput{
        TwitterID:     xUser.ID,
        TwitterHandle: xUser.Username,
        Name:          xUser.Name,
        Image:         xUser.ProfileImageURL,
    })
    if err != nil {
        return nil, fmt.Errorf("upsert user: %w", err)
    }

    // 4. X tokens を DB に保存（暗号化）
    if err := u.twitterConnRepo.Upsert(ctx, entity.TwitterConnection{
        UserID:       user.ID,
        AccessToken:  tokens.AccessToken,
        RefreshToken: tokens.RefreshToken,
        ExpiresAt:    tokens.ExpiresAt,
    }); err != nil {
        return nil, fmt.Errorf("save twitter connection: %w", err)
    }

    // 5. JWT 発行
    jwt, err := u.jwtService.GenerateToken(user.ID)
    if err != nil {
        return nil, fmt.Errorf("generate jwt: %w", err)
    }

    return &XAuthOutput{
        User: user,
        JWT:  jwt,
    }, nil
}
```

### 5.4 JWT Cookie セキュリティ設定

```go
const (
    jwtCookieName = "tb_jwt"
    jwtMaxAge     = 7 * 24 * 60 * 60 // 7日間（秒）
)

// buildJWTCookie は JWT を HttpOnly Cookie として構築する。
func (h *AuthHandler) buildJWTCookie(token string) *http.Cookie {
    return &http.Cookie{
        Name:     jwtCookieName,
        Value:    token,
        Path:     "/",
        MaxAge:   jwtMaxAge,
        HttpOnly: true,
        Secure:   true,           // 本番は常に HTTPS
        SameSite: http.SameSiteStrictMode,
    }
}

// clearCookie は Cookie を削除する。
func (h *AuthHandler) clearCookie(name string) *http.Cookie {
    return &http.Cookie{
        Name:     name,
        Value:    "",
        Path:     "/",
        MaxAge:   -1,
        HttpOnly: true,
        Secure:   true,
        SameSite: http.SameSiteStrictMode,
    }
}
```

| 属性 | 値 | 理由 |
|------|-----|------|
| `HttpOnly` | `true` | JavaScript からのアクセスを防止（XSS 対策） |
| `Secure` | `true` | HTTPS 接続でのみ送信 |
| `SameSite` | `Strict` | クロスサイトリクエストで送信されない（CSRF 対策） |
| `Path` | `/` | 全パスで有効 |
| `MaxAge` | 604800 (7日) | JWT の有効期限と一致 |

> **注意:** JWT Cookie は `SameSite=Strict`。OAuth リダイレクト用の `tb_cv` / `tb_state` Cookie は `SameSite=Lax`（X からのリダイレクトで必要なため）。

### 5.5 AuthInterceptor の Cookie 取得パターン

既存の `AuthInterceptor` は `Authorization` ヘッダーから JWT を取得するが、Cookie からも取得できるようにする。

```go
// internal/adapter/middleware/auth.go
func (i *AuthInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
    return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
        // パブリック RPC はスキップ
        if isPublicRPC(req.Spec().Procedure) {
            return next(ctx, req)
        }

        // 1. Authorization ヘッダーから取得（Connect RPC 標準）
        token := extractBearerToken(req.Header().Get("Authorization"))

        // 2. Cookie から取得（ブラウザ向け）
        if token == "" {
            token = extractTokenFromCookie(req.Header().Get("Cookie"))
        }

        if token == "" {
            return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("missing auth token"))
        }

        userID, err := i.jwtService.ValidateToken(token)
        if err != nil {
            return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid token"))
        }

        ctx = context.WithValue(ctx, userIDKey{}, userID)
        return next(ctx, req)
    }
}

// extractTokenFromCookie は Cookie ヘッダーから JWT を取得する。
func extractTokenFromCookie(cookieHeader string) string {
    header := http.Header{}
    header.Add("Cookie", cookieHeader)
    request := http.Request{Header: header}
    cookie, err := request.Cookie("tb_jwt")
    if err != nil {
        return ""
    }
    return cookie.Value
}

// isPublicRPC は認証不要な RPC を判定する。
func isPublicRPC(procedure string) bool {
    publicRPCs := []string{
        "/trendbird.v1.AuthService/XAuth",
    }
    for _, rpc := range publicRPCs {
        if procedure == rpc {
            return true
        }
    }
    return false
}
```

### 5.6 TwitterConnection GORM モデル

X の access_token / refresh_token を暗号化して保存する。

```go
// internal/infrastructure/persistence/model/twitter_connection.go
package model

import "time"

type TwitterConnection struct {
    ID                string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
    UserID            string    `gorm:"uniqueIndex;type:uuid;not null"`
    AccessToken       string    `gorm:"type:text;not null"`       // 暗号化して保存
    RefreshToken      string    `gorm:"type:text;not null"`       // 暗号化して保存
    TokenExpiresAt    time.Time `gorm:"not null"`
    CreatedAt         time.Time
    UpdatedAt         time.Time
}
```

### 5.7 Access Token リフレッシュ戦略

X の access_token は **2 時間で失効** する。X API を呼び出す前にトークンの有効期限をチェックし、必要に応じてリフレッシュする。

```go
// internal/usecase/twitter.go
func (u *TwitterUsecase) getValidAccessToken(ctx context.Context, userID string) (string, error) {
    conn, err := u.twitterConnRepo.GetByUserID(ctx, userID)
    if err != nil {
        return "", fmt.Errorf("get twitter connection: %w", err)
    }

    // 有効期限の5分前にリフレッシュ（余裕を持たせる）
    if time.Now().Before(conn.TokenExpiresAt.Add(-5 * time.Minute)) {
        return conn.AccessToken, nil
    }

    // リフレッシュ
    tokens, err := u.twitterClient.RefreshToken(ctx, conn.RefreshToken)
    if err != nil {
        return "", fmt.Errorf("refresh token: %w", err)
    }

    // DB 更新
    if err := u.twitterConnRepo.Upsert(ctx, entity.TwitterConnection{
        UserID:       userID,
        AccessToken:  tokens.AccessToken,
        RefreshToken: tokens.RefreshToken,
        ExpiresAt:    tokens.ExpiresAt,
    }); err != nil {
        return "", fmt.Errorf("update twitter connection: %w", err)
    }

    return tokens.AccessToken, nil
}
```

> **参照:** リフレッシュの X API 仕様は `/x-api` セクション 2.1 を参照

---

## 6. フロントエンド実装パターン

### 6.1 transport.ts: credentials: 'include' 追加

Connect RPC リクエストに Cookie を自動送信するため、`credentials: 'include'` を追加する。

```typescript
// frontend/src/lib/transport.ts
import type { Transport } from '@connectrpc/connect';

const USE_MOCKS = process.env.NEXT_PUBLIC_USE_MOCKS === 'true';

let cachedTransport: Transport | null = null;

export async function getTransport(): Promise<Transport> {
  if (cachedTransport) return cachedTransport;

  if (USE_MOCKS) {
    const { createMockTransport } = await import('@/api/mock/mock-transport');
    cachedTransport = createMockTransport();
  } else {
    const { createConnectTransport } = await import('@connectrpc/connect-web');
    cachedTransport = createConnectTransport({
      baseUrl: process.env.NEXT_PUBLIC_API_URL ?? '',
      credentials: 'include',  // ← Cookie を自動送信
    });
  }

  return cachedTransport;
}
```

### 6.2 login/page.tsx: OAuth リダイレクト

モックの `setTimeout` を実際の OAuth フローに置き換える。

```typescript
// frontend/src/app/(auth)/login/page.tsx
const handleXLogin = () => {
  setAuthError(null);
  setIsXLoading(true);
  // バックエンドの OAuth 開始エンドポイントにリダイレクト
  window.location.href = `${process.env.NEXT_PUBLIC_API_URL}/auth/x`;
};
```

> `router.push` ではなく `window.location.href` を使用する。バックエンドが 302 リダイレクトで X の認可画面に遷移させるため、SPA ルーティングでは処理できない。

### 6.3 callback/page.tsx: 新規作成

X の認可画面からのリダイレクト先。URL パラメータから `code` を取得し、`XAuth` RPC を呼び出す。

```typescript
// frontend/src/app/(auth)/callback/page.tsx
'use client';

import { Suspense, useEffect, useState } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import { useAuthStore } from '@/stores/authStore';
import { getTransport } from '@/lib/transport';
import { createClient } from '@connectrpc/connect';
import { AuthService } from '@/gen/trendbird/v1/auth_connect';

function CallbackContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const setUser = useAuthStore((s) => s.setUser);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const code = searchParams.get('code');
    const state = searchParams.get('state');

    if (!code) {
      // ユーザーが認証を拒否した場合
      const errorParam = searchParams.get('error');
      if (errorParam === 'access_denied') {
        router.push('/login');
        return;
      }
      setError('認証コードが取得できませんでした');
      return;
    }

    const authenticate = async () => {
      try {
        const transport = await getTransport();
        const client = createClient(AuthService, transport);
        const res = await client.xAuth({ oauthCode: code });

        setUser({
          id: res.user!.id,
          name: res.user!.name,
          email: res.user!.email,
          image: res.user!.image,
          twitterHandle: res.user!.twitterHandle,
          planId: res.user!.planId as any,
          createdAt: res.user!.createdAt,
        });

        router.push('/dashboard');
      } catch (err) {
        console.error('Authentication failed:', err);
        setError('認証に失敗しました。もう一度お試しください。');
      }
    };

    authenticate();
  }, [searchParams, router, setUser]);

  if (error) {
    return (
      <div style={{ textAlign: 'center', padding: 40 }}>
        <p style={{ color: '#ef4444', marginBottom: 16 }}>{error}</p>
        <a href="/login">ログインページに戻る</a>
      </div>
    );
  }

  return (
    <div style={{ textAlign: 'center', padding: 40 }}>
      <p>認証処理中...</p>
    </div>
  );
}

export default function CallbackPage() {
  return (
    <Suspense>
      <CallbackContent />
    </Suspense>
  );
}
```

### 6.4 セッション復元（getCurrentUser）

アプリのルートレイアウトまたは `(app)` レイアウトで、ページ読み込み時にセッションを復元する。

```typescript
// frontend/src/app/(app)/layout.tsx 内の useEffect
'use client';

import { useEffect } from 'react';
import { useAuthStore } from '@/stores/authStore';
import { getTransport } from '@/lib/transport';
import { createClient } from '@connectrpc/connect';
import { AuthService } from '@/gen/trendbird/v1/auth_connect';

export default function AppLayout({ children }: { children: React.ReactNode }) {
  const setUser = useAuthStore((s) => s.setUser);

  useEffect(() => {
    const restore = async () => {
      try {
        const transport = await getTransport();
        const client = createClient(AuthService, transport);
        const res = await client.getCurrentUser({});

        setUser({
          id: res.user!.id,
          name: res.user!.name,
          email: res.user!.email,
          image: res.user!.image,
          twitterHandle: res.user!.twitterHandle,
          planId: res.user!.planId as any,
          createdAt: res.user!.createdAt,
        });
      } catch {
        // JWT 無効 or 期限切れ → LP へ
        window.location.href = '/';
      }
    };

    restore();
  }, [setUser]);

  return <>{children}</>;
}
```

### 6.5 middleware.ts: ルートガード

Next.js Middleware で Cookie の存在のみチェックする。**JWT の検証はバックエンドに委譲** する（middleware はEdge Runtimeで実行されるため、JWT 検証ライブラリが使えない場合がある）。

```typescript
// frontend/src/middleware.ts
import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

// 認証不要のパス
const publicPaths = ['/login', '/callback', '/', '/pricing'];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // パブリックパスはスキップ
  if (publicPaths.some((p) => pathname === p || pathname.startsWith(p + '/'))) {
    return NextResponse.next();
  }

  // 静的アセット・API ルートはスキップ
  if (pathname.startsWith('/_next') || pathname.startsWith('/api')) {
    return NextResponse.next();
  }

  // ログイン済みユーザーが /login にアクセス → /dashboard へリダイレクト
  if (pathname === '/login') {
    const token = request.cookies.get('tb_jwt');
    if (token) {
      return NextResponse.redirect(new URL('/dashboard', request.url));
    }
  }

  // JWT Cookie の存在チェック（検証はバックエンドに委譲）
  const token = request.cookies.get('tb_jwt');
  if (!token) {
    return NextResponse.redirect(new URL('/', request.url));  // LP へリダイレクト
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/((?!_next/static|_next/image|favicon.ico).*)'],
};
```

---

## 7. エラーハンドリング

### OAuth フロー中のエラーパターン

| エラー | 発生箇所 | 対応 |
|--------|---------|------|
| ユーザーが認証拒否 | X 認可画面 → callback | `error=access_denied` → ログインページへ遷移 |
| state 不一致 | Backend (XAuth) | `Unauthenticated` → ログインページへ遷移 |
| code 期限切れ（30秒超過） | Backend (token exchange) | `InvalidArgument` → エラー表示 + ログインページへリンク |
| X API 障害 | Backend (token exchange) | `Internal` → エラー表示 + リトライ促進 |
| JWT 期限切れ | Backend (AuthInterceptor) | `Unauthenticated` → LP（`/`）へ自動遷移 |
| Refresh token 無効化 | Backend (token refresh) | `Unauthenticated` → LP（`/`）へ遷移、再ログイン要求 |

### Unauthenticated 自動ハンドリング

Connect RPC の `Unauthenticated` エラーを共通ハンドリングし、自動的にログインページへ遷移する。

```typescript
// frontend/src/lib/api-error-handler.ts
import { ConnectError, Code } from '@connectrpc/connect';

export function handleApiError(error: unknown): never {
  if (error instanceof ConnectError && error.code === Code.Unauthenticated) {
    // セッション切れ → LP へ
    window.location.href = '/';
  }
  throw error;
}
```

---

## 8. 環境変数

### バックエンド

| 変数名 | 必須 | デフォルト | 説明 |
|--------|------|-----------|------|
| `JWT_SECRET` | Yes | — | JWT 署名用シークレットキー（既存） |
| `JWT_EXPIRY` | No | `168h` | JWT 有効期限（Go duration 形式、7日） |
| `X_CLIENT_ID` | Yes | — | X OAuth App の Client ID |
| `X_CLIENT_SECRET` | Yes | — | X OAuth App の Client Secret |
| `X_REDIRECT_URI` | Yes | — | OAuth コールバック URL（例: `https://api.trendbird.app/auth/x/callback`） |
| `COOKIE_ENCRYPTION_KEY` | Yes | — | code_verifier Cookie の AES-GCM 暗号化キー（32バイト） |

**config.go への追加:**

```go
// 既存の Config struct に追加:
JWTExpiry            time.Duration `env:"JWT_EXPIRY" envDefault:"168h"`
XClientID            string        `env:"X_CLIENT_ID,required"`
XClientSecret        string        `env:"X_CLIENT_SECRET,required"`
XRedirectURI         string        `env:"X_REDIRECT_URI,required"`
CookieEncryptionKey  string        `env:"COOKIE_ENCRYPTION_KEY,required"`
```

### フロントエンド

| 変数名 | 必須 | 説明 |
|--------|------|------|
| `NEXT_PUBLIC_API_URL` | Yes | バックエンド API URL（例: `http://localhost:8080`） |

> **注意:** `NEXTAUTH_*` 環境変数は使用しない（NextAuth.js 不使用のため）。`package.json` に NextAuth.js が残っていても設定ファイルは作成しないこと。
