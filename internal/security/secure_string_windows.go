//go:build windows
// +build windows

package security

import (
	"syscall"
	"unsafe"
)

var (
	kernel32           = syscall.NewLazyDLL("kernel32.dll")
	virtualLockProc   = kernel32.NewProc("VirtualLock")
	virtualUnlockProc = kernel32.NewProc("VirtualUnlock")
)

// lock attempts to lock the memory pages containing the secure data
// This prevents the data from being swapped to disk on Windows
func (s *SecureString) lock() error {
	if len(s.data) == 0 {
		return nil
	}

	ret, _, err := virtualLockProc.Call(
		uintptr(unsafe.Pointer(&s.data[0])),
		uintptr(len(s.data)),
	)

	if ret == 0 {
		return err
	}

	s.locked = true
	return nil
}

// unlock unlocks the memory pages on Windows
func (s *SecureString) unlock() error {
	if len(s.data) == 0 {
		return nil
	}

	ret, _, err := virtualUnlockProc.Call(
		uintptr(unsafe.Pointer(&s.data[0])),
		uintptr(len(s.data)),
	)

	if ret == 0 {
		return err
	}

	return nil
}