package security

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

// SecureConfig manages sensitive configuration values using SecureString
type SecureConfig struct {
	mu      sync.RWMutex
	secrets map[string]*SecureString
}

// NewSecureConfig creates a new secure configuration manager
func NewSecureConfig() *SecureConfig {
	return &SecureConfig{
		secrets: make(map[string]*SecureString),
	}
}

// Set stores a sensitive value with the given key
func (sc *SecureConfig) Set(key string, value []byte) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}

	// Validate the key to prevent injection
	if err := ValidateAlphanumeric(strings.ReplaceAll(key, "_", "")); err != nil {
		return fmt.Errorf("invalid config key: %w", err)
	}

	secureStr, err := NewSecureString(value)
	if err != nil {
		return fmt.Errorf("failed to create secure string: %w", err)
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Clear old value if exists
	if old, exists := sc.secrets[key]; exists {
		old.Clear()
	}

	sc.secrets[key] = secureStr
	return nil
}

// SetString stores a sensitive string value with the given key
func (sc *SecureConfig) SetString(key string, value string) error {
	return sc.Set(key, []byte(value))
}

// Get retrieves a sensitive value by key
func (sc *SecureConfig) Get(key string) ([]byte, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	secureStr, exists := sc.secrets[key]
	if !exists {
		return nil, fmt.Errorf("key %s not found", key)
	}

	return secureStr.Get()
}

// GetString retrieves a sensitive string value by key
func (sc *SecureConfig) GetString(key string) (string, error) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	secureStr, exists := sc.secrets[key]
	if !exists {
		return "", fmt.Errorf("key %s not found", key)
	}

	return secureStr.GetString()
}

// Has checks if a key exists in the configuration
func (sc *SecureConfig) Has(key string) bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	_, exists := sc.secrets[key]
	return exists
}

// Delete removes a sensitive value and clears its memory
func (sc *SecureConfig) Delete(key string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if secureStr, exists := sc.secrets[key]; exists {
		secureStr.Clear()
		delete(sc.secrets, key)
	}
}

// Clear removes all sensitive values and clears their memory
func (sc *SecureConfig) Clear() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	for _, secureStr := range sc.secrets {
		secureStr.Clear()
	}

	sc.secrets = make(map[string]*SecureString)
}

// Keys returns a list of all configuration keys (not the values)
func (sc *SecureConfig) Keys() []string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	keys := make([]string, 0, len(sc.secrets))
	for k := range sc.secrets {
		keys = append(keys, k)
	}
	return keys
}