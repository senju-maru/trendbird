package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"connectrpc.com/connect"

	"github.com/trendbird/backend/internal/adapter/middleware"
)

// fakeRequest は connect.AnyRequest を満たす最小限のスタブ。
type fakeRequest struct {
	connect.AnyRequest
	header http.Header
	spec   connect.Spec
}

func newFakeRequest() *fakeRequest {
	return &fakeRequest{
		header: make(http.Header),
		spec:   connect.Spec{Procedure: "/test.Service/Method"},
	}
}

func (r *fakeRequest) Header() http.Header { return r.header }
func (r *fakeRequest) Spec() connect.Spec  { return r.spec }

// fakeResponse は connect.AnyResponse を満たす最小限のスタブ。
type fakeResponse struct {
	connect.AnyResponse
	header http.Header
}

func newFakeResponse() *fakeResponse {
	return &fakeResponse{header: make(http.Header)}
}

func (r *fakeResponse) Header() http.Header { return r.header }

func TestTraceIDInterceptor_GeneratesID(t *testing.T) {
	interceptor := middleware.NewTraceIDInterceptor()

	var capturedCtx context.Context
	next := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		capturedCtx = ctx
		return newFakeResponse(), nil
	}

	req := newFakeRequest()
	resp, err := interceptor(next)(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// context にトレースIDが設定されていること
	traceID := middleware.GetTraceID(capturedCtx)
	if traceID == "" {
		t.Fatal("trace ID should be set in context")
	}

	// レスポンスヘッダーにトレースIDが含まれること
	got := resp.Header().Get("X-Trace-Id")
	if got != traceID {
		t.Errorf("response header X-Trace-Id = %q, want %q", got, traceID)
	}
}

func TestTraceIDInterceptor_PropagatesClientID(t *testing.T) {
	interceptor := middleware.NewTraceIDInterceptor()
	clientID := "client-supplied-trace-id"

	var capturedCtx context.Context
	next := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		capturedCtx = ctx
		return newFakeResponse(), nil
	}

	req := newFakeRequest()
	req.Header().Set("X-Trace-Id", clientID)

	resp, err := interceptor(next)(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	traceID := middleware.GetTraceID(capturedCtx)
	if traceID != clientID {
		t.Errorf("context trace ID = %q, want %q", traceID, clientID)
	}

	got := resp.Header().Get("X-Trace-Id")
	if got != clientID {
		t.Errorf("response header X-Trace-Id = %q, want %q", got, clientID)
	}
}

func TestTraceIDInterceptor_RejectsTooLongID(t *testing.T) {
	interceptor := middleware.NewTraceIDInterceptor()
	longID := strings.Repeat("a", 200)

	var capturedCtx context.Context
	next := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		capturedCtx = ctx
		return newFakeResponse(), nil
	}

	req := newFakeRequest()
	req.Header().Set("X-Trace-Id", longID)

	_, err := interceptor(next)(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	traceID := middleware.GetTraceID(capturedCtx)
	if traceID == longID {
		t.Error("should reject trace ID exceeding max length")
	}
	if traceID == "" {
		t.Error("should generate a new trace ID when client ID is rejected")
	}
}

func TestTraceIDInterceptor_RejectsControlChars(t *testing.T) {
	interceptor := middleware.NewTraceIDInterceptor()
	maliciousID := "trace-id\ninjected-header: value"

	var capturedCtx context.Context
	next := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		capturedCtx = ctx
		return newFakeResponse(), nil
	}

	req := newFakeRequest()
	req.Header().Set("X-Trace-Id", maliciousID)

	_, err := interceptor(next)(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	traceID := middleware.GetTraceID(capturedCtx)
	if traceID == maliciousID {
		t.Error("should reject trace ID containing control characters")
	}
	if traceID == "" {
		t.Error("should generate a new trace ID when client ID is rejected")
	}
}

func TestTraceIDInterceptor_ErrorResponse(t *testing.T) {
	interceptor := middleware.NewTraceIDInterceptor()

	next := func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		return nil, connect.NewError(connect.CodeInternal, errors.New("boom"))
	}

	req := newFakeRequest()
	_, err := interceptor(next)(context.Background(), req)
	if err == nil {
		t.Fatal("expected error")
	}

	connectErr, ok := err.(*connect.Error)
	if !ok {
		t.Fatalf("expected *connect.Error, got %T", err)
	}

	got := connectErr.Meta().Get("X-Trace-Id")
	if got == "" {
		t.Fatal("error metadata should contain X-Trace-Id")
	}
}
