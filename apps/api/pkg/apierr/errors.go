package apierr

import (
	"errors"
	"net/http"
)

type AppError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e *AppError) Error() string { return e.Message }

var ErrBadLogin = &AppError{Code: http.StatusUnauthorized, Message: "invalid email or password"}

func Validation(field, message string) *AppError {
	return &AppError{Code: http.StatusUnprocessableEntity, Field: field, Message: message}
}

func NotFound(resource string) *AppError {
	return &AppError{Code: http.StatusNotFound, Message: resource + " not found"}
}

func Forbidden(reason string) *AppError {
	return &AppError{Code: http.StatusForbidden, Message: reason}
}

func IsNotFound(err error) bool {
	var e *AppError
	return errors.As(err, &e) && e.Code == http.StatusNotFound
}
