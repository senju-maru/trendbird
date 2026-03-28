package apperror

import (
	"errors"
	"fmt"
)

// Code represents an application-level error code.
type Code int

const (
	CodeNotFound         Code = 1
	CodePermissionDenied Code = 2
	CodeInvalidArgument  Code = 3
	CodeResourceExhausted Code = 4
	CodeUnauthenticated  Code = 5
	CodeInternal         Code = 6
	CodeAlreadyExists    Code = 7
)

// AppError is a framework-independent domain error.
type AppError struct {
	Code    Code
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// NotFound creates a not-found error.
func NotFound(msg string) *AppError {
	return &AppError{Code: CodeNotFound, Message: msg}
}

// PermissionDenied creates a permission-denied error.
func PermissionDenied(msg string) *AppError {
	return &AppError{Code: CodePermissionDenied, Message: msg}
}

// InvalidArgument creates an invalid-argument error.
func InvalidArgument(msg string) *AppError {
	return &AppError{Code: CodeInvalidArgument, Message: msg}
}

// ResourceExhausted creates a resource-exhausted error.
func ResourceExhausted(msg string) *AppError {
	return &AppError{Code: CodeResourceExhausted, Message: msg}
}

// Unauthenticated creates an unauthenticated error.
func Unauthenticated(msg string) *AppError {
	return &AppError{Code: CodeUnauthenticated, Message: msg}
}

// Internal creates an internal error.
func Internal(msg string) *AppError {
	return &AppError{Code: CodeInternal, Message: msg}
}

// AlreadyExists creates an already-exists error.
func AlreadyExists(msg string) *AppError {
	return &AppError{Code: CodeAlreadyExists, Message: msg}
}

// Wrap wraps an existing error with an application error code and message.
func Wrap(code Code, msg string, err error) *AppError {
	return &AppError{Code: code, Message: msg, Err: err}
}

// IsCode checks whether any error in the chain matches the given code.
func IsCode(err error, code Code) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}
