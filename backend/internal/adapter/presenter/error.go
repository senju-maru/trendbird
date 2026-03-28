package presenter

import (
	"errors"

	"connectrpc.com/connect"

	"github.com/trendbird/backend/internal/domain/apperror"
)

// ToConnectError は AppError を Connect エラーに変換する。
func ToConnectError(err error) *connect.Error {
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) {
		return connect.NewError(connect.CodeInternal, err)
	}

	var code connect.Code
	switch appErr.Code {
	case apperror.CodeNotFound:
		code = connect.CodeNotFound
	case apperror.CodePermissionDenied:
		code = connect.CodePermissionDenied
	case apperror.CodeInvalidArgument:
		code = connect.CodeInvalidArgument
	case apperror.CodeResourceExhausted:
		code = connect.CodeResourceExhausted
	case apperror.CodeUnauthenticated:
		code = connect.CodeUnauthenticated
	case apperror.CodeAlreadyExists:
		code = connect.CodeAlreadyExists
	default:
		code = connect.CodeInternal
	}

	return connect.NewError(code, appErr)
}
