package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"

	"connectrpc.com/connect"
)

// NewRecoveryInterceptor は panic をキャッチしてクライアントに Internal エラーを返すインターセプタ。
// TraceIDInterceptor の外側で動くため、traceID はリクエストヘッダーから直接取得する。
func NewRecoveryInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (resp connect.AnyResponse, retErr error) {
			defer func() {
				if r := recover(); r != nil {
					slog.ErrorContext(ctx, "panic recovered",
						"trace_id", req.Header().Get(traceIDHeader),
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
