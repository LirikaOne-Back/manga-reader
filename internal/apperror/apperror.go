package apperror

import (
	"fmt"
	"net/http"
)

type AppError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Err        error  `json:"-"`
	Details    any    `json:"details,omitempty"`
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

const (
	ErrBadRequest          = "BAD_REQUEST"
	ErrUnauthorized        = "UNAUTHORIZED"
	ErrNotFound            = "NOT_FOUND"
	ErrInternalServerError = "INTERNAL_SERVER_ERROR"
	ErrValidation          = "VALIDATION_ERROR"
	ErrDatabaseError       = "DATABASE_ERROR"
)

func NewBadRequestError(msg string, err error) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		Code:       ErrBadRequest,
		Message:    msg,
		Err:        err,
	}
}

func NewUnauthorizedError(msg string, err error) *AppError {
	return &AppError{
		StatusCode: http.StatusUnauthorized,
		Code:       ErrUnauthorized,
		Message:    msg,
		Err:        err,
	}
}

func NewNotFoundError(msg string, err error) *AppError {
	return &AppError{
		StatusCode: http.StatusNotFound,
		Code:       ErrNotFound,
		Message:    msg,
		Err:        err,
	}
}

func NewInternalServerError(msg string, err error) *AppError {
	return &AppError{
		StatusCode: http.StatusInternalServerError,
		Code:       ErrInternalServerError,
		Message:    msg,
		Err:        err,
	}
}

func NewDatabaseError(msg string, err error) *AppError {
	return &AppError{
		StatusCode: http.StatusInternalServerError,
		Code:       ErrDatabaseError,
		Message:    msg,
		Err:        err,
	}
}

func NewValidationError(msg string, details any) *AppError {
	return &AppError{
		StatusCode: http.StatusBadRequest,
		Code:       ErrValidation,
		Message:    msg,
		Details:    details,
	}
}
