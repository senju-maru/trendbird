package middleware

import (
	"context"

	"github.com/trendbird/backend/internal/domain/apperror"
)

type contextKey string

const userIDKey contextKey = "userID"
const traceIDKey contextKey = "traceID"

// SetUserID は context にユーザーIDを格納する。
func SetUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserID は context からユーザーIDを取得する。未認証の場合は Unauthenticated エラーを返す。
func GetUserID(ctx context.Context) (string, error) {
	v, ok := ctx.Value(userIDKey).(string)
	if !ok || v == "" {
		return "", apperror.Unauthenticated("authentication required")
	}
	return v, nil
}

// SetTraceID は context にトレースIDを格納する。
func SetTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID は context からトレースIDを取得する。未設定の場合は空文字を返す。
func GetTraceID(ctx context.Context) string {
	v, _ := ctx.Value(traceIDKey).(string)
	return v
}
