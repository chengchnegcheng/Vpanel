package errors

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Property 2: Database Error Context
// For any database operation that fails, the returned error SHALL contain
// contextual information including the operation type and relevant entity identifiers.
// **Validates: Requirements 2.4**

func TestDatabaseErrorContext_HasOperation(t *testing.T) {
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("database errors contain operation context", prop.ForAll(
		func(operation string) bool {
			if operation == "" {
				return true
			}

			cause := fmt.Errorf("connection failed")
			err := NewDatabaseError(operation, cause)

			// Check that operation is set
			if err.Operation != operation {
				t.Logf("Operation not set: expected %q, got %q", operation, err.Operation)
				return false
			}

			// Check that error message contains operation
			errMsg := err.Error()
			if !strings.Contains(errMsg, operation) {
				t.Logf("Error message does not contain operation: %s", errMsg)
				return false
			}

			// Check that HasContext returns true
			if !err.HasContext() {
				t.Log("HasContext returned false")
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	properties.TestingRun(t)
}

func TestDatabaseErrorContext_HasEntityInfo(t *testing.T) {
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("database errors with entity contain entity info", prop.ForAll(
		func(operation, entity string, entityID int64) bool {
			if operation == "" || entity == "" {
				return true
			}

			cause := fmt.Errorf("record not found")
			err := NewDatabaseErrorWithEntity(operation, entity, entityID, cause)

			// Check that entity info is set
			if err.Entity != entity {
				t.Logf("Entity not set: expected %q, got %q", entity, err.Entity)
				return false
			}

			if err.EntityID != entityID {
				t.Logf("EntityID not set: expected %v, got %v", entityID, err.EntityID)
				return false
			}

			// Check that error message contains entity info
			errMsg := err.Error()
			if !strings.Contains(errMsg, entity) {
				t.Logf("Error message does not contain entity: %s", errMsg)
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.OneConstOf("user", "proxy", "traffic", "protocol"),
		gen.Int64Range(1, 10000),
	))

	properties.TestingRun(t)
}

func TestDatabaseErrorContext_PreservesCause(t *testing.T) {
	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	properties.Property("database errors preserve the underlying cause", prop.ForAll(
		func(operation, causeMsg string) bool {
			if operation == "" || causeMsg == "" {
				return true
			}

			cause := fmt.Errorf(causeMsg)
			err := NewDatabaseError(operation, cause)

			// Check that cause is preserved
			if err.Cause == nil {
				t.Log("Cause is nil")
				return false
			}

			if err.Cause.Error() != causeMsg {
				t.Logf("Cause message mismatch: expected %q, got %q", causeMsg, err.Cause.Error())
				return false
			}

			// Check that Unwrap returns the cause
			unwrapped := err.Unwrap()
			if unwrapped == nil || unwrapped.Error() != causeMsg {
				t.Log("Unwrap did not return the cause")
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	properties.TestingRun(t)
}

func TestAppError_HTTPStatus(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected int
	}{
		{ErrCodeValidation, 400},
		{ErrCodeBadRequest, 400},
		{ErrCodeNotFound, 404},
		{ErrCodeUnauthorized, 401},
		{ErrCodeForbidden, 403},
		{ErrCodeConflict, 409},
		{ErrCodeDatabase, 500},
		{ErrCodeInternal, 500},
		{ErrCodeConfig, 500},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			err := New(tt.code, "test")
			if status := err.HTTPStatus(); status != tt.expected {
				t.Errorf("Expected status %d, got %d", tt.expected, status)
			}
		})
	}
}

func TestAppError_WithContext(t *testing.T) {
	err := New(ErrCodeInternal, "test error")
	err.WithContext("key1", "value1").WithContext("key2", 123)

	if err.Context["key1"] != "value1" {
		t.Errorf("Expected context key1=value1, got %v", err.Context["key1"])
	}
	if err.Context["key2"] != 123 {
		t.Errorf("Expected context key2=123, got %v", err.Context["key2"])
	}
}

func TestAppError_Is(t *testing.T) {
	err1 := New(ErrCodeNotFound, "user not found")
	err2 := New(ErrCodeNotFound, "proxy not found")
	err3 := New(ErrCodeValidation, "invalid input")

	if !errors.Is(err1, err2) {
		t.Error("Expected err1 to match err2 (same code)")
	}
	if errors.Is(err1, err3) {
		t.Error("Expected err1 to not match err3 (different code)")
	}
}

func TestErrorHelpers(t *testing.T) {
	notFoundErr := NewNotFoundError("user", 123)
	if !IsNotFound(notFoundErr) {
		t.Error("Expected IsNotFound to return true")
	}

	validationErr := NewValidationError("invalid email", nil)
	if !IsValidation(validationErr) {
		t.Error("Expected IsValidation to return true")
	}

	unauthorizedErr := NewUnauthorizedError("invalid token")
	if !IsUnauthorized(unauthorizedErr) {
		t.Error("Expected IsUnauthorized to return true")
	}

	dbErr := NewDatabaseError("insert", fmt.Errorf("constraint violation"))
	if !IsDatabase(dbErr) {
		t.Error("Expected IsDatabase to return true")
	}
}

func TestWrap(t *testing.T) {
	cause := fmt.Errorf("original error")
	wrapped := Wrap(cause, ErrCodeInternal, "wrapped message")

	if wrapped.Cause != cause {
		t.Error("Expected cause to be preserved")
	}
	if wrapped.Code != ErrCodeInternal {
		t.Error("Expected code to be ErrCodeInternal")
	}
	if wrapped.Message != "wrapped message" {
		t.Error("Expected message to be 'wrapped message'")
	}
}
