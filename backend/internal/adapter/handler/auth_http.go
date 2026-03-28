package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/trendbird/backend/internal/domain/gateway"
)

// AuthHTTPHandler は OAuth リダイレクト用の HTTP ハンドラ。
// Connect RPC とは別に net/http の ServeMux に登録する。
type AuthHTTPHandler struct {
	twitterGW    gateway.TwitterGateway
	secure       bool
	cookieDomain string
}

// NewAuthHTTPHandler creates a new AuthHTTPHandler.
func NewAuthHTTPHandler(twitterGW gateway.TwitterGateway, secure bool, cookieDomain string) *AuthHTTPHandler {
	return &AuthHTTPHandler{twitterGW: twitterGW, secure: secure, cookieDomain: cookieDomain}
}

// ServeHTTP は X OAuth 認可画面へリダイレクトする。
func (h *AuthHTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	result, err := h.twitterGW.BuildAuthorizationURL(r.Context())
	if err != nil {
		slog.Error("failed to build authorization URL", "error", err)
		http.Error(w, "failed to start oauth", http.StatusInternalServerError)
		return
	}

	sameSite := http.SameSiteLaxMode
	if h.secure {
		sameSite = http.SameSiteNoneMode
	}

	// code_verifier を HttpOnly Cookie に保存
	http.SetCookie(w, &http.Cookie{
		Name:     "tb_cv",
		Value:    result.CodeVerifier,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: sameSite,
		Domain:   h.cookieDomain,
	})

	// state を HttpOnly Cookie に保存
	http.SetCookie(w, &http.Cookie{
		Name:     "tb_state",
		Value:    result.State,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: sameSite,
		Domain:   h.cookieDomain,
	})

	slog.Info("redirecting to X authorization", "url_length", len(result.AuthorizationURL))
	http.Redirect(w, r, result.AuthorizationURL, http.StatusFound)
}

// String returns handler description for logging.
func (h *AuthHTTPHandler) String() string {
	return fmt.Sprintf("AuthHTTPHandler{secure=%v, domain=%q}", h.secure, h.cookieDomain)
}
