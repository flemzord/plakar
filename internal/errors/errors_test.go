package errors

import (
	"errors"
	"strings"
	"testing"
)

func TestNewAppError(t *testing.T) {
	err := NewAppError(ErrInvalidConfig, "test error message")

	if err == nil {
		t.Fatal("NewAppError() returned nil")
	}

	if err.Code != ErrInvalidConfig {
		t.Errorf("Expected error code %d, got %d", ErrInvalidConfig, err.Code)
	}

	if err.Message != "test error message" {
		t.Errorf("Expected message 'test error message', got '%s'", err.Message)
	}

	if err.StackTrace == "" {
		t.Error("StackTrace not captured")
	}

	if err.Context == nil {
		t.Error("Context map not initialized")
	}
}

func TestNewAppErrorf(t *testing.T) {
	err := NewAppErrorf(ErrInvalidParameter, "invalid parameter: %s", "test")

	if err == nil {
		t.Fatal("NewAppErrorf() returned nil")
	}

	expectedMsg := "invalid parameter: test"
	if err.Message != expectedMsg {
		t.Errorf("Expected message '%s', got '%s'", expectedMsg, err.Message)
	}
}

func TestAppErrorWithCause(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewAppError(ErrFileSystem, "file operation failed").WithCause(cause)

	if err.Cause != cause {
		t.Error("Cause not set correctly")
	}

	// Test error string includes cause
	errStr := err.Error()
	if !strings.Contains(errStr, "underlying error") {
		t.Errorf("Error string doesn't contain cause: %s", errStr)
	}
}

func TestAppErrorWithContext(t *testing.T) {
	err := NewAppError(ErrNetwork, "connection failed").
		WithContext("host", "example.com").
		WithContext("port", 8080)

	if err.Context["host"] != "example.com" {
		t.Errorf("Context 'host' not set correctly")
	}

	if err.Context["port"] != 8080 {
		t.Errorf("Context 'port' not set correctly")
	}
}

func TestAppErrorUnwrap(t *testing.T) {
	cause := errors.New("root cause")
	err := NewAppError(ErrTimeout, "operation timed out").WithCause(cause)

	unwrapped := err.Unwrap()
	if unwrapped != cause {
		t.Error("Unwrap() didn't return the correct cause")
	}
}

func TestAppErrorIs(t *testing.T) {
	err1 := NewAppError(ErrAuthentication, "auth failed")
	err2 := NewAppError(ErrAuthentication, "different message")
	err3 := NewAppError(ErrAuthorization, "not authorized")

	// Same error code should match
	if !err1.Is(err2) {
		t.Error("Errors with same code should match")
	}

	// Different error codes should not match
	if err1.Is(err3) {
		t.Error("Errors with different codes should not match")
	}

	// Non-AppError should not match
	regularErr := errors.New("regular error")
	if err1.Is(regularErr) {
		t.Error("AppError should not match regular error")
	}
}

func TestIsCode(t *testing.T) {
	err := NewAppError(ErrPluginLoad, "plugin failed to load")

	if !IsCode(err, ErrPluginLoad) {
		t.Error("IsCode() should return true for matching code")
	}

	if IsCode(err, ErrPluginExecute) {
		t.Error("IsCode() should return false for non-matching code")
	}

	regularErr := errors.New("regular error")
	if IsCode(regularErr, ErrPluginLoad) {
		t.Error("IsCode() should return false for non-AppError")
	}
}

func TestGetCode(t *testing.T) {
	err := NewAppError(ErrEncryption, "encryption failed")

	code, ok := GetCode(err)
	if !ok {
		t.Error("GetCode() should return true for AppError")
	}

	if code != ErrEncryption {
		t.Errorf("Expected code %d, got %d", ErrEncryption, code)
	}

	regularErr := errors.New("regular error")
	_, ok = GetCode(regularErr)
	if ok {
		t.Error("GetCode() should return false for non-AppError")
	}
}

func TestErrorHandler(t *testing.T) {
	var loggedError error
	handler := NewErrorHandler(func(err error) {
		loggedError = err
	})

	// Test handling nil error (should not log)
	handler.Handle(nil)
	if loggedError != nil {
		t.Error("Handler should not log nil error")
	}

	// Test handling regular error
	testErr := errors.New("test error")
	handler.Handle(testErr)
	if loggedError != testErr {
		t.Error("Handler should log the error")
	}
}

func TestErrorHandlerWithCode(t *testing.T) {
	var loggedError error
	handler := NewErrorHandler(func(err error) {
		loggedError = err
	})

	cause := errors.New("root cause")
	appErr := handler.HandleWithCode(ErrDiskFull, "disk is full", cause)

	if appErr == nil {
		t.Fatal("HandleWithCode() returned nil")
	}

	if appErr.Code != ErrDiskFull {
		t.Errorf("Expected code %d, got %d", ErrDiskFull, appErr.Code)
	}

	if appErr.Message != "disk is full" {
		t.Errorf("Expected message 'disk is full', got '%s'", appErr.Message)
	}

	if appErr.Cause != cause {
		t.Error("Cause not set correctly")
	}

	if loggedError != appErr {
		t.Error("Error not logged by handler")
	}
}

func TestCaptureStackTrace(t *testing.T) {
	trace := captureStackTrace(1)

	if trace == "" {
		t.Error("captureStackTrace() returned empty string")
	}

	// Stack trace should contain this test function name
	if !strings.Contains(trace, "TestCaptureStackTrace") {
		t.Errorf("Stack trace doesn't contain test function name: %s", trace)
	}
}

func TestErrorCodes(t *testing.T) {
	// Test that error codes are unique within their ranges
	codes := []ErrorCode{
		// Configuration errors
		ErrInvalidConfig,
		ErrConfigNotFound,
		ErrConfigParseFailed,

		// Repository errors
		ErrRepositoryAccess,
		ErrRepositoryNotFound,
		ErrRepositoryLocked,
		ErrRepositoryCorrupted,
		ErrRepositoryVersionMismatch,

		// Security errors
		ErrAuthentication,
		ErrAuthorization,
		ErrEncryption,
		ErrDecryption,
		ErrInvalidPassphrase,

		// Plugin errors
		ErrPluginLoad,
		ErrPluginExecute,
		ErrPluginNotFound,
		ErrPluginVersion,

		// System errors
		ErrFileSystem,
		ErrNetwork,
		ErrTimeout,
		ErrCancelled,
		ErrOutOfMemory,
		ErrDiskFull,

		// Validation errors
		ErrInvalidInput,
		ErrMissingParameter,
		ErrInvalidParameter,
		ErrValidationFailed,
	}

	seen := make(map[ErrorCode]bool)
	for _, code := range codes {
		if seen[code] {
			t.Errorf("Duplicate error code: %d", code)
		}
		seen[code] = true
	}
}