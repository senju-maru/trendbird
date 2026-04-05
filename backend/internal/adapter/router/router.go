package router

import (
	"context"
	"net/http"
	"time"

	"connectrpc.com/connect"

	"github.com/trendbird/backend/gen/trendbird/v1/trendbirdv1connect"
	"github.com/trendbird/backend/internal/adapter/middleware"
	"github.com/trendbird/backend/internal/di"
)

// New は Connect RPC サービスと Webhook エンドポイントを登録した http.Handler を返す。
func New(c *di.Container) http.Handler {
	interceptors := connect.WithInterceptors(
		c.RecoveryInterceptor,
		c.TraceIDInterceptor,
		c.AuthInterceptor,    // Auth を先に（userID を context に格納）
		c.LoggingInterceptor, // Logging が後（userID を読み取れる）
	)

	mux := http.NewServeMux()

	// ヘルスチェック (JWT 認証外、コンテナオーケストレーションのプローブ用)
	mux.Handle("GET /health", middleware.TraceIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sqlDB, err := c.DB.DB()
		if err != nil {
			http.Error(w, "db unavailable", http.StatusServiceUnavailable)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(ctx); err != nil {
			http.Error(w, "db ping failed", http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})))

	// Connect RPC services
	services := []route{
		wrap(trendbirdv1connect.NewAnalyticsServiceHandler(c.AnalyticsHandler, interceptors)),
		wrap(trendbirdv1connect.NewAuthServiceHandler(c.AuthHandler, interceptors)),
		wrap(trendbirdv1connect.NewAutoDMServiceHandler(c.AutoDMHandler, interceptors)),
		wrap(trendbirdv1connect.NewAutoReplyServiceHandler(c.AutoReplyHandler, interceptors)),
		wrap(trendbirdv1connect.NewDashboardServiceHandler(c.DashboardHandler, interceptors)),
		wrap(trendbirdv1connect.NewNotificationServiceHandler(c.NotificationHandler, interceptors)),
		wrap(trendbirdv1connect.NewPostServiceHandler(c.PostHandler, interceptors)),
		wrap(trendbirdv1connect.NewSettingsServiceHandler(c.SettingsHandler, interceptors)),
		wrap(trendbirdv1connect.NewTopicServiceHandler(c.TopicHandler, interceptors)),
		wrap(trendbirdv1connect.NewTwitterServiceHandler(c.TwitterHandler, interceptors)),
	}
	for _, s := range services {
		mux.Handle(s.path, s.handler)
	}

	// OAuth HTTP endpoint (JWT 認証外、ブラウザリダイレクト)
	mux.Handle("GET /auth/x", middleware.TraceIDMiddleware(c.AuthHTTPHandler))

	return mux
}

type route struct {
	path    string
	handler http.Handler
}

func wrap(path string, h http.Handler) route {
	return route{path: path, handler: h}
}
