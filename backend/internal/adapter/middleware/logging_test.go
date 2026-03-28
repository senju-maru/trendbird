package middleware

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"connectrpc.com/connect"
)

// captureHandler は slog.Record をキャプチャするテスト用ハンドラ。
type captureHandler struct {
	records []slog.Record
}

func (h *captureHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.records = append(h.records, r)
	return nil
}
func (h *captureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(string) slog.Handler       { return h }

type empty struct{}

func TestLoggingInterceptor_ErrorLevel(t *testing.T) {
	tests := []struct {
		name      string
		code      connect.Code
		wantLevel slog.Level
	}{
		{"permission_denied_is_warn", connect.CodePermissionDenied, slog.LevelWarn},
		{"not_found_is_warn", connect.CodeNotFound, slog.LevelWarn},
		{"invalid_argument_is_warn", connect.CodeInvalidArgument, slog.LevelWarn},
		{"resource_exhausted_is_warn", connect.CodeResourceExhausted, slog.LevelWarn},
		{"already_exists_is_warn", connect.CodeAlreadyExists, slog.LevelWarn},
		{"unauthenticated_is_warn", connect.CodeUnauthenticated, slog.LevelWarn},
		{"internal_is_error", connect.CodeInternal, slog.LevelError},
		{"unknown_is_error", connect.CodeUnknown, slog.LevelError},
		{"data_loss_is_error", connect.CodeDataLoss, slog.LevelError},
		{"unavailable_is_error", connect.CodeUnavailable, slog.LevelError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := &captureHandler{}
			old := slog.Default()
			slog.SetDefault(slog.New(ch))
			defer slog.SetDefault(old)

			interceptor := NewLoggingInterceptor()

			next := connect.UnaryFunc(func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
				return nil, connect.NewError(tt.code, errors.New("test error"))
			})

			handler := interceptor(next)

			req := connect.NewRequest(&empty{})
			_, _ = handler(context.Background(), req)

			var found bool
			for _, r := range ch.records {
				if r.Message == "rpc error" {
					found = true
					if r.Level != tt.wantLevel {
						t.Errorf("log level: want %s, got %s", tt.wantLevel, r.Level)
					}
					break
				}
			}
			if !found {
				t.Error("no 'rpc error' log entry found")
			}
		})
	}

	t.Run("success_is_info", func(t *testing.T) {
		ch := &captureHandler{}
		old := slog.Default()
		slog.SetDefault(slog.New(ch))
		defer slog.SetDefault(old)

		interceptor := NewLoggingInterceptor()

		next := connect.UnaryFunc(func(_ context.Context, _ connect.AnyRequest) (connect.AnyResponse, error) {
			return nil, nil
		})

		handler := interceptor(next)

		req := connect.NewRequest(&empty{})
		_, _ = handler(context.Background(), req)

		var found bool
		for _, r := range ch.records {
			if r.Message == "rpc success" {
				found = true
				if r.Level != slog.LevelInfo {
					t.Errorf("log level: want INFO, got %s", r.Level)
				}
				break
			}
		}
		if !found {
			t.Error("no 'rpc success' log entry found")
		}
	})
}
