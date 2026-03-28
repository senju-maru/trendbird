package middleware

import (
	"context"
	"log/slog"
	"time"

	"connectrpc.com/connect"
)

// NewLoggingInterceptor はリクエストの procedure 名と所要時間をログに記録するインターセプタ。
func NewLoggingInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			start := time.Now()
			resp, err := next(ctx, req)
			duration := time.Since(start)

			traceID := GetTraceID(ctx)
			userID, _ := GetUserID(ctx)

			if err != nil {
				attrs := []any{
					"trace_id", traceID,
					"user_id", userID,
					"procedure", req.Spec().Procedure,
					"duration_ms", duration.Milliseconds(),
					"code", connect.CodeOf(err).String(),
					"error", err.Error(),
				}
				switch connect.CodeOf(err) {
				case connect.CodeInternal, connect.CodeUnknown, connect.CodeDataLoss, connect.CodeUnavailable:
					slog.ErrorContext(ctx, "rpc error", attrs...)
				default:
					slog.WarnContext(ctx, "rpc error", attrs...)
				}
			} else {
				slog.InfoContext(ctx, "rpc success",
					"trace_id", traceID,
					"user_id", userID,
					"procedure", req.Spec().Procedure,
					"duration_ms", duration.Milliseconds(),
				)
			}

			return resp, err
		}
	}
}
