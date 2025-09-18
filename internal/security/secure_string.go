package security

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"runtime"
	"sync"
)

// SecureString provides a secure wrapper for sensitive string data
// It implements protection against memory dumps and ensures cleanup
type SecureString struct {
	mu       sync.RWMutex
	data     []byte
	locked   bool
	cleared  bool
}

// NewSecureString creates a new SecureString from a byte slice
// The input data is copied and the original should be cleared by the caller
func NewSecureString(data []byte) (*SecureString, error) {
	if len(data) == 0 {
		return nil, errors.New("cannot create SecureString with empty data")
	}

	s := &SecureString{
		data: make([]byte, len(data)),
	}

	// Copy data using constant-time operation to prevent timing attacks
	copy(s.data, data)

	// Try to lock memory (platform-specific)
	if err := s.lock(); err != nil {
		// Log warning but don't fail - locking is best-effort
		// In production, you might want to handle this differently
		_ = err
	}

	// Set finalizer to ensure cleanup even if Clear() is not called
	runtime.SetFinalizer(s, (*SecureString).Clear)

	return s, nil
}

// NewSecureStringFromString creates a SecureString from a regular string
func NewSecureStringFromString(str string) (*SecureString, error) {
	if str == "" {
		return nil, errors.New("cannot create SecureString with empty string")
	}

	// Convert string to bytes
	data := []byte(str)

	// Create SecureString
	s, err := NewSecureString(data)

	// Clear the temporary byte slice
	for i := range data {
		data[i] = 0
	}

	return s, err
}

// Get returns a copy of the secure data
// The caller is responsible for clearing the returned data
func (s *SecureString) Get() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.cleared {
		return nil, errors.New("SecureString has been cleared")
	}

	// Return a copy of the data
	result := make([]byte, len(s.data))
	copy(result, s.data)

	return result, nil
}

// GetString returns the secure data as a string
// The caller should be careful with the returned string as it's immutable
func (s *SecureString) GetString() (string, error) {
	data, err := s.Get()
	if err != nil {
		return "", err
	}

	// Convert to string
	str := string(data)

	// Clear the temporary byte slice
	for i := range data {
		data[i] = 0
	}

	return str, nil
}

// Equals performs constant-time comparison with another SecureString
func (s *SecureString) Equals(other *SecureString) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.cleared {
		return false, errors.New("SecureString has been cleared")
	}

	other.mu.RLock()
	defer other.mu.RUnlock()

	if other.cleared {
		return false, errors.New("other SecureString has been cleared")
	}

	if len(s.data) != len(other.data) {
		return false, nil
	}

	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(s.data, other.data) == 1, nil
}

// EqualsBytes performs constant-time comparison with a byte slice
func (s *SecureString) EqualsBytes(data []byte) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.cleared {
		return false, errors.New("SecureString has been cleared")
	}

	if len(s.data) != len(data) {
		return false, nil
	}

	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(s.data, data) == 1, nil
}

// Clear securely overwrites and clears the secure data
func (s *SecureString) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cleared {
		return
	}

	// Overwrite with random data first
	rand.Read(s.data)

	// Then overwrite with zeros
	for i := range s.data {
		s.data[i] = 0
	}

	// Unlock memory if it was locked
	if s.locked {
		s.unlock()
		s.locked = false
	}

	// Mark as cleared
	s.cleared = true

	// Remove finalizer since we've already cleaned up
	runtime.SetFinalizer(s, nil)
}

// Length returns the length of the secure data
func (s *SecureString) Length() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.cleared {
		return 0
	}
	return len(s.data)
}

// IsCleared returns whether the SecureString has been cleared
func (s *SecureString) IsCleared() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cleared
}