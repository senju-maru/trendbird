package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"connectrpc.com/connect"

	trendbirdv1 "github.com/trendbird/backend/gen/trendbird/v1"
	"github.com/trendbird/backend/gen/trendbird/v1/trendbirdv1connect"
	"github.com/trendbird/backend/internal/adapter/converter"
	"github.com/trendbird/backend/internal/adapter/middleware"
	"github.com/trendbird/backend/internal/adapter/presenter"
	"github.com/trendbird/backend/internal/usecase"
)

// AuthHandler implements the AuthService Connect RPC handler.
type AuthHandler struct {
	uc           *usecase.AuthUsecase
	secure       bool
	cookieDomain string
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(uc *usecase.AuthUsecase, secure bool, cookieDomain string) *AuthHandler {
	return &AuthHandler{uc: uc, secure: secure, cookieDomain: cookieDomain}
}

// XAuth exchanges an OAuth code for authentication tokens.
func (h *AuthHandler) XAuth(
	ctx context.Context,
	req *connect.Request[trendbirdv1.XAuthRequest],
) (*connect.Response[trendbirdv1.XAuthResponse], error) {
	codeVerifier := getCookieValue(req.Header(), "tb_cv")
	userAgent := req.Header().Get("User-Agent")

	slog.InfoContext(ctx, "XAuth called",
		"has_code_verifier", codeVerifier != "",
		"user_agent", userAgent,
	)

	result, err := h.uc.XAuth(ctx, req.Msg.GetOauthCode(), codeVerifier)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	resp := connect.NewResponse(&trendbirdv1.XAuthResponse{
		User:            converter.UserToProto(result.User),
		TutorialPending: result.TutorialPending,
	})

	resp.Header().Add("Set-Cookie", buildJWTCookie(result.Token, h.secure, h.cookieDomain))
	resp.Header().Add("Set-Cookie", clearCookie("tb_cv", h.secure, ""))
	resp.Header().Add("Set-Cookie", clearCookie("tb_state", h.secure, ""))

	return resp, nil
}

// Logout clears the user's session.
func (h *AuthHandler) Logout(
	ctx context.Context,
	req *connect.Request[trendbirdv1.LogoutRequest],
) (*connect.Response[trendbirdv1.LogoutResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	if err := h.uc.Logout(ctx, userID); err != nil {
		return nil, presenter.ToConnectError(err)
	}

	resp := connect.NewResponse(&trendbirdv1.LogoutResponse{})
	resp.Header().Add("Set-Cookie", clearCookie("tb_jwt", h.secure, h.cookieDomain))

	return resp, nil
}

// GetCurrentUser returns the authenticated user's info.
func (h *AuthHandler) GetCurrentUser(
	ctx context.Context,
	req *connect.Request[trendbirdv1.GetCurrentUserRequest],
) (*connect.Response[trendbirdv1.GetCurrentUserResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	result, err := h.uc.GetCurrentUser(ctx, userID)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.GetCurrentUserResponse{
		User:            converter.UserToProto(result.User),
		TutorialPending: result.TutorialPending,
	}), nil
}

// DeleteAccount deletes the authenticated user's account.
func (h *AuthHandler) DeleteAccount(
	ctx context.Context,
	req *connect.Request[trendbirdv1.DeleteAccountRequest],
) (*connect.Response[trendbirdv1.DeleteAccountResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	if err := h.uc.DeleteAccount(ctx, userID); err != nil {
		return nil, presenter.ToConnectError(err)
	}

	resp := connect.NewResponse(&trendbirdv1.DeleteAccountResponse{})
	resp.Header().Add("Set-Cookie", clearCookie("tb_jwt", h.secure, h.cookieDomain))

	return resp, nil
}

// CompleteTutorial marks the user's onboarding tutorial as completed.
func (h *AuthHandler) CompleteTutorial(
	ctx context.Context,
	req *connect.Request[trendbirdv1.CompleteTutorialRequest],
) (*connect.Response[trendbirdv1.CompleteTutorialResponse], error) {
	userID, err := middleware.GetUserID(ctx)
	if err != nil {
		return nil, presenter.ToConnectError(err)
	}

	if err := h.uc.CompleteTutorial(ctx, userID); err != nil {
		return nil, presenter.ToConnectError(err)
	}

	return connect.NewResponse(&trendbirdv1.CompleteTutorialResponse{}), nil
}

// buildJWTCookie creates a Set-Cookie header value for the JWT token.
// domain が非空の場合、Domain 属性を付与してクロスサブドメインで共有可能にする。
func buildJWTCookie(token string, secure bool, domain string) string {
	sameSite := "Lax"
	if secure {
		sameSite = "None"
	}
	cookie := fmt.Sprintf("tb_jwt=%s; HttpOnly; SameSite=%s; Path=/; Max-Age=2592000", token, sameSite)
	if secure {
		cookie += "; Secure"
	}
	if domain != "" {
		cookie += "; Domain=" + domain
	}
	return cookie
}

// clearCookie creates a Set-Cookie header value that clears the named cookie.
// domain が非空の場合、Domain 属性を付与する（Cookie の正常なクリアに必要）。
func clearCookie(name string, secure bool, domain string) string {
	sameSite := "Lax"
	if secure {
		sameSite = "None"
	}
	cookie := fmt.Sprintf("%s=; HttpOnly; SameSite=%s; Path=/; Max-Age=0", name, sameSite)
	if secure {
		cookie += "; Secure"
	}
	if domain != "" {
		cookie += "; Domain=" + domain
	}
	return cookie
}

// getCookieValue extracts a cookie value from the request headers.
func getCookieValue(header http.Header, name string) string {
	prefix := name + "="
	for _, cookieHeader := range header.Values("Cookie") {
		for part := range strings.SplitSeq(cookieHeader, ";") {
			part = strings.TrimSpace(part)
			if v, ok := strings.CutPrefix(part, prefix); ok {
				return v
			}
		}
	}
	return ""
}

var _ trendbirdv1connect.AuthServiceHandler = (*AuthHandler)(nil)
