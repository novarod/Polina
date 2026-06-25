package apierr

import "net/http"

type AppError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e *AppError) Error() string { return e.Message }

var (
	ErrNotFound     = &AppError{Code: http.StatusNotFound, Message: "resource not found"}
	ErrUnauthorized = &AppError{Code: http.StatusUnauthorized, Message: "unauthorized"}
	ErrForbidden    = &AppError{Code: http.StatusForbidden, Message: "forbidden"}
	ErrBadLogin     = &AppError{Code: http.StatusUnauthorized, Message: "invalid email or password"}
	ErrConflict     = &AppError{Code: http.StatusConflict, Message: "resource already exists"}
)

func Validation(field, message string) *AppError {
	return &AppError{Code: http.StatusUnprocessableEntity, Field: field, Message: message}
}

func NotFound(resource string) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: resource + " not found"}
}

func Forbidden(reason string) *AppError {
	return &AppError{Code: http.StatusForbidden, Message: reason}
}
