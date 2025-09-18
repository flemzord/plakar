package security

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

// ValidateAndSanitizePath validates and sanitizes a file path for safe use
// It returns the cleaned path or an error if validation fails
func ValidateAndSanitizePath(path string) (string, error) {
	// First sanitize the path
	sanitized, err := SanitizePath(path)
	if err != nil {
		return "", fmt.Errorf("path validation failed: %w", err)
	}

	// Additional checks for absolute paths
	if filepath.IsAbs(path) {
		// Ensure the path is within acceptable boundaries
		// This is a basic check - you might want to add more restrictions
		if err := validateAbsolutePath(sanitized); err != nil {
			return "", err
		}
	}

	return sanitized, nil
}

// validateAbsolutePath performs additional validation for absolute paths
func validateAbsolutePath(path string) error {
	// Don't allow access to system directories
	restrictedPaths := []string{
		"/etc",
		"/sys",
		"/proc",
		"/dev",
		"/boot",
		"C:\\Windows",
		"C:\\Program Files",
	}

	for _, restricted := range restrictedPaths {
		if filepath.HasPrefix(path, restricted) {
			return fmt.Errorf("%w: access to system directory not allowed", ErrInvalidInput)
		}
	}

	return nil
}

// ValidateEnvVar validates an environment variable value
func ValidateEnvVar(value string) error {
	// Check for command injection patterns
	if ContainsSQLInjection(value) {
		return fmt.Errorf("%w: potential injection detected in environment variable", ErrInvalidInput)
	}

	// Check for shell injection patterns
	shellPatterns := []string{
		"$(", "${", "`", "\\n", "\\r", "\n", "\r",
		"&&", "||", ";", "|", "&", ">", "<",
	}

	for _, pattern := range shellPatterns {
		if strings.Contains(value, pattern) {
			return fmt.Errorf("%w: invalid characters in environment variable", ErrInvalidInput)
		}
	}

	return nil
}

// SanitizeArgs validates and sanitizes command-line arguments
func SanitizeArgs(args []string) ([]string, error) {
	sanitized := make([]string, 0, len(args))

	for _, arg := range args {
		// Remove null bytes
		arg = RemoveNullBytes(arg)

		// Check for injection patterns
		if ContainsSQLInjection(arg) {
			return nil, fmt.Errorf("%w: potential injection in argument", ErrInvalidInput)
		}

		// Basic length check
		if len(arg) > 4096 {
			return nil, fmt.Errorf("%w: argument too long", ErrInvalidInput)
		}

		sanitized = append(sanitized, arg)
	}

	return sanitized, nil
}

// ValidateFilePermissions checks if file permissions are secure
func ValidateFilePermissions(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot stat file: %w", err)
	}

	mode := info.Mode()

	// Check for world-writable files
	if mode.Perm()&0022 != 0 {
		return fmt.Errorf("%w: file is world-writable", ErrInvalidInput)
	}

	// Check for setuid/setgid bits
	if mode&os.ModeSetuid != 0 || mode&os.ModeSetgid != 0 {
		return fmt.Errorf("%w: file has setuid/setgid bit set", ErrInvalidInput)
	}

	return nil
}

// SecureDelete attempts to securely delete a file by overwriting it before removal
func SecureDelete(path string) error {
	// Validate path first
	cleaned, err := ValidateAndSanitizePath(path)
	if err != nil {
		return err
	}

	// Open file for writing
	file, err := os.OpenFile(cleaned, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("cannot open file for secure deletion: %w", err)
	}
	defer file.Close()

	// Get file size
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("cannot stat file: %w", err)
	}

	size := info.Size()

	// Overwrite with random data
	randomData := make([]byte, 4096)
	for written := int64(0); written < size; {
		toWrite := size - written
		if toWrite > 4096 {
			toWrite = 4096
		}

		// Fill with random data
		for i := range randomData[:toWrite] {
			randomData[i] = byte(rand.Intn(256))
		}

		n, err := file.Write(randomData[:toWrite])
		if err != nil {
			return fmt.Errorf("failed to overwrite file: %w", err)
		}
		written += int64(n)
	}

	// Sync to ensure data is written
	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	// Close and remove
	file.Close()
	return os.Remove(cleaned)
}

// CreateSecureFile creates a file with secure permissions (0600)
func CreateSecureFile(path string) (*os.File, error) {
	// Validate path
	cleaned, err := ValidateAndSanitizePath(path)
	if err != nil {
		return nil, err
	}

	// Create with secure permissions
	return os.OpenFile(cleaned, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0600)
}

// CreateSecureDirectory creates a directory with secure permissions (0700)
func CreateSecureDirectory(path string) error {
	// Validate path
	cleaned, err := ValidateAndSanitizePath(path)
	if err != nil {
		return err
	}

	// Create with secure permissions
	return os.MkdirAll(cleaned, 0700)
}