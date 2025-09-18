package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// ErrorCode represents different types of application errors
type ErrorCode int

const (
	// Configuration errors
	ErrInvalidConfig ErrorCode = iota + 1000
	ErrConfigNotFound
	ErrConfigParseFailed

	// Repository errors
	ErrRepositoryAccess ErrorCode = iota + 2000
	ErrRepositoryNotFound
	ErrRepositoryLocked
	ErrRepositoryCorrupted
	ErrRepositoryVersionMismatch

	// Security errors
	ErrAuthentication ErrorCode = iota + 3000
	ErrAuthorization
	ErrEncryption
	ErrDecryption
	ErrInvalidPassphrase

	// Plugin errors
	ErrPluginLoad ErrorCode = iota + 4000
	ErrPluginExecute
	ErrPluginNotFound
	ErrPluginVersion

	// System errors
	ErrFileSystem ErrorCode = iota + 5000
	ErrNetwork
	ErrTimeout
	ErrCancelled
	ErrOutOfMemory
	ErrDiskFull

	// Validation errors
	ErrInvalidInput ErrorCode = iota + 6000
	ErrMissingParameter
	ErrInvalidParameter
	ErrValidationFailed
)

// AppError represents an application error with context
type AppError struct {
	Code       ErrorCode
	Message    string
	Cause      error
	StackTrace string
	Context    map[string]interface{}
}

// NewAppError creates a new application error
func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StackTrace: captureStackTrace(2),
		Context:    make(map[string]interface{}),
	}
}

// NewAppErrorf creates a new application error with formatted message
func NewAppErrorf(code ErrorCode, format string, args ...interface{}) *AppError {
	return &AppError{
		Code:       code,
		Message:    fmt.Sprintf(format, args...),
		StackTrace: captureStackTrace(2),
		Context:    make(map[string]interface{}),
	}
}

// WithCause adds a cause error to the AppError
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// WithContext adds context information to the error
func (e *AppError) WithContext(key string, value interface{}) *AppError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// Error implements the error interface
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap returns the cause error
func (e *AppError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target error
func (e *AppError) Is(target error) bool {
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// captureStackTrace captures the current stack trace
func captureStackTrace(skip int) string {
	var builder strings.Builder
	pcs := make([]uintptr, 10)
	n := runtime.Callers(skip, pcs)

	if n == 0 {
		return ""
	}

	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		builder.WriteString(fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))
		if !more {
			break
		}
	}

	return builder.String()
}

// ErrorHandler provides centralized error handling
type ErrorHandler struct {
	logFunc func(error)
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(logFunc func(error)) *ErrorHandler {
	return &ErrorHandler{
		logFunc: logFunc,
	}
}

// Handle processes an error
func (h *ErrorHandler) Handle(err error) {
	if err == nil {
		return
	}

	if h.logFunc != nil {
		h.logFunc(err)
	}
}

// HandleWithCode creates and handles an AppError with the given code
func (h *ErrorHandler) HandleWithCode(code ErrorCode, message string, cause error) *AppError {
	appErr := NewAppError(code, message).WithCause(cause)
	h.Handle(appErr)
	return appErr
}

// IsCode checks if an error has a specific error code
func IsCode(err error, code ErrorCode) bool {
	appErr, ok := err.(*AppError)
	if !ok {
		return false
	}
	return appErr.Code == code
}

// GetCode extracts the error code from an error
func GetCode(err error) (ErrorCode, bool) {
	appErr, ok := err.(*AppError)
	if !ok {
		return 0, false
	}
	return appErr.Code, true
}