package context

import (
	"sync"
)

// SecurityContext holds security-related information
type SecurityContext struct {
	secret      []byte // Encryption key/secret
	keyFromFile string // Key loaded from file
	mu          sync.RWMutex
}

// NewSecurityContext creates a new security context
func NewSecurityContext() *SecurityContext {
	return &SecurityContext{}
}

// SetSecret sets the encryption secret
func (s *SecurityContext) SetSecret(secret []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear old secret if exists
	if s.secret != nil {
		for i := range s.secret {
			s.secret[i] = 0
		}
	}

	// Copy new secret
	s.secret = make([]byte, len(secret))
	copy(s.secret, secret)
}

// GetSecret returns a copy of the encryption secret
func (s *SecurityContext) GetSecret() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.secret == nil {
		return nil
	}

	// Return a copy to prevent external modification
	secret := make([]byte, len(s.secret))
	copy(secret, s.secret)
	return secret
}

// SetKeyFromFile sets the key loaded from a file
func (s *SecurityContext) SetKeyFromFile(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keyFromFile = key
}

// GetKeyFromFile returns the key loaded from file
func (s *SecurityContext) GetKeyFromFile() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.keyFromFile
}

// Clear clears all sensitive data
func (s *SecurityContext) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear secret
	if s.secret != nil {
		for i := range s.secret {
			s.secret[i] = 0
		}
		s.secret = nil
	}

	// Clear key from file
	s.keyFromFile = ""
}