package middleware

import (
	"net/http"
)

// TraceIDMiddleware は非 RPC エンドポイント（Webhook, healthz, OAuth）用の
// HTTP ミドルウェア。Connect RPC インターセプターと同じ context キーとバリデーションを使用する。
func TraceIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get(traceIDHeader)
		if !isValidTraceID(traceID) {
			traceID = newTraceID()
		}

		ctx := SetTraceID(r.Context(), traceID)
		w.Header().Set(traceIDHeader, traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
