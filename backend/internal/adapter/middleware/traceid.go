package middleware

import (
	"context"
	"unicode"

	"connectrpc.com/connect"
	"github.com/google/uuid"
)

const traceIDHeader = "X-Trace-Id"

// maxTraceIDLen はクライアント供給 traceID の最大長。
const maxTraceIDLen = 128

// isValidTraceID はクライアント供給の traceID が安全かを検証する。
// 長さ制限と制御文字（改行等）の排除を行い、ログインジェクションを防止する。
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

// newTraceID は UUID v7 を生成して返す。
func newTraceID() string {
	return uuid.Must(uuid.NewV7()).String()
}

// NewTraceIDInterceptor はリクエストにトレースIDを付与するインターセプタ。
// クライアントから X-Trace-Id ヘッダーが送られていればバリデーション後に使用し、
// なければ UUID v7 を自動生成する。
func NewTraceIDInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			traceID := req.Header().Get(traceIDHeader)
			if !isValidTraceID(traceID) {
				traceID = newTraceID()
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
