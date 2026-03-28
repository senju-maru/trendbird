package mcp

import (
	"errors"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/trendbird/backend/internal/domain/apperror"
)

// errorResult は apperror を日本語メッセージの MCP エラー結果に変換する。
func errorResult(err error) *mcp.CallToolResult {
	return mcp.NewToolResultError(translateError(err))
}

func translateError(err error) string {
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		return "内部エラーが発生しました: " + err.Error()
	}

	switch appErr.Code {
	case apperror.CodeNotFound:
		return "見つかりませんでした: " + appErr.Message
	case apperror.CodeInvalidArgument:
		return "入力内容に誤りがあります: " + appErr.Message
	case apperror.CodePermissionDenied:
		return "権限がありません: " + appErr.Message
	case apperror.CodeAlreadyExists:
		return "既に存在します: " + appErr.Message
	default:
		return "エラーが発生しました: " + appErr.Message
	}
}
