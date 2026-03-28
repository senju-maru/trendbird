# コード例テンプレート

新しい機能を追加する際のコピペ用テンプレート。`Xxx` を機能名に置換して使用する。

## 1. エンティティ追加

```go
// internal/domain/entity/xxx.go
package entity

import "time"

type Xxx struct {
    ID        string
    UserID    string
    // ... フィールド
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

## 2. リポジトリインターフェース追加

```go
// internal/domain/repository/xxx.go
package repository

import (
    "context"

    "github.com/trendbird/backend/internal/domain/entity"
)

type XxxRepository interface {
    FindByID(ctx context.Context, id string) (*entity.Xxx, error)
    ListByUserID(ctx context.Context, userID string) ([]entity.Xxx, error)
    Create(ctx context.Context, xxx *entity.Xxx) error
    Update(ctx context.Context, xxx *entity.Xxx) error
    Delete(ctx context.Context, id string) error
}
```

## 3. ユースケース追加

```go
// internal/usecase/xxx.go
package usecase

import (
    "context"
    "fmt"

    "github.com/trendbird/backend/internal/domain/entity"
    "github.com/trendbird/backend/internal/domain/repository"
)

type XxxUsecase struct {
    xxxRepo repository.XxxRepository
}

func NewXxxUsecase(xxxRepo repository.XxxRepository) *XxxUsecase {
    return &XxxUsecase{xxxRepo: xxxRepo}
}

func (uc *XxxUsecase) Get(ctx context.Context, id string) (*entity.Xxx, error) {
    xxx, err := uc.xxxRepo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("find xxx: %w", err)
    }
    return xxx, nil
}
```

## 4. ハンドラ追加

```go
// internal/adapter/handler/xxx.go
package handler

import (
    "context"

    "connectrpc.com/connect"

    pb "github.com/trendbird/backend/gen/trendbird/v1"
    "github.com/trendbird/backend/internal/adapter/converter"
    "github.com/trendbird/backend/internal/adapter/middleware"
    "github.com/trendbird/backend/internal/adapter/presenter"
    "github.com/trendbird/backend/internal/usecase"
)

type XxxHandler struct {
    uc *usecase.XxxUsecase
}

func NewXxxHandler(uc *usecase.XxxUsecase) *XxxHandler {
    return &XxxHandler{uc: uc}
}

func (h *XxxHandler) GetXxx(
    ctx context.Context,
    req *connect.Request[pb.GetXxxRequest],
) (*connect.Response[pb.GetXxxResponse], error) {
    userID := middleware.UserIDFromContext(ctx)

    result, err := h.uc.Get(ctx, req.Msg.GetId())
    if err != nil {
        return nil, presenter.ToConnectError(err)
    }

    return connect.NewResponse(&pb.GetXxxResponse{
        Xxx: converter.XxxToProto(result),
    }), nil
}
```

## 5. GORMモデル + mapper + リポジトリ実装追加

```go
// internal/infrastructure/persistence/model/xxx.go
package model

import "time"

type Xxx struct {
    ID        string    `gorm:"type:varchar(36);primaryKey"`
    UserID    string    `gorm:"type:varchar(36);index;not null"`
    // ... GORMタグ付きフィールド
    CreatedAt time.Time `gorm:"autoCreateTime"`
    UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
```

```go
// internal/infrastructure/persistence/mapper/xxx.go
package mapper

import (
    "github.com/trendbird/backend/internal/domain/entity"
    "github.com/trendbird/backend/internal/infrastructure/persistence/model"
)

func XxxToEntity(m *model.Xxx) *entity.Xxx {
    return &entity.Xxx{
        ID:        m.ID,
        UserID:    m.UserID,
        CreatedAt: m.CreatedAt,
        UpdatedAt: m.UpdatedAt,
    }
}

func XxxToModel(e *entity.Xxx) *model.Xxx {
    return &model.Xxx{
        ID:        e.ID,
        UserID:    e.UserID,
        CreatedAt: e.CreatedAt,
        UpdatedAt: e.UpdatedAt,
    }
}
```

```go
// internal/infrastructure/persistence/repository/xxx.go
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

var _ domainrepo.XxxRepository = (*XxxRepository)(nil)

type XxxRepository struct {
    db *gorm.DB
}

func NewXxxRepository(db *gorm.DB) *XxxRepository {
    return &XxxRepository{db: db}
}

func (r *XxxRepository) FindByID(ctx context.Context, id string) (*entity.Xxx, error) {
    var m model.Xxx
    if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, apperror.NotFound("xxx not found")
        }
        return nil, apperror.Internal("find xxx", err)
    }
    return mapper.XxxToEntity(&m), nil
}

func (r *XxxRepository) Create(ctx context.Context, xxx *entity.Xxx) error {
    m := mapper.XxxToModel(xxx)
    if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
        return apperror.Internal("create xxx", err)
    }
    xxx.ID = m.ID
    xxx.CreatedAt = m.CreatedAt
    xxx.UpdatedAt = m.UpdatedAt
    return nil
}
```

## 6. DI Container に登録

```go
// di/wire.go の NewContainer 内に追加:
xxxRepo := persrepo.NewXxxRepository(db)
xxxUC := usecase.NewXxxUsecase(xxxRepo)
// Container struct に追加:
XxxHandler: handler.NewXxxHandler(xxxUC),
```

## 7. E2E テスト追加

新機能の handler + repository を実装したら、E2E テストで全パスを検証する。
詳細は `/backend-e2e-test` スキルを参照。
