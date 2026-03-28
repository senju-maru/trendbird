---
name: backend-architecture
description: TrendBirdバックエンドのクリーンアーキテクチャ設計ガイド。Go + Connect RPC + GORM + PostgreSQL構成。バックエンドのコード追加・修正時に自動参照される。
---

# TrendBird バックエンド クリーンアーキテクチャガイド v1.0

**Go + Connect RPC + GORM + PostgreSQL**
最終更新: 2026-02-17

---

## 1. コンセプト

TrendBird バックエンドは **Clean Architecture** の4層構造を採用する。
依存の方向は **外→内の一方向のみ** とし、内側のレイヤーが外側を知ることは決してない。

```
┌──────────────────────────────────────────────┐
│  infrastructure（GORM, JWT, Config, Server）  │ ← 最も外側
│  ┌──────────────────────────────────────────┐ │
│  │  adapter（Handler, Converter, Presenter）│ │
│  │  ┌──────────────────────────────────────┐│ │
│  │  │  usecase（アプリケーションロジック） ││ │
│  │  │  ┌──────────────────────────────────┐││ │
│  │  │  │  domain（Entity, Repository IF） │││ │ ← 最も内側
│  │  │  └──────────────────────────────────┘││ │
│  │  └──────────────────────────────────────┘│ │
│  └──────────────────────────────────────────┘ │
└──────────────────────────────────────────────┘
```

**依存ルール:**

| レイヤー | 依存してよいもの | 依存してはいけないもの |
|----------|-----------------|----------------------|
| domain | 標準ライブラリのみ | GORM, Connect, protobuf, 外部パッケージ |
| usecase | domain | adapter, infrastructure |
| adapter | domain, usecase | infrastructure（の具体実装） |
| infrastructure | domain, usecase, adapter | — |

---

## 2. ディレクトリ構成

```
backend/
├── cmd/server/main.go                    # エントリポイント（config→DI→起動のみ）
├── internal/
│   ├── domain/                           # Layer 1: ドメイン（外部依存ゼロ）
│   │   ├── entity/                       # 純粋ドメインエンティティ（gormタグ無し）
│   │   │   ├── user.go
│   │   │   ├── topic.go
│   │   │   ├── post.go
│   │   │   └── ...
│   │   ├── repository/                   # リポジトリインターフェース（ポート）
│   │   │   ├── user.go
│   │   │   ├── topic.go
│   │   │   ├── post.go
│   │   │   ├── transaction.go            # TransactionManager インターフェース
│   │   │   └── ...
│   │   ├── gateway/                      # 外部APIインターフェース（ポート）
│   │   │   ├── ai.go                     # AIGenerator interface
│   │   │   └── twitter.go               # TwitterClient interface
│   │   ├── service/                      # ドメインサービス（複数エンティティのルール）
│   │   └── apperror/                     # ドメインエラー（connect/gorm非依存）
│   │       └── errors.go
│   ├── usecase/                          # Layer 2: ユースケース（domain のみ依存）
│   │   ├── auth.go
│   │   ├── topic.go
│   │   ├── post.go
│   │   ├── dashboard.go
│   │   ├── notification.go
│   │   ├── settings.go
│   │   └── twitter.go
│   ├── adapter/                          # Layer 3: インターフェースアダプター
│   │   ├── handler/                      # Connect RPCハンドラ（薄い変換層）
│   │   │   ├── auth.go
│   │   │   ├── topic.go
│   │   │   ├── post.go
│   │   │   └── ...
│   │   ├── converter/                    # entity ↔ proto 変換
│   │   │   ├── topic.go
│   │   │   ├── post.go
│   │   │   └── ...
│   │   ├── middleware/                   # ミドルウェアスタック
│   │   │   ├── auth.go                   # JWT検証、コンテキストにUserID設定
│   │   │   ├── context.go               # コンテキストヘルパー
│   │   │   ├── recovery.go              # パニック回復、500エラー返却
│   │   │   ├── traceid.go               # TraceID Connect RPCインターセプタ
│   │   │   ├── traceid_http.go          # TraceID HTTPミドルウェア（非RPC用）
│   │   │   └── logging.go              # リクエスト/レスポンスのslogログ
│   │   └── presenter/                    # ドメインエラー → Connectエラー変換
│   │       └── error.go
│   └── infrastructure/                   # Layer 4: フレームワーク＆ドライバ
│       ├── persistence/
│       │   ├── model/                    # GORMモデル（gormタグ付き）
│       │   │   ├── user.go
│       │   │   ├── topic.go
│       │   │   └── ...
│       │   ├── repository/              # リポジトリ実装（domain interface実装）
│       │   │   ├── user.go
│       │   │   ├── topic.go
│       │   │   └── ...
│       │   ├── mapper/                   # GORMモデル ↔ entity 変換
│       │   │   ├── user.go
│       │   │   ├── topic.go
│       │   │   └── ...
│       │   ├── db.go                     # DB接続
│       │   ├── migrate.go               # マイグレーション
│       │   └── transaction.go           # GORM TransactionManager 実装
│       ├── external/                     # 外部API実装（gateway インターフェースの実装）
│       │   ├── openai.go                # OpenAI API 実装
│       │   ├── twitter.go              # X API 実装
│       │   └── mock/                    # テスト用モック
│       ├── worker/                       # バックグラウンドジョブ（River）
│       │   ├── worker.go                # River client 初期化
│       │   ├── trend_detection.go       # TrendDetectionJob
│       │   ├── scheduled_post.go        # ScheduledPostJob
│       │   └── notification.go          # NotificationJob
│       ├── auth/                         # JWT実装
│       │   └── jwt.go
│       ├── config/                       # 設定管理（caarlos0/env）
│       │   └── config.go
│       └── server/                       # HTTPサーバー, CORS, Graceful Shutdown
│           └── server.go
│   └── testutil/                         # テスト共通ヘルパー
│       ├── db.go                         # SetupTestDB（testcontainers）, SetupTestTx
│       ├── server.go                     # SetupTestServer（httptest + Connect handlers）
│       ├── token.go                      # GenerateTestToken（JWT）
│       └── fixture.go                    # NewTestUser, NewTestTopic 等の Factory
├── di/                                   # DI Container
│   └── wire.go
├── e2e/                                  # E2E テスト（Connect client → 実DB）
│   ├── setup_test.go                     # TestMain（共有コンテナ起動）
│   ├── auth_test.go
│   ├── topic_test.go
│   └── post_test.go
├── gen/                                  # 生成済みprotobufコード（変更しない）
│   └── trendbird/v1/
├── migrations/                           # 本番用SQLマイグレーション（golang-migrate）
└── go.mod
```

---

## 3. 各レイヤーの責務と実装パターン

### 3.1 domain層（外部依存ゼロ）

domain層は **標準ライブラリ以外の import を一切持たない**。GORM, Connect, protobuf の型は使用禁止。

#### 3.1.1 entity — 純粋ドメインエンティティ

GORMタグ・JSONタグは付けない。ドメインの値とバリデーションロジックのみ持つ。

```go
// internal/domain/entity/user.go
package entity

import "time"

type User struct {
    ID                string
    TwitterID         string
    Name              string
    Email             string
    Image             string
    TwitterHandle     string
    TutorialCompleted bool
    CreatedAt         time.Time
    UpdatedAt         time.Time
}
```

```go
// internal/domain/entity/topic.go
package entity

import "time"

type TopicStatus int

const (
    TopicSpiking  TopicStatus = 1
    TopicRising   TopicStatus = 2
    TopicStable   TopicStatus = 3
    TopicDeclining TopicStatus = 4
)

type Topic struct {
    ID                  string
    UserID              string
    Name                string
    Keywords            []string
    Genre               string
    Status              TopicStatus
    ChangePercent       float64
    ZScore              *float64
    CurrentVolume       int32
    BaselineVolume      int32
    Context             *string
    ContextSummary      *string
    ContextKeywords     []string
    SpikeStartedAt      *time.Time
    NotificationEnabled bool
    CreatedAt           time.Time
    UpdatedAt           time.Time

    // 関連エンティティ（Preload相当）
    SparklineData []SparklineData
    SpikeHistory  []SpikeHistory
    PostingTip    *PostingTip
}
```

**ポイント:**
- enum は domain 独自の型として定義する（proto enum は使わない）
- `pq.StringArray` ではなく `[]string` を使う
- ポインタは「値がない」ことがドメイン上意味を持つフィールドのみに使用

#### 3.1.2 repository — インターフェース（ポート）

domain層で定義するインターフェース。infrastructure層が実装する。

```go
// internal/domain/repository/user.go
package repository

import (
    "context"

    "github.com/trendbird/backend/internal/domain/entity"
)

type UserRepository interface {
    FindByID(ctx context.Context, id string) (*entity.User, error)
    FindByTwitterHandle(ctx context.Context, handle string) (*entity.User, error)
    Create(ctx context.Context, user *entity.User) error
    Delete(ctx context.Context, id string) error
}
```

```go
// internal/domain/repository/topic.go
package repository

import (
    "context"

    "github.com/trendbird/backend/internal/domain/entity"
)

type TopicRepository interface {
    ListByUserID(ctx context.Context, userID string) ([]entity.Topic, error)
    FindByID(ctx context.Context, id string) (*entity.Topic, error)
    Create(ctx context.Context, topic *entity.Topic) error
    Delete(ctx context.Context, id string) error
    CountByUserID(ctx context.Context, userID string) (int64, error)
    CountDistinctGenresByUserID(ctx context.Context, userID string) (int64, error)
    ExistsGenreForUser(ctx context.Context, userID string, genre string) (bool, error)
    UpdateNotification(ctx context.Context, topicID string, enabled bool) error
}
```

**ポイント:**
- 全メソッドの第1引数に `context.Context` を渡す
- 戻り値はドメインエンティティのみ（GORMモデルではない）
- メソッド名は汎用的に（`FindByID` であり `FindUserByID` ではない）

#### 3.1.3 service — ドメインサービス

複数エンティティにまたがるドメインルール。外部依存なし（例: z-score 計算）。

#### 3.1.4 apperror — ドメインエラー

Connect や GORM に **一切依存しない** 純粋なエラー定義（`internal/domain/apperror/errors.go`）。

**エラーコード一覧:**

| コンストラクタ | 用途 |
|---------------|------|
| `NotFound(msg)` | リソース未発見 |
| `PermissionDenied(msg)` | 権限不足 |
| `InvalidArgument(msg)` | 入力バリデーション失敗 |
| `ResourceExhausted(msg)` | プラン制限到達 |
| `Unauthenticated(msg)` | 未認証 |
| `Internal(msg, err)` | 内部エラー（元エラーをラップ） |

**判定:** `IsNotFound(err)` で `errors.As` によるアンラップ判定が可能。

---

### 3.2 usecase層（アプリケーションビジネスロジック）

usecase は **domain 層のみに依存** する。リポジトリインターフェースをコンストラクタで受け取る。

```go
// internal/usecase/topic.go
package usecase

import (
    "context"
    "fmt"

    "github.com/trendbird/backend/internal/domain/apperror"
    "github.com/trendbird/backend/internal/domain/entity"
    "github.com/trendbird/backend/internal/domain/repository"
    "github.com/trendbird/backend/internal/domain/service"
)

type TopicUsecase struct {
    userRepo  repository.UserRepository
    topicRepo repository.TopicRepository
}

func NewTopicUsecase(
    userRepo repository.UserRepository,
    topicRepo repository.TopicRepository,
) *TopicUsecase {
    return &TopicUsecase{
        userRepo:  userRepo,
        topicRepo: topicRepo,
    }
}

func (uc *TopicUsecase) List(ctx context.Context, userID string) ([]entity.Topic, error) {
    return uc.topicRepo.ListByUserID(ctx, userID)
}

func (uc *TopicUsecase) Create(ctx context.Context, userID string, name string, keywords []string, genre string) (*entity.Topic, error) {
    topic := &entity.Topic{
        UserID:              userID,
        Name:                name,
        Keywords:            keywords,
        Genre:               genre,
        Status:              entity.TopicStable,
        NotificationEnabled: true,
    }

    if err := uc.topicRepo.Create(ctx, topic); err != nil {
        return nil, fmt.Errorf("create topic: %w", err)
    }

    return topic, nil
}
```

**ポイント:**
- ビジネスロジックはここに集約
- リポジトリはインターフェース経由で呼ぶ（テスト時にモック可能）
- エラーは `apperror` のドメインエラーか `fmt.Errorf` でラップ
- GORM, Connect, proto の型は一切登場しない

---

### 3.3 adapter層（Connect RPC ハンドラ + 変換）

adapter層は **domain + usecase に依存** する。フレームワーク固有の型（proto, Connect）をドメイン型に変換する薄い層。

#### 3.3.1 handler — Connect RPCハンドラ

ハンドラは「リクエスト解析 → usecase 呼出 → レスポンス変換」のみ。ビジネスロジックは書かない。
テンプレートは `references/code-templates.md` セクション4を参照。

**ハンドラの3ステップ:**
1. `middleware.UserIDFromContext(ctx)` でユーザーID取得 + `req.Msg.GetXxx()` でリクエスト値取得
2. `h.uc.Method(ctx, ...)` でユースケース呼出
3. `converter.XxxToProto(...)` で変換 + `presenter.ToConnectError(err)` でエラー変換

#### 3.3.2 converter — entity ↔ proto 変換

ドメインエンティティと proto メッセージの相互変換を担う（`internal/adapter/converter/`）。

**ルール:**
- `time.Time` → `timestamppb.New(t)` で変換
- ポインタフィールドは nil チェック後に変換
- enum はキャスト変換（`pb.TopicStatus(t.Status)`）
- 関数名: `XxxToProto(entity) *pb.Xxx` / `XxxStatusToEntity(pb) entity.XxxStatus`

#### 3.3.3 presenter — ドメインエラー → Connect エラー変換

apperror をフレームワーク固有のエラーに変換する唯一の場所。

```go
// internal/adapter/presenter/error.go
package presenter

import (
    "errors"

    "connectrpc.com/connect"

    "github.com/trendbird/backend/internal/domain/apperror"
)

func ToConnectError(err error) *connect.Error {
    if err == nil {
        return nil
    }

    var appErr *apperror.AppError
    if !errors.As(err, &appErr) {
        return connect.NewError(connect.CodeInternal, err)
    }

    var code connect.Code
    switch appErr.Code {
    case apperror.CodeNotFound:
        code = connect.CodeNotFound
    case apperror.CodePermissionDenied:
        code = connect.CodePermissionDenied
    case apperror.CodeInvalidArgument:
        code = connect.CodeInvalidArgument
    case apperror.CodeResourceExhausted:
        code = connect.CodeResourceExhausted
    case apperror.CodeUnauthenticated:
        code = connect.CodeUnauthenticated
    default:
        code = connect.CodeInternal
    }

    return connect.NewError(code, appErr)
}
```

---

### 3.4 infrastructure層（フレームワーク＆ドライバ）

最も外側のレイヤー。GORM, JWT, 環境変数など、外部技術の具体実装を置く。

#### 3.4.1 persistence/model — GORMモデル

GORM タグを持つ永続化専用の構造体。domain entity とは別の型。

```go
// internal/infrastructure/persistence/model/user.go
package model

import "time"

type User struct {
    ID                string  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
    TwitterID         string  `gorm:"type:varchar(64);uniqueIndex;not null"`
    Name              string  `gorm:"type:varchar(100);not null"`
    Email             string  `gorm:"type:varchar(255);not null;default:''"`
    Image             string  `gorm:"type:text;not null;default:''"`
    TwitterHandle     string `gorm:"type:varchar(15);uniqueIndex;not null"`
    TutorialCompleted bool   `gorm:"not null;default:false"`
    CreatedAt         time.Time
    UpdatedAt         time.Time
}
```

**ポイント:**
- JSON タグは不要（API レスポンスには proto を使うため）
- リレーションのカスケード削除は GORM タグで定義

#### 3.4.2 persistence/mapper — GORMモデル ↔ entity 変換

```go
// internal/infrastructure/persistence/mapper/user.go
package mapper

import (
    "github.com/trendbird/backend/internal/domain/entity"
    "github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

func UserToEntity(m *model.User) *entity.User {
    return &entity.User{
        ID:                m.ID,
        TwitterID:         m.TwitterID,
        Name:              m.Name,
        Email:             m.Email,
        Image:             m.Image,
        TwitterHandle:     m.TwitterHandle,
        TutorialCompleted: m.TutorialCompleted,
        CreatedAt:         m.CreatedAt,
        UpdatedAt:         m.UpdatedAt,
    }
}

func UserToModel(e *entity.User) *model.User {
    return &model.User{
        ID:                e.ID,
        TwitterID:         e.TwitterID,
        Name:              e.Name,
        Email:             e.Email,
        Image:             e.Image,
        TwitterHandle:     e.TwitterHandle,
        TutorialCompleted: e.TutorialCompleted,
        CreatedAt:         e.CreatedAt,
        UpdatedAt:         e.UpdatedAt,
    }
}
```

#### 3.4.3 persistence/repository — リポジトリ実装

domain のリポジトリインターフェースを GORM で実装する。

```go
// internal/infrastructure/persistence/repository/user.go
package repository

import (
    "context"
    "errors"

    "gorm.io/gorm"

    "github.com/trendbird/backend/internal/domain/apperror"
    "github.com/trendbird/backend/internal/domain/entity"
    domainrepo "github.com/trendbird/backend/internal/domain/repository"
    "github.com/trendbird/backend/internal/infrastructure/persistence/mapper"
    "github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

// インターフェース実装の保証
var _ domainrepo.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
    var m model.User
    if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, apperror.NotFound("user not found")
        }
        return nil, apperror.Internal("find user", err)
    }
    return mapper.UserToEntity(&m), nil
}

func (r *UserRepository) FindByTwitterHandle(ctx context.Context, handle string) (*entity.User, error) {
    var m model.User
    if err := r.db.WithContext(ctx).First(&m, "twitter_handle = ?", handle).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, apperror.NotFound("user not found")
        }
        return nil, apperror.Internal("find user by handle", err)
    }
    return mapper.UserToEntity(&m), nil
}

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
    m := mapper.UserToModel(user)
    if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
        return apperror.Internal("create user", err)
    }
    user.ID = m.ID
    user.CreatedAt = m.CreatedAt
    user.UpdatedAt = m.UpdatedAt
    return nil
}

func (r *UserRepository) Delete(ctx context.Context, id string) error {
    err := r.db.WithContext(ctx).
        Where("id = ?", id).
        Delete(&model.User{}).Error
    if err != nil {
        return apperror.Internal("delete user", err)
    }
    return nil
}
```

**ポイント:**
- `var _ domainrepo.UserRepository = (*UserRepository)(nil)` でインターフェース実装をコンパイル時に保証
- `gorm.ErrRecordNotFound` → `apperror.NotFound()` への変換はこの層で行う
- 全メソッドで `r.db.WithContext(ctx)` を使い、コンテキストを伝搬する
- 戻り値は常に `entity` 型（mapper で変換）

---

## 4. DI（依存性注入）

### Container パターン

```go
// di/wire.go
package di

import (
    "gorm.io/gorm"

    "github.com/trendbird/backend/internal/adapter/handler"
    domainrepo "github.com/trendbird/backend/internal/domain/repository"
    "github.com/trendbird/backend/internal/infrastructure/auth"
    "github.com/trendbird/backend/internal/infrastructure/config"
    persrepo "github.com/trendbird/backend/internal/infrastructure/persistence/repository"
    "github.com/trendbird/backend/internal/usecase"
)

type Container struct {
    // Handlers（サーバーが使う）
    AuthHandler         *handler.AuthHandler
    TopicHandler        *handler.TopicHandler
    PostHandler         *handler.PostHandler
    DashboardHandler    *handler.DashboardHandler
    NotificationHandler *handler.NotificationHandler
    SettingsHandler     *handler.SettingsHandler
    TwitterHandler      *handler.TwitterHandler
}

func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
    // Layer 4: Infrastructure
    userRepo := persrepo.NewUserRepository(db)
    topicRepo := persrepo.NewTopicRepository(db)
    postRepo := persrepo.NewPostRepository(db)
    // ... 他のリポジトリ

    jwtService := auth.NewJWTService(cfg.JWTSecret)

    // Layer 2: Usecases
    authUC := usecase.NewAuthUsecase(userRepo, jwtService)
    topicUC := usecase.NewTopicUsecase(userRepo, topicRepo)
    postUC := usecase.NewPostUsecase(postRepo, topicRepo)
    // ... 他のユースケース

    // Layer 3: Handlers
    return &Container{
        AuthHandler:      handler.NewAuthHandler(authUC),
        TopicHandler:     handler.NewTopicHandler(topicUC),
        PostHandler:      handler.NewPostHandler(postUC),
        // ... 他のハンドラ
    }
}
```

### main.go（エントリポイント）

```go
// cmd/server/main.go
package main

import (
    "github.com/trendbird/backend/di"
    "github.com/trendbird/backend/internal/infrastructure/config"
    "github.com/trendbird/backend/internal/infrastructure/persistence"
    "github.com/trendbird/backend/internal/infrastructure/server"
)

func main() {
    cfg := config.Load()

    db := persistence.Open(cfg)
    persistence.AutoMigrate(db)

    container := di.NewContainer(db, cfg)

    srv := server.New(cfg, container)
    srv.Run()
}
```

main.go の責務は **設定読込 → DB接続 → DI構築 → サーバー起動** の4行のみ。

---

## 5. エラーハンドリング

エラーは各レイヤーを以下の流れで伝搬する:

```
infrastructure (GORM error)
  → persistence/repository: gorm.ErrRecordNotFound → apperror.NotFound() に変換
    → usecase: apperror をそのまま返すか、fmt.Errorf でラップ
      → adapter/handler: presenter.ToConnectError(err) で Connect エラーに変換
        → クライアント
```

**各レイヤーの責務:**

| レイヤー | やること | やらないこと |
|----------|---------|-------------|
| persistence/repository | GORM エラー → apperror に変換 | Connect エラーを返す |
| usecase | apperror を生成・伝搬 | GORM / Connect に触れる |
| adapter/handler | `presenter.ToConnectError()` を呼ぶ | エラー変換ロジックを自前実装 |
| adapter/presenter | apperror.Code → connect.Code マッピング | ビジネスロジック判定 |

**ルール:**
- `gorm.ErrRecordNotFound` の判定は **persistence/repository 層のみ** で行う
- usecase 層は `apperror.IsNotFound(err)` で判定する（GORM を知らない）
- handler でのエラーハンドリングは常に `return nil, presenter.ToConnectError(err)` の1行

---

## 6. テスト戦略

### 6.1 テスト方針: E2E ファースト

Connect RPC + GORM + PostgreSQL の構成では、最も多いバグは **統合境界** で発生する:
- SQL クエリの誤り（WHERE 条件、JOIN、NULL ハンドリング）
- proto ↔ entity 変換の漏れ
- 認証ミドルウェアの適用漏れ
- apperror → Connect エラーのマッピングミス

handler は薄い変換層のため、モックで usecase をテストする価値は低い。
**実 DB + 実 HTTP の E2E テストを主軸** とし、純粋関数のみ単体テストで補完する。

```
テストピラミッド（TrendBird版）:

    ┌─────────┐
    │  単体   │  domain/entity, domain/service,
    │ テスト  │  converter, presenter, mapper
    ├─────────┤  （純粋関数・外部依存ゼロ）
    │         │
    │  E2E    │  Connect client → handler → usecase
    │ テスト  │  → repository → PostgreSQL（実DB）
    │         │
    └─────────┘
```

**モックを使うケース:** 外部 API（Twitter API, OpenAI API）のみ。DB・HTTP はモックしない。

### 6.2 テストインフラ構成

```
┌─────────────────────────────────────────────────────────┐
│  internal/e2etest/*_test.go                             │
│  ┌───────────────────┐    ┌──────────────────────────┐  │
│  │ Connect Client    │───▶│ httptest.Server          │  │
│  │ (生成済みclient)  │    │  ┌────────────────────┐  │  │
│  └───────────────────┘    │  │ AuthInterceptor    │  │  │
│                           │  ├────────────────────┤  │  │
│  ┌───────────────────┐    │  │ Handler (adapter)  │  │  │
│  │ seed ヘルパー群   │    │  ├────────────────────┤  │  │
│  │ (テストデータ投入)│    │  │ Usecase            │  │  │
│  └───────┬───────────┘    │  ├────────────────────┤  │  │
│          │                │  │ Repository (infra) │  │  │
│          ▼                │  └────────┬───────────┘  │  │
│  ┌───────────────────┐    └──────────┼───────────────┘  │
│  │ GORM (*gorm.DB)   │◀─────────────┘                  │
│  └───────┬───────────┘                                  │
│          ▼                                              │
│  ┌───────────────────┐                                  │
│  │ PostgreSQL        │  ← ローカル PostgreSQL 直接接続  │
│  │ (ローカルDB)      │    TestMain で接続 + AutoMigrate │
│  └───────────────────┘                                  │
└─────────────────────────────────────────────────────────┘
```

**重要:** Docker/testcontainers は使用しない（CLAUDE.md ルール: ローカル PostgreSQL 直接接続）。
テスト用 DSN: `postgres://localhost:5432/trendbird_test?sslmode=disable`（環境変数 `TEST_DATABASE_URL` で上書き可能）

### 6.3 レイヤー別テスト方針

| レイヤー | テスト種別 | 方法 | テスト内容 |
|----------|-----------|------|-----------|
| domain/entity | 単体テスト | 純粋関数テスト | バリデーション、enum メソッド |
| domain/service | 単体テスト | 純粋関数テスト | プラン制限ロジック |
| adapter/converter | 単体テスト | 純粋関数テスト | entity ↔ proto マッピング |
| adapter/presenter | 単体テスト | 純粋関数テスト | apperror → connect.Code マッピング |
| persistence/mapper | 単体テスト | 純粋関数テスト | model ↔ entity マッピング |
| usecase | **E2E で網羅** | Connect client → 実DB | ビジネスロジック全パターン |
| adapter/handler | **E2E で網羅** | Connect client → 実DB | リクエスト→レスポンス変換 |
| adapter/middleware | **E2E で網羅** | Connect client → 実DB | 認証・認可 |
| persistence/repository | **E2E で網羅** | Connect client → 実DB | SQL クエリの正しさ |

### 6.4 テストインフラ実装

テスト実装の詳細（testEnv, setupTest, seed ヘルパー, モックゲートウェイ, 完全なコード例）は **`/backend-e2e-test` スキル** を参照すること。

ここでは概要のみ記載:
- テストファイルは全て `backend/internal/e2etest/` に配置（`package e2etest`）
- `setupTest(t)` で truncateAll + モック生成 + DI + httptest.Server を構築
- seed ヘルパー群は Functional Options パターン（`seedUser(t, db, withEmail("test@example.com"))`）
- 外部 API のみモック（Fn フィールドパターン）。DB・HTTP はモックしない
- `connectClient[T]` ジェネリクスで認証付きクライアント生成
- `assertConnectCode` でエラーコード検証

### 6.8 テスト実行コマンド

```bash
# E2E テスト全実行（ローカル PostgreSQL 必須）
cd backend && go test ./internal/e2etest/... -v -race -count=1

# 特定のサービスのみ
cd backend && go test ./internal/e2etest/... -run TestAutoDMService -v -race -count=1

# 単体テストのみ（DB 不要、高速）
cd backend && go test ./internal/... -v -short

# usecase 単体テスト
cd backend && go test ./internal/usecase/... -v -race -count=1

# 全テスト（CI 向け、タイムアウト付き）
cd backend && go test ./... -v -timeout 300s
```

### 6.9 テスト設計チェックリスト

新しい RPC の E2E テストを追加する際は以下を確認:

- [ ] `success` ケース（正常系、レスポンス全体 + DB 副作用検証）
- [ ] `unauthenticated` ケース（認証なしクライアント → `CodeUnauthenticated`）
- [ ] `permission_denied` ケース（他ユーザーのリソース → `CodeNotFound` で情報漏洩防止）
- [ ] `not_found` ケース（存在しない ID → `CodeNotFound`）
- [ ] エッジケース（空リスト、上限到達、ステータス遷移）
- [ ] モック差し替え（外部 API 失敗時の挙動検証）

---

## 7. コード例テンプレート

新しい機能を追加する際のコピペ用テンプレート。
**詳細は `references/code-templates.md` を参照。** `Xxx` を機能名に置換して使用する。

追加が必要なファイル一覧:

| # | ファイル | レイヤー |
|---|---------|---------|
| 1 | `internal/domain/entity/xxx.go` | domain |
| 2 | `internal/domain/repository/xxx.go` | domain |
| 3 | `internal/usecase/xxx.go` | usecase |
| 4 | `internal/adapter/handler/xxx.go` | adapter |
| 5 | `internal/adapter/converter/xxx.go` | adapter |
| 6 | `internal/infrastructure/persistence/model/xxx.go` | infrastructure |
| 7 | `internal/infrastructure/persistence/mapper/xxx.go` | infrastructure |
| 8 | `internal/infrastructure/persistence/repository/xxx.go` | infrastructure |
| 9 | `di/wire.go` に登録 | DI |
| 10 | `internal/e2etest/xxx_test.go` | テスト（→ `/backend-e2e-test` スキル参照） |

---

## 8. 設定管理（Config）

`github.com/caarlos0/env/v11` を使い、構造体タグで環境変数をマッピングする。

```go
// internal/infrastructure/config/config.go
package config

import (
    "log"

    "github.com/caarlos0/env/v11"
)

type Config struct {
    Port            int    `env:"PORT" envDefault:"8080"`
    DatabaseURL     string `env:"DATABASE_URL,required"`
    JWTSecret       string `env:"JWT_SECRET,required"`
    CORSOrigins     string `env:"CORS_ORIGINS" envDefault:"http://localhost:3000"`

    // 外部API
    TwitterClientID     string `env:"TWITTER_CLIENT_ID"`
    TwitterClientSecret string `env:"TWITTER_CLIENT_SECRET"`
    OpenAIAPIKey        string `env:"OPENAI_API_KEY"`
}

func Load() *Config {
    cfg := &Config{}
    if err := env.Parse(cfg); err != nil {
        log.Fatalf("parse config: %v", err)
    }
    return cfg
}
```

**ポイント:**
- `required` タグで必須項目を宣言 — 未設定なら起動時に即エラー
- `envDefault` でデフォルト値を定義
- `os.Getenv` より型安全、Viper ほど大きくない

---

## 9. ロギング

`log/slog`（Go 標準ライブラリ）を使用する。依存ゼロ、構造化ログ（JSON 出力対応）。

**レイヤー別ログ配置ルール:**

| レイヤー | ログ内容 |
|----------|---------|
| middleware | リクエスト/レスポンスログ（method, path, status, duration） |
| usecase | ビジネスイベント（"topic spike detected", "post published"） |
| infrastructure | 外部API呼び出し結果、DBエラー |
| domain | **ログしない**（純粋関数） |

```go
// middleware でのリクエストログ例
slog.Info("request completed",
    "method", r.Method,
    "path", r.URL.Path,
    "status", status,
    "duration", time.Since(start),
    "trace_id", middleware.GetTraceID(r.Context()),
)

// usecase でのビジネスイベントログ例
slog.Info("topic spike detected",
    "topic_id", topic.ID,
    "z_score", zScore,
    "user_id", userID,
)

// infrastructure での外部APIログ例
slog.Error("openai api call failed",
    "error", err,
    "model", model,
    "trace_id", middleware.GetTraceID(ctx),
)
```

**ルール:**
- domain 層では `slog` を import しない
- `trace_id` を全ログに含め、リクエスト単位でトレース可能にする
- エラーログには `"error"` キーで元のエラーを含める

---

## 10. 外部APIクライアント（Gateway パターン）

外部API のインターフェースは `domain/gateway/` に定義し、具体実装は `infrastructure/external/` に置く。
usecase は gateway インターフェースにのみ依存する。

### 10.1 gateway インターフェース定義

```go
// internal/domain/gateway/ai.go
package gateway

import (
    "context"

    "github.com/trendbird/backend/internal/domain/entity"
)

type GeneratePostsInput struct {
    TopicName string
    Keywords  []string
    Context   string
    Style     entity.PostStyle
}

type GeneratePostsOutput struct {
    Content string
}

type AIGenerator interface {
    GeneratePosts(ctx context.Context, input GeneratePostsInput) ([]GeneratePostsOutput, error)
}
```

```go
// internal/domain/gateway/twitter.go
package gateway

import "context"

type PostTweetInput struct {
    Content string
    ReplyTo *string
}

type PostTweetOutput struct {
    TweetID string
    URL     string
}

type TwitterClient interface {
    PostTweet(ctx context.Context, accessToken string, input PostTweetInput) (*PostTweetOutput, error)
    GetTrendVolume(ctx context.Context, keywords []string) (map[string]int32, error)
}
```

### 10.2 具体実装（infrastructure/external）

```go
// internal/infrastructure/external/openai.go
package external

import (
    "context"
    "log/slog"

    "github.com/trendbird/backend/internal/domain/gateway"
)

var _ gateway.AIGenerator = (*OpenAIClient)(nil)

type OpenAIClient struct {
    apiKey string
}

func NewOpenAIClient(apiKey string) *OpenAIClient {
    return &OpenAIClient{apiKey: apiKey}
}

func (c *OpenAIClient) GeneratePosts(ctx context.Context, input gateway.GeneratePostsInput) ([]gateway.GeneratePostsOutput, error) {
    slog.Info("openai generate posts", "topic", input.TopicName)
    // OpenAI API 呼び出し実装
    // ...
    return nil, nil
}
```

**ポイント:**
- usecase は `gateway.AIGenerator` インターフェースに依存（実装を知らない）
- テスト時は `infrastructure/external/mock/` のモック実装を DI に注入
- `var _ gateway.Interface = (*Impl)(nil)` でコンパイル時チェック

---

## 11. トランザクション管理

domain にインターフェースを定義し、infrastructure で GORM 実装を提供する。

### 11.1 インターフェース

```go
// internal/domain/repository/transaction.go
package repository

import "context"

type TransactionManager interface {
    RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}
```

### 11.2 GORM 実装

```go
// internal/infrastructure/persistence/transaction.go
package persistence

import (
    "context"

    "gorm.io/gorm"

    domainrepo "github.com/trendbird/backend/internal/domain/repository"
)

type contextKey string

const txKey contextKey = "gorm_tx"

var _ domainrepo.TransactionManager = (*GORMTransactionManager)(nil)

type GORMTransactionManager struct {
    db *gorm.DB
}

func NewGORMTransactionManager(db *gorm.DB) *GORMTransactionManager {
    return &GORMTransactionManager{db: db}
}

func (tm *GORMTransactionManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
    return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        txCtx := context.WithValue(ctx, txKey, tx)
        return fn(txCtx)
    })
}

// DBFromContext は context から tx を取得する。tx がなければ通常の db を返す。
// 各リポジトリの内部で使用する。
func DBFromContext(ctx context.Context, db *gorm.DB) *gorm.DB {
    if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
        return tx.WithContext(ctx)
    }
    return db.WithContext(ctx)
}
```

### 11.3 リポジトリでの使用

```go
// リポジトリメソッド内で tx を自動取得するパターン
func (r *UserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
    db := persistence.DBFromContext(ctx, r.db)
    var m model.User
    if err := db.First(&m, "id = ?", id).Error; err != nil {
        // ...
    }
    return mapper.UserToEntity(&m), nil
}
```

**使用場面:**
- ユーザー削除のような複合操作（複数テーブルへの書き込み）
- 通常の単一 CRUD 操作では不要（GORM の自動トランザクションで十分）

### 11.4 TOCTOU 競合（Check-then-Create）の防止

`Find... → if nil { Create... }` パターンは並行リクエストで競合が起きる（TOCTOU: Time-Of-Check-To-Time-Of-Use）。
複数ユーザーが同じリソース（topics テーブル等）を共有する箇所では必ず `ON CONFLICT DO NOTHING` パターンを使う。

```go
// NG: TOCTOU 競合が起きる
existing, _ := repo.FindByName(ctx, name)
if existing == nil {
    repo.Create(ctx, entity) // 並行リクエストが同時にここに到達すると重複作成
}

// OK: ON CONFLICT DO NOTHING + 競合時の再取得
func (r *TopicRepository) Create(ctx context.Context, e *entity.Topic) error {
    db := persistence.DBFromContext(ctx, r.db)
    m := mapper.TopicToModel(e)
    result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&m)
    if result.Error != nil {
        return result.Error
    }
    if m.ID == "" {
        // 競合発生（別リクエストが先に INSERT した）→ 既存レコードを取得
        if err := db.Where("name = ?", e.Name).First(&m).Error; err != nil {
            return err
        }
    }
    e.ID = m.ID
    return nil
}
```

**適用基準:**
- 複数ユーザーが同じキー（name, (user_id, topic_id) 等）で同時に Create する可能性がある → 必須
- 単一ユーザーのみが書き込む（user_settings 等）→ 不要

### 11.5 best-effort 操作とアトミシティの判断フロー

「主操作」と「補助操作」を混在させると孤立レコードが生まれる。以下のフローで判断する。

```
補助操作（ログ・リンク等）が主操作の前にある場合
  ↓
「主操作が失敗したとき、補助操作の結果が残ると業務上問題があるか？」
  ├─ Yes（カウントに影響する・整合性が崩れる）→ 同一 tx に格上げ
  └─ No（残っても無害、次回フェッチで上書きされる等）→ best-effort でも可

補助操作（ログ・リンク等）が主操作の後にある場合
  ↓
「補助操作の失敗で主操作をロールバックすべきか？」
  ├─ Yes → 同一 tx に格上げ
  └─ No → best-effort（ただしエラーログは出力する）
```

**具体例:**

```go
// NG: aiGenLog（補助）が先に作成され、genPost（主）が失敗すると孤立する
logID, _ := aiGenLogRepo.Create(ctx, log)   // best-effort で作成
err := genPostRepo.BulkCreate(ctx, posts)    // 主操作が失敗 → logID だけが残る

// OK: 同一 tx でアトミックに処理
return txManager.RunInTx(ctx, func(ctx context.Context) error {
    logID, err := aiGenLogRepo.Create(ctx, log)
    if err != nil {
        return err
    }
    return genPostRepo.BulkCreate(ctx, posts) // 失敗時は log も一緒にロールバック
})
```

### 11.6 usecase の実装後チェックリスト

新しい usecase 関数を実装したら、PR 作成前に以下を確認する：

| # | チェック項目 | NG パターン |
|---|------------|------------|
| 1 | 2テーブル以上への書き込みは tx で囲んでいるか | `txManager.RunInTx` なし |
| 2 | Find→Create パターンに ON CONFLICT DO NOTHING を使っているか | `if existing == nil { Create }` |
| 3 | best-effort 補助操作が孤立しないか | tx の外で補助ログを先に作成 |
| 4 | `CreateInBatches` でポインタスライスを使っているか | `[]Model`（値型）で UUID 書き戻し漏れ |

---

## 12. ミドルウェアスタック

リクエスト処理の共通関心事をミドルウェアとして実装する。

### 12.1 適用順序

```
リクエスト → CORS → Recovery → TraceID → Logging → Auth → Handler
```

**Connect RPC エンドポイント**: インターセプターチェーンで処理（Recovery → TraceID → Logging → Auth）
**非 RPC エンドポイント（Webhook, OAuth, healthz）**: `TraceIDMiddleware`（HTTP ミドルウェア）を個別に適用

| ミドルウェア | 配置 | 役割 |
|------------|------|------|
| Recovery | `middleware/recovery.go` | パニック回復、500エラー返却 |
| TraceID | `middleware/traceid.go` | UUID v7生成、ヘッダー付与、ログ紐付け（Connect RPC用） |
| TraceID (HTTP) | `middleware/traceid_http.go` | 同上（Webhook/OAuth/healthz用） |
| Logging | `middleware/logging.go` | リクエスト/レスポンスのslogログ |
| CORS | `server/server.go` | `rs/cors` パッケージで設定 |
| Auth | `middleware/auth.go` | JWT検証、コンテキストにUserID設定 |

### 12.2 Recovery インターセプタ（Connect RPC 用）

```go
// internal/adapter/middleware/recovery.go
package middleware

// NewRecoveryInterceptor は panic をキャッチしてクライアントに Internal エラーを返すインターセプタ。
// TraceIDInterceptor の外側で動くため、traceID はリクエストヘッダーから直接取得する。
func NewRecoveryInterceptor() connect.UnaryInterceptorFunc {
    return func(next connect.UnaryFunc) connect.UnaryFunc {
        return func(ctx context.Context, req connect.AnyRequest) (resp connect.AnyResponse, retErr error) {
            defer func() {
                if r := recover(); r != nil {
                    slog.ErrorContext(ctx, "panic recovered",
                        "trace_id", req.Header().Get("X-Trace-Id"),
                        "panic", fmt.Sprintf("%v", r),
                        "procedure", req.Spec().Procedure,
                        "stack", string(debug.Stack()),
                    )
                    resp = nil
                    retErr = connect.NewError(connect.CodeInternal, nil)
                }
            }()
            return next(ctx, req)
        }
    }
}
```

### 12.3 TraceID インターセプタ（Connect RPC 用）

```go
// internal/adapter/middleware/traceid.go
package middleware

const traceIDHeader = "X-Trace-Id"
const maxTraceIDLen = 128

// isValidTraceID はクライアント供給の traceID が安全かを検証する。
func isValidTraceID(id string) bool {
    if len(id) == 0 || len(id) > maxTraceIDLen {
        return false
    }
    for _, r := range id {
        if unicode.IsControl(r) {
            return false
        }
    }
    return true
}

// NewTraceIDInterceptor はリクエストにトレースIDを付与するインターセプタ。
func NewTraceIDInterceptor() connect.UnaryInterceptorFunc {
    return func(next connect.UnaryFunc) connect.UnaryFunc {
        return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
            traceID := req.Header().Get(traceIDHeader)
            if !isValidTraceID(traceID) {
                traceID = uuid.Must(uuid.NewV7()).String()
            }
            ctx = SetTraceID(ctx, traceID)
            resp, err := next(ctx, req)
            if err != nil {
                if connectErr, ok := err.(*connect.Error); ok {
                    connectErr.Meta().Set(traceIDHeader, traceID)
                }
            } else if resp != nil {
                resp.Header().Set(traceIDHeader, traceID)
            }
            return resp, err
        }
    }
}
```

### 12.4 TraceID HTTP ミドルウェア（非 RPC 用）

```go
// internal/adapter/middleware/traceid_http.go
package middleware

// TraceIDMiddleware は非 RPC エンドポイント（Webhook, healthz, OAuth）用の HTTP ミドルウェア。
func TraceIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        traceID := r.Header.Get(traceIDHeader)
        if !isValidTraceID(traceID) {
            traceID = uuid.Must(uuid.NewV7()).String()
        }
        ctx := SetTraceID(r.Context(), traceID)
        w.Header().Set(traceIDHeader, traceID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 12.5 Logging インターセプタ

```go
// internal/adapter/middleware/logging.go
package middleware

// NewLoggingInterceptor はリクエスト/レスポンスの構造化ログを出力するインターセプタ。
func NewLoggingInterceptor() connect.UnaryInterceptorFunc {
    return func(next connect.UnaryFunc) connect.UnaryFunc {
        return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
            start := time.Now()
            resp, err := next(ctx, req)
            slog.InfoContext(ctx, "rpc completed",
                "procedure", req.Spec().Procedure,
                "duration", time.Since(start),
                "trace_id", GetTraceID(ctx),
            )
            return resp, err
        }
    }
}
```

### 12.6 サーバーでの組み立て

```go
// Connect RPC インターセプタチェーン（router.go）
interceptors := connect.WithInterceptors(
    c.RecoveryInterceptor,   // 1. 最外層（全ての panic をキャッチ）
    c.TraceIDInterceptor,    // 2. traceID 設定
    c.LoggingInterceptor,    // 3. リクエスト/レスポンスログ
    c.AuthInterceptor,       // 4. JWT認証
)

// 非 RPC エンドポイントには個別に TraceIDMiddleware を適用（router.go）
mux.Handle("GET /auth/x", middleware.TraceIDMiddleware(c.AuthHTTPHandler))

// server.go では TraceIDMiddleware をラップしない（二重生成を防止）
Handler: corsHandler.Handler(mux),
```

---

## 13. Graceful Shutdown

`signal.NotifyContext` で SIGINT/SIGTERM を受け取り、HTTP サーバーと DB 接続を安全に閉じる。

```go
// internal/infrastructure/server/server.go
package server

import (
    "context"
    "fmt"
    "log/slog"
    "net/http"
    "os/signal"
    "syscall"
    "time"

    "gorm.io/gorm"

    "github.com/trendbird/backend/internal/infrastructure/config"
)

type Server struct {
    httpServer *http.Server
    db         *gorm.DB
    cfg        *config.Config
}

func New(cfg *config.Config, handler http.Handler, db *gorm.DB) *Server {
    return &Server{
        httpServer: &http.Server{
            Addr:    fmt.Sprintf(":%d", cfg.Port),
            Handler: handler,
        },
        db:  db,
        cfg: cfg,
    }
}

func (s *Server) Run() {
    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    go func() {
        slog.Info("server starting", "port", s.cfg.Port)
        if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            slog.Error("server error", "error", err)
        }
    }()

    <-ctx.Done()
    slog.Info("shutting down...")

    shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
        slog.Error("server shutdown error", "error", err)
    }

    sqlDB, _ := s.db.DB()
    sqlDB.Close()

    slog.Info("server stopped")
}
```

---

## 14. Pagination

Offset + Limit 方式を採用する。TrendBird の規模（個人ユーザーのデータ量）では十分。

### 14.1 リポジトリ実装パターン

```go
// repository 層でのページネーション実装
func (r *PostHistoryRepository) ListByUserID(ctx context.Context, userID string, page, limit int32) ([]entity.PostHistory, int64, error) {
    db := persistence.DBFromContext(ctx, r.db)

    var total int64
    if err := db.Model(&model.PostHistory{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
        return nil, 0, apperror.Internal("count post histories", err)
    }

    var models []model.PostHistory
    offset := (page - 1) * limit
    if err := db.Where("user_id = ?", userID).
        Order("published_at DESC").
        Offset(int(offset)).Limit(int(limit)).
        Find(&models).Error; err != nil {
        return nil, 0, apperror.Internal("list post histories", err)
    }

    entities := make([]entity.PostHistory, len(models))
    for i, m := range models {
        entities[i] = *mapper.PostHistoryToEntity(&m)
    }
    return entities, total, nil
}
```

### 14.2 リポジトリインターフェース

```go
// ページネーション対応メソッドのシグネチャ
type PostHistoryRepository interface {
    ListByUserID(ctx context.Context, userID string, page, limit int32) ([]entity.PostHistory, int64, error)
    // page: 1始まり、limit: 1ページあたりの件数
    // 戻り値: エンティティスライス、総件数、エラー
}
```

### 14.3 handler でのレスポンス変換

```go
// proto の PaginationResponse にマッピング
resp := &pb.ListPostHistoriesResponse{
    PostHistories: pbHistories,
    Pagination: &pb.PaginationResponse{
        CurrentPage: req.Msg.GetPage(),
        TotalPages:  int32((total + int64(limit) - 1) / int64(limit)),
        TotalCount:  int32(total),
        HasMore:     int64(req.Msg.GetPage())*int64(limit) < total,
    },
}
```

---

## 15. バックグラウンドジョブ（River）

`github.com/riverqueue/river` を使用する。PostgreSQL をジョブキューとして活用し、追加インフラ不要。

### 15.1 ジョブ種別

| ジョブ | 頻度 | 内容 |
|--------|------|------|
| `TrendDetectionJob` | 5分ごと | X API でキーワード言及量取得 → z-score計算 → 状態更新 |
| `ScheduledPostJob` | 1分ごと | `scheduled_at` 到達の投稿を X API で実行 |
| `NotificationJob` | イベント駆動 | spike検知時に通知レコード作成 |

### 15.2 Worker 初期化

```go
// internal/infrastructure/worker/worker.go
package worker

import (
    "context"
    "log/slog"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/riverqueue/river"
    "github.com/riverqueue/river/riverdriver/riverpgxv5"
)

type Worker struct {
    client *river.Client[pgx.Tx]
}

func New(pool *pgxpool.Pool, workers *river.Workers) (*Worker, error) {
    client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
        Queues: map[string]river.QueueConfig{
            river.QueueDefault: {MaxWorkers: 10},
        },
        Workers: workers,
    })
    if err != nil {
        return nil, err
    }
    return &Worker{client: client}, nil
}

func (w *Worker) Start(ctx context.Context) error {
    slog.Info("worker starting")
    return w.client.Start(ctx)
}

func (w *Worker) Stop(ctx context.Context) error {
    w.client.Stop(ctx)
    slog.Info("worker stopped")
    return nil
}
```

### 15.3 ジョブ定義例

```go
// internal/infrastructure/worker/trend_detection.go
package worker

import (
    "context"
    "log/slog"

    "github.com/riverqueue/river"

    "github.com/trendbird/backend/internal/usecase"
)

type TrendDetectionArgs struct{}

func (TrendDetectionArgs) Kind() string { return "trend_detection" }

type TrendDetectionWorker struct {
    river.WorkerDefaults[TrendDetectionArgs]
    uc *usecase.TopicUsecase
}

func (w *TrendDetectionWorker) Work(ctx context.Context, job *river.Job[TrendDetectionArgs]) error {
    slog.Info("trend detection job started")
    // uc.DetectSpikes(ctx) を呼び出す
    return nil
}
```

### 15.4 定期実行の設定

River のスケジューリング機能を使い、cron のようにジョブを定期実行する。

```go
// River client の設定に PeriodicJobs を追加
periodicJobs := []*river.PeriodicJob{
    river.NewPeriodicJob(
        river.PeriodicInterval(5*time.Minute),
        func() (river.JobArgs, *river.InsertOpts) {
            return TrendDetectionArgs{}, nil
        },
        &river.PeriodicJobOpts{RunOnStart: true},
    ),
    river.NewPeriodicJob(
        river.PeriodicInterval(1*time.Minute),
        func() (river.JobArgs, *river.InsertOpts) {
            return ScheduledPostArgs{}, nil
        },
        nil,
    ),
}
```

**ポイント:**
- Worker は usecase を DI で受け取る（直接 repository を使わない）
- ジョブのエラーは River が自動リトライ（デフォルト最大25回、指数バックオフ）
- Worker の起動/停止は `cmd/server/main.go` の Graceful Shutdown と連携

---

## 16. ヘルスチェック

`/healthz` エンドポイントを Connect RPC 外の素の HTTP ハンドラとして提供する。

```go
// internal/infrastructure/server/server.go 内（mux への登録）
mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
    sqlDB, err := db.DB()
    if err != nil || sqlDB.Ping() != nil {
        http.Error(w, "unhealthy", http.StatusServiceUnavailable)
        return
    }
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
})
```

**ポイント:**
- DB への Ping で実際の疎通を確認する
- ロードバランサー / コンテナオーケストレーターのヘルスチェックに使用
- 認証不要（ミドルウェアスタックの外に配置）

---

## 17. Do's and Don'ts

### Do's（やること）

- **依存の方向を守る**: domain → usecase → adapter → infrastructure の順に依存。逆方向は禁止
- **handler は薄く保つ**: unmarshal → usecase 呼出 → marshal の3ステップのみ
- **repository は interface 経由で使う**: usecase からは必ずインターフェースを参照する
- **entity と model を分離する**: domain/entity は GORM タグなし、persistence/model は GORM タグ付き
- **エラーは apperror で統一する**: domain 層の純粋エラーを使い、Connect 変換は presenter のみ
- **context.Context を伝搬する**: 全レイヤーの public メソッドの第1引数に ctx を渡す
- **コンパイル時のインターフェース実装チェック**: `var _ Interface = (*Impl)(nil)` を書く
- **enum は domain 独自に定義する**: proto enum に依存しない domain enum を持ち、converter で変換
- **新規機能はセクション7のテンプレートに従う**: entity → repository IF → usecase → handler → model → mapper → repo impl → DI → E2E テストの順で追加
- **E2E テストを最初に書く**: 新機能の handler + repository を実装したら、E2E テストで全パスを検証する
- **testutil.SetupTestTx でテスト分離する**: 各テストでトランザクションを使い、自動ロールバックでデータを分離する
- **testify の require/assert を使い分ける**: 致命的な検証（err == nil 等）は `require`（失敗で即停止）、補助的な検証は `assert`
- **テストデータは fixture.go の Factory 関数で作る**: 直接 `db.Create()` せず `NewTestUser`, `NewTestTopic` を使う
- **設定は `caarlos0/env` の構造体タグで管理する**: `os.Getenv` 直呼びではなく `Config` 構造体経由でアクセス
- **ロギングは `log/slog` を使う**: domain 層以外で構造化ログを出力、`trace_id` を全ログに含める
- **外部API は gateway インターフェース経由で呼ぶ**: usecase から具体実装に依存しない
- **トランザクションが必要な複合操作では `TransactionManager.RunInTx` を使う**: リポジトリ内で `persistence.DBFromContext` で tx を取得
- **Connect RPC インターセプタは Recovery → TraceID → Logging → Auth の順で適用する**
- **非 RPC エンドポイントには `TraceIDMiddleware` を個別に適用する**（server.go で全体をラップしない）
- **Graceful Shutdown を実装する**: SIGINT/SIGTERM で HTTP サーバーと DB を安全に閉じる
- **ID 生成は UUID v7 を使う**: `uuid.Must(uuid.NewV7())` で時系列ソート可能な ID を生成
- **バックグラウンドジョブは River で実装する**: Worker は usecase を DI で受け取り、直接 repository を使わない
- **GORM で bool フィールドの `false` を確実に INSERT/UPDATE する**: `default:true` タグが付いた bool フィールドでは、`false` がゼロ値としてスキップされ DB デフォルト `TRUE` が使われる。対策として `map[string]any` で値を明示指定する: `db.Model(&m).Create(map[string]any{"spike_enabled": m.SpikeEnabled, ...})`。Upsert では `Clauses(clause.OnConflict{...}).Model(&m).Create(map[string]any{...})` を使う

### Don'ts（やらないこと）

- **handler にビジネスロジックを書かない**: プラン制限チェック、バリデーション、状態遷移は usecase に置く
- **domain 層に GORM/Connect/proto を import しない**: 標準ライブラリ以外は禁止
- **repository を自由関数にしない**: `func FindByID(db *gorm.DB, ...)` ではなく struct メソッド + interface
- **usecase で `*gorm.DB` を直接使わない**: repository interface 経由でのみデータにアクセスする
- **handler で `h.db.Model(...).Where(...)` と直接クエリを書かない**: 必ず usecase → repository 経由
- **apperror で connect パッケージを import しない**: Connect 変換は adapter/presenter の責務
- **model の JSON タグに依存しない**: API レスポンスは proto で返す。model は DB 専用
- **E2E テスト間でデータを共有しない**: 各テストは `SetupTestTx` で独立したトランザクションを使う
- **testcontainers のコンテナをテストごとに起動しない**: `TestMain` で1度だけ起動し、全テストで共有する
- **テストで `time.Sleep` を使わない**: DB の状態変化を待つ場合はリトライまたは直接確認する
- **DI を main.go に直書きしない**: di/wire.go の Container に集約する
- **domain 層で `slog` を import しない**: ロギングは middleware / usecase / infrastructure で行う
- **`os.Getenv` で設定を直接取得しない**: 必ず `Config` 構造体経由
- **usecase から gateway の具体実装に依存しない**: `infrastructure/external/*` を直接 import しない
- **リポジトリ内で `r.db` を直接使わずトランザクション対応する場合は `persistence.DBFromContext(ctx, r.db)` を使う**
- **Worker（バックグラウンドジョブ）から直接 repository を呼ばない**: usecase 経由でのみデータにアクセスする
- **ヘルスチェックに認証ミドルウェアを適用しない**: `/healthz` はミドルウェアスタックの外に配置
- **`default:true` の bool フィールドを `db.Create(&struct)` で保存しない**: ゼロ値の `false` がスキップされる。`map[string]any` を使うこと

---

## 18. バッチジョブ設計指針

`backend/cmd/batch/main.go` で実行されるバッチジョブの設計ルール。ローカルでは `make batch-run JOB=<name>` で手動実行、または `make scheduler` でスケジュール実行する。

### 18.1 タイムアウト設計

タイムアウトはジョブの**特性**に合わせて個別に設定する。デフォルト値で一律に設定してはいけない。

```go
// backend/cmd/batch/main.go
const defaultBatchTimeout = 10 * time.Minute  // 通常ジョブのデフォルト

// ジョブ種別ごとのタイムアウト（外部 API 呼び出し数 × 処理時間で見積もる）
var jobTimeouts = map[string]time.Duration{
    "trend-fetch":            25 * time.Minute, // 複数トピック × X API 呼び出し
    "source-post-collection": 25 * time.Minute, // 同上
    // 通常ジョブは defaultBatchTimeout を使用
}
```

**見積もり方法:**
1. 外部 API 呼び出し数を数える（トピック数 × API コール数）
2. 1呼び出しあたりの最大待ち時間（レート制限含む）を掛ける
3. 安全係数 × 1.5 を掛けて余裕を持たせる
4. `defaultBatchTimeout` の2倍を超えるなら個別設定を追加する

**NG パターン:**
```go
// NG: 全ジョブを同じタイムアウトで実行（外部 API 呼び出しが多いジョブでタイムアウトする）
timeout := defaultBatchTimeout

// OK: ジョブ特性に応じて切り替え
timeout := defaultBatchTimeout
if t, ok := jobTimeouts[jobName]; ok {
    timeout = t
}
```

### 18.2 新規バッチジョブ追加時のチェックリスト

新しいジョブを `main.go` に追加するとき：

| # | チェック項目 |
|---|------------|
| 1 | 外部 API 呼び出しを含む場合、`jobTimeouts` に個別タイムアウトを追加したか |
| 2 | ジョブが冪等か（同じジョブを2回実行しても整合性が保たれるか） |
| 3 | ジョブ失敗時にリトライしても安全か（DB への副作用を考慮） |
| 4 | ローカルスケジューラの `scheduler.go` にジョブを登録したか |
