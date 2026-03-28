package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"connectrpc.com/connect"

	"github.com/trendbird/backend/internal/infrastructure/auth"
)

// publicProcedures は認証をスキップする公開エンドポイント。
var publicProcedures = map[string]bool{
	"/trendbird.v1.AuthService/XAuth": true,
}

// NewAuthInterceptor は JWT 認証インターセプタを返す。
func NewAuthInterceptor(jwt *auth.JWTService) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if publicProcedures[req.Spec().Procedure] {
				return next(ctx, req)
			}

			token := extractBearerToken(req.Header())
			if token == "" {
				token = extractCookieToken(req.Header())
			}
			if token == "" {
				slog.WarnContext(ctx, "auth: token not found",
					"trace_id", GetTraceID(ctx),
					"procedure", req.Spec().Procedure,
				)
				return nil, connect.NewError(connect.CodeUnauthenticated, nil)
			}

			userID, err := jwt.ValidateToken(token)
			if err != nil {
				slog.WarnContext(ctx, "auth: invalid token",
					"trace_id", GetTraceID(ctx),
					"procedure", req.Spec().Procedure,
					"error", err.Error(),
				)
				return nil, connect.NewError(connect.CodeUnauthenticated, err)
			}

			ctx = SetUserID(ctx, userID)
			return next(ctx, req)
		}
	}
}

func extractBearerToken(h http.Header) string {
	v := h.Get("Authorization")
	if after, ok := strings.CutPrefix(v, "Bearer "); ok {
		return after
	}
	return ""
}

func extractCookieToken(h http.Header) string {
	r := &http.Request{Header: h}
	c, err := r.Cookie("tb_jwt")
	if err != nil {
		return ""
	}
	return c.Value
}
