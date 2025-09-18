package context

import (
	"sync"

	"github.com/PlakarKorp/plakar/internal/security"
)

// SecurityContext holds security-related information using SecureString
type SecurityContext struct {
	secret      *security.SecureString // Encryption key/secret (protected)
	keyFromFile *security.SecureString // Key loaded from file (protected)
	config      *security.SecureConfig  // Secure configuration storage
	mu          sync.RWMutex
}

// NewSecurityContext creates a new security context
func NewSecurityContext() *SecurityContext {
	return &SecurityContext{
		config: security.NewSecureConfig(),
	}
}

// SetSecret sets the encryption secret using SecureString for protection
func (s *SecurityContext) SetSecret(secret []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear old secret if exists
	if s.secret != nil {
		s.secret.Clear()
	}

	// Create new SecureString (ignore error for empty secrets)
	if len(secret) > 0 {
		s.secret, _ = security.NewSecureString(secret)
	} else {
		s.secret = nil
	}
}

// GetSecret returns a copy of the encryption secret
// The caller is responsible for clearing the returned data
func (s *SecurityContext) GetSecret() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.secret == nil || s.secret.IsCleared() {
		return nil
	}

	// Get returns a copy, errors are ignored since we checked IsCleared
	secret, _ := s.secret.Get()
	return secret
}

// SetKeyFromFile sets the key loaded from a file using SecureString
func (s *SecurityContext) SetKeyFromFile(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear old key if exists
	if s.keyFromFile != nil {
		s.keyFromFile.Clear()
	}

	// Create new SecureString (ignore error for empty keys)
	if key != "" {
		s.keyFromFile, _ = security.NewSecureStringFromString(key)
	} else {
		s.keyFromFile = nil
	}
}

// GetKeyFromFile returns the key loaded from file
func (s *SecurityContext) GetKeyFromFile() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.keyFromFile == nil || s.keyFromFile.IsCleared() {
		return ""
	}

	// GetString returns a copy, errors are ignored since we checked IsCleared
	key, _ := s.keyFromFile.GetString()
	return key
}

// Clear securely clears all sensitive data
func (s *SecurityContext) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear secret using SecureString's secure clearing
	if s.secret != nil {
		s.secret.Clear()
		s.secret = nil
	}

	// Clear key from file using SecureString's secure clearing
	if s.keyFromFile != nil {
		s.keyFromFile.Clear()
		s.keyFromFile = nil
	}

	// Clear secure config
	if s.config != nil {
		s.config.Clear()
	}
}

// SetConfigValue stores a secure configuration value
func (s *SecurityContext) SetConfigValue(key string, value []byte) error {
	return s.config.Set(key, value)
}

// GetConfigValue retrieves a secure configuration value
func (s *SecurityContext) GetConfigValue(key string) ([]byte, error) {
	return s.config.Get(key)
}