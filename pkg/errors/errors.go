// Package errors provides structured error types for the V Panel application.
// It supports error codes, contextual information, and error wrapping.
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode represents an error code type.
type ErrorCode string

// Error codes
const (
	ErrCodeValidation   ErrorCode = "VALIDATION_ERROR"
	ErrCodeNotFound     ErrorCode = "NOT_FOUND"
	ErrCodeUnauthorized ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden    ErrorCode = "FORBIDDEN"
	ErrCodeInternal     ErrorCode = "INTERNAL_ERROR"
	ErrCodeDatabase     ErrorCode = "DATABASE_ERROR"
	ErrCodeConfig       ErrorCode = "CONFIG_ERROR"
	ErrCodeConflict     ErrorCode = "CONFLICT"
	ErrCodeBadRequest   ErrorCode = "BAD_REQUEST"
)

// AppError represents an application error with code, message, and context.
type AppError struct {
	Code      ErrorCode      `json:"code"`
	Message   string         `json:"message"`
	Details   any            `json:"details,omitempty"`
	Context   map[string]any `json:"context,omitempty"`
	Cause     error          `json:"-"`
	Operation string         `json:"-"` // The operation that failed
	Entity    string         `json:"-"` // The entity involved (e.g., "user", "proxy")
	EntityID  any            `json:"-"` // The entity ID if applicable
}

// Error implements the error interface.
func (e *AppError) Error() string {
	msg := fmt.Sprintf("%s: %s", e.Code, e.Message)
	if e.Operation != "" {
		msg = fmt.Sprintf("%s (operation: %s)", msg, e.Operation)
	}
	if e.Entity != "" {
		msg = fmt.Sprintf("%s (entity: %s", msg, e.Entity)
		if e.EntityID != nil {
			msg = fmt.Sprintf("%s, id: %v", msg, e.EntityID)
		}
		msg += ")"
	}
	if e.Cause != nil {
		msg = fmt.Sprintf("%s: %v", msg, e.Cause)
	}
	return msg
}

// Unwrap returns the underlying error.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target.
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// HTTPStatus returns the appropriate HTTP status code for the error.
func (e *AppError) HTTPStatus() int {
	switch e.Code {
	case ErrCodeValidation, ErrCodeBadRequest:
		return http.StatusBadRequest
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeForbidden:
		return http.StatusForbidden
	case ErrCodeConflict:
		return http.StatusConflict
	case ErrCodeDatabase, ErrCodeInternal, ErrCodeConfig:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// WithContext adds context to the error.
func (e *AppError) WithContext(key string, value any) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// WithOperation sets the operation that failed.
func (e *AppError) WithOperation(op string) *AppError {
	e.Operation = op
	return e
}

// WithEntity sets the entity information.
func (e *AppError) WithEntity(entity string, id any) *AppError {
	e.Entity = entity
	e.EntityID = id
	return e
}

// HasContext checks if the error has context information.
func (e *AppError) HasContext() bool {
	return e.Operation != "" || e.Entity != "" || len(e.Context) > 0
}

// New creates a new AppError.
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an error with an AppError.
func Wrap(err error, code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   err,
	}
}

// NewValidationError creates a validation error.
func NewValidationError(message string, details any) *AppError {
	return &AppError{
		Code:    ErrCodeValidation,
		Message: message,
		Details: details,
	}
}

// NewNotFoundError creates a not found error.
func NewNotFoundError(entity string, id any) *AppError {
	return &AppError{
		Code:     ErrCodeNotFound,
		Message:  fmt.Sprintf("%s not found", entity),
		Entity:   entity,
		EntityID: id,
	}
}

// NewUnauthorizedError creates an unauthorized error.
func NewUnauthorizedError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeUnauthorized,
		Message: message,
	}
}

// NewForbiddenError creates a forbidden error.
func NewForbiddenError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeForbidden,
		Message: message,
	}
}

// NewInternalError creates an internal error.
func NewInternalError(message string, cause error) *AppError {
	return &AppError{
		Code:    ErrCodeInternal,
		Message: message,
		Cause:   cause,
	}
}

// NewDatabaseError creates a database error with operation context.
func NewDatabaseError(operation string, cause error) *AppError {
	return &AppError{
		Code:      ErrCodeDatabase,
		Message:   fmt.Sprintf("database operation failed: %s", operation),
		Operation: operation,
		Cause:     cause,
	}
}

// NewDatabaseErrorWithEntity creates a database error with entity context.
func NewDatabaseErrorWithEntity(operation string, entity string, entityID any, cause error) *AppError {
	return &AppError{
		Code:      ErrCodeDatabase,
		Message:   fmt.Sprintf("database operation failed: %s", operation),
		Operation: operation,
		Entity:    entity,
		EntityID:  entityID,
		Cause:     cause,
	}
}

// NewConfigError creates a configuration error.
func NewConfigError(message string, cause error) *AppError {
	return &AppError{
		Code:    ErrCodeConfig,
		Message: message,
		Cause:   cause,
	}
}

// NewConflictError creates a conflict error.
func NewConflictError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeConflict,
		Message: message,
	}
}

// NewBadRequestError creates a bad request error.
func NewBadRequestError(message string) *AppError {
	return &AppError{
		Code:    ErrCodeBadRequest,
		Message: message,
	}
}

// IsAppError checks if an error is an AppError.
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

// AsAppError converts an error to an AppError if possible.
func AsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// GetCode returns the error code if the error is an AppError.
func GetCode(err error) ErrorCode {
	if appErr, ok := AsAppError(err); ok {
		return appErr.Code
	}
	return ErrCodeInternal
}

// IsNotFound checks if the error is a not found error.
func IsNotFound(err error) bool {
	return GetCode(err) == ErrCodeNotFound
}

// IsValidation checks if the error is a validation error.
func IsValidation(err error) bool {
	return GetCode(err) == ErrCodeValidation
}

// IsUnauthorized checks if the error is an unauthorized error.
func IsUnauthorized(err error) bool {
	return GetCode(err) == ErrCodeUnauthorized
}

// IsDatabase checks if the error is a database error.
func IsDatabase(err error) bool {
	return GetCode(err) == ErrCodeDatabase
}
