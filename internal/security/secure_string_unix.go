//go:build !windows
// +build !windows

package security

import (
	"syscall"
)

// lock attempts to lock the memory pages containing the secure data
// This prevents the data from being swapped to disk
func (s *SecureString) lock() error {
	// mlock requires memory to be page-aligned
	// For simplicity, we're using the data as-is
	// In production, you might want to allocate page-aligned memory
	return syscall.Mlock(s.data)
}

// unlock unlocks the memory pages
func (s *SecureString) unlock() error {
	return syscall.Munlock(s.data)
}