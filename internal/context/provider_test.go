package context

import (
	"io"
	"os"
	"testing"

	"github.com/PlakarKorp/kloset/logger"
	"github.com/PlakarKorp/plakar/cookies"
)

// Mock implementations for testing
type mockLogger struct{}

func (m *mockLogger) Trace(s string, args ...interface{})   {}
func (m *mockLogger) Debug(s string, args ...interface{})   {}
func (m *mockLogger) Info(s string, args ...interface{})    {}
func (m *mockLogger) Warn(s string, args ...interface{})    {}
func (m *mockLogger) Error(s string, args ...interface{})   {}
func (m *mockLogger) Fatal(s string, args ...interface{})   { os.Exit(1) }
func (m *mockLogger) SetLevel(level logger.LogLevel)         {}
func (m *mockLogger) SetDebugSource(source string)          {}

type mockCookiesManager struct{}

func (m *mockCookiesManager) Close() error                                      { return nil }
func (m *mockCookiesManager) GetDir() string                                    { return "/tmp" }
func (m *mockCookiesManager) GetAuthToken() (string, error)                     { return "token", nil }
func (m *mockCookiesManager) HasAuthToken() bool                                { return true }
func (m *mockCookiesManager) DeleteAuthToken() error                            { return nil }
func (m *mockCookiesManager) PutAuthToken(token string) error                   { return nil }
func (m *mockCookiesManager) HasRepositoryCookie(id interface{}, name string) bool { return false }
func (m *mockCookiesManager) PutRepositoryCookie(id interface{}, name string) error { return nil }
func (m *mockCookiesManager) IsFirstRun() bool                                  { return false }
func (m *mockCookiesManager) SetFirstRun() error                                { return nil }
func (m *mockCookiesManager) IsDisabledSecurityCheck() bool                     { return false }
func (m *mockCookiesManager) SetDisabledSecurityCheck() error                   { return nil }
func (m *mockCookiesManager) RemoveDisabledSecurityCheck() error                { return nil }

func TestNewDefaultProvider(t *testing.T) {
	logger := &mockLogger{}
	cookies := &mockCookiesManager{}

	provider := NewDefaultProvider(logger, cookies, io.Discard, io.Discard)

	if provider == nil {
		t.Fatal("NewDefaultProvider() returned nil")
	}

	// Test that contexts are accessible
	execCtx := provider.Execution()
	if execCtx == nil {
		t.Error("Execution() returned nil")
	}

	secCtx := provider.Security()
	if secCtx == nil {
		t.Error("Security() returned nil")
	}

	sysCtx := provider.System()
	if sysCtx == nil {
		t.Error("System() returned nil")
	}
}

func TestExecutionContext(t *testing.T) {
	logger := &mockLogger{}
	cookies := &mockCookiesManager{}

	provider := NewDefaultProvider(logger, cookies, io.Discard, io.Discard)
	execCtx := provider.Execution()

	// Test logger
	if execCtx.GetLogger() == nil {
		t.Error("GetLogger() returned nil")
	}

	// Test cookies
	if execCtx.GetCookies() == nil {
		t.Error("GetCookies() returned nil")
	}

	// Test verbose flag
	execCtx.SetVerbose(true)
	if !execCtx.IsVerbose() {
		t.Error("SetVerbose(true) didn't work")
	}

	execCtx.SetVerbose(false)
	if execCtx.IsVerbose() {
		t.Error("SetVerbose(false) didn't work")
	}

	// Test debug flag
	execCtx.SetDebug(true)
	if !execCtx.IsDebug() {
		t.Error("SetDebug(true) didn't work")
	}

	execCtx.SetDebug(false)
	if execCtx.IsDebug() {
		t.Error("SetDebug(false) didn't work")
	}
}

func TestSecurityContext(t *testing.T) {
	logger := &mockLogger{}
	cookies := &mockCookiesManager{}

	provider := NewDefaultProvider(logger, cookies, io.Discard, io.Discard)
	secCtx := provider.Security()

	// Test setting and getting secret
	testSecret := []byte("test-secret-123")
	secCtx.SetSecret(testSecret)

	retrievedSecret := secCtx.GetSecret()
	if string(retrievedSecret) != string(testSecret) {
		t.Error("SetSecret/GetSecret didn't work correctly")
	}

	// Verify that returned secret is a copy, not the original
	retrievedSecret[0] = 'X'
	secondRetrieve := secCtx.GetSecret()
	if secondRetrieve[0] == 'X' {
		t.Error("GetSecret() should return a copy, not the original")
	}

	// Test clearing secret
	secCtx.ClearSecret()
	clearedSecret := secCtx.GetSecret()
	if len(clearedSecret) != 0 {
		t.Error("ClearSecret() didn't clear the secret")
	}

	// Test HasSecret
	if secCtx.HasSecret() {
		t.Error("HasSecret() should return false after clearing")
	}

	secCtx.SetSecret([]byte("new-secret"))
	if !secCtx.HasSecret() {
		t.Error("HasSecret() should return true after setting secret")
	}
}

func TestSystemContext(t *testing.T) {
	logger := &mockLogger{}
	cookies := &mockCookiesManager{}

	provider := NewDefaultProvider(logger, cookies, io.Discard, io.Discard)
	sysCtx := provider.System()

	// Test stdout
	if sysCtx.Stdout() == nil {
		t.Error("Stdout() returned nil")
	}

	// Test stderr
	if sysCtx.Stderr() == nil {
		t.Error("Stderr() returned nil")
	}

	// Test environment variables
	testKey := "TEST_VAR"
	testValue := "test_value"

	sysCtx.SetEnv(testKey, testValue)

	value, exists := sysCtx.GetEnv(testKey)
	if !exists {
		t.Error("GetEnv() should return true for existing variable")
	}
	if value != testValue {
		t.Errorf("Expected value '%s', got '%s'", testValue, value)
	}

	// Test non-existent variable
	_, exists = sysCtx.GetEnv("NON_EXISTENT_VAR")
	if exists {
		t.Error("GetEnv() should return false for non-existent variable")
	}

	// Test ListEnv
	sysCtx.SetEnv("VAR1", "value1")
	sysCtx.SetEnv("VAR2", "value2")

	envList := sysCtx.ListEnv()
	if len(envList) < 2 {
		t.Error("ListEnv() should return at least 2 environment variables")
	}
}

func TestProviderIsolation(t *testing.T) {
	// Create two providers
	logger1 := &mockLogger{}
	cookies1 := &mockCookiesManager{}
	provider1 := NewDefaultProvider(logger1, cookies1, io.Discard, io.Discard)

	logger2 := &mockLogger{}
	cookies2 := &mockCookiesManager{}
	provider2 := NewDefaultProvider(logger2, cookies2, io.Discard, io.Discard)

	// Set different values in each provider
	provider1.Security().SetSecret([]byte("secret1"))
	provider2.Security().SetSecret([]byte("secret2"))

	// Verify isolation
	secret1 := provider1.Security().GetSecret()
	secret2 := provider2.Security().GetSecret()

	if string(secret1) == string(secret2) {
		t.Error("Providers should be isolated from each other")
	}
}

func TestContextThreadSafety(t *testing.T) {
	logger := &mockLogger{}
	cookies := &mockCookiesManager{}
	provider := NewDefaultProvider(logger, cookies, io.Discard, io.Discard)

	secCtx := provider.Security()

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			secret := []byte("secret" + string(rune(id)))
			secCtx.SetSecret(secret)
			_ = secCtx.GetSecret()
			secCtx.ClearSecret()
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we get here without panicking, thread safety is working
}