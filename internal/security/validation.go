package security

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

var (
	// ErrInvalidInput indicates that the input failed validation
	ErrInvalidInput = errors.New("invalid input")

	// Common regex patterns for validation
	alphanumericRegex = regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	emailRegex       = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	uuidRegex        = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	sha256Regex      = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

	// Dangerous patterns that might indicate injection attempts
	sqlPatterns = []string{
		"';", "--", "/*", "*/", "xp_", "sp_", "exec", "execute",
		"select", "insert", "update", "delete", "drop", "create",
		"union", "having", "group by", "order by",
	}

	// Path traversal patterns
	pathTraversalPatterns = []string{
		"../", "..\\", "..", "~", "%00", "%2e%2e",
	}
)

// Validator provides input validation functionality
type Validator struct {
	maxLength int
	minLength int
	required  bool
	patterns  []*regexp.Regexp
}

// NewValidator creates a new validator with default settings
func NewValidator() *Validator {
	return &Validator{
		maxLength: 1024,
		minLength: 0,
		required:  false,
	}
}

// WithMaxLength sets the maximum allowed length
func (v *Validator) WithMaxLength(length int) *Validator {
	v.maxLength = length
	return v
}

// WithMinLength sets the minimum required length
func (v *Validator) WithMinLength(length int) *Validator {
	v.minLength = length
	return v
}

// WithRequired marks the input as required (non-empty)
func (v *Validator) WithRequired() *Validator {
	v.required = true
	return v
}

// WithPattern adds a regex pattern that the input must match
func (v *Validator) WithPattern(pattern *regexp.Regexp) *Validator {
	v.patterns = append(v.patterns, pattern)
	return v
}

// Validate checks if the input meets all validation criteria
func (v *Validator) Validate(input string) error {
	// Check if required
	if v.required && len(input) == 0 {
		return fmt.Errorf("%w: input is required", ErrInvalidInput)
	}

	// Check length constraints
	if len(input) > v.maxLength {
		return fmt.Errorf("%w: input exceeds maximum length of %d", ErrInvalidInput, v.maxLength)
	}

	if len(input) < v.minLength {
		return fmt.Errorf("%w: input is shorter than minimum length of %d", ErrInvalidInput, v.minLength)
	}

	// Check patterns
	for _, pattern := range v.patterns {
		if !pattern.MatchString(input) {
			return fmt.Errorf("%w: input does not match required pattern", ErrInvalidInput)
		}
	}

	return nil
}

// SanitizePath cleans and validates file paths to prevent directory traversal
func SanitizePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%w: empty path", ErrInvalidInput)
	}

	// Clean the path
	cleaned := filepath.Clean(path)

	// Check for path traversal attempts
	for _, pattern := range pathTraversalPatterns {
		if strings.Contains(path, pattern) {
			return "", fmt.Errorf("%w: potential path traversal detected", ErrInvalidInput)
		}
	}

	// Ensure the path doesn't start with dangerous characters
	if strings.HasPrefix(cleaned, "~") || strings.HasPrefix(cleaned, "/") && !filepath.IsAbs(path) {
		return "", fmt.Errorf("%w: invalid path prefix", ErrInvalidInput)
	}

	// Check for null bytes
	if strings.Contains(cleaned, "\x00") {
		return "", fmt.Errorf("%w: null byte in path", ErrInvalidInput)
	}

	return cleaned, nil
}

// SanitizeFilename validates and cleans filenames
func SanitizeFilename(filename string) (string, error) {
	if filename == "" {
		return "", fmt.Errorf("%w: empty filename", ErrInvalidInput)
	}

	// Remove any directory components
	filename = filepath.Base(filename)

	// Check for special filenames
	if filename == "." || filename == ".." {
		return "", fmt.Errorf("%w: invalid filename", ErrInvalidInput)
	}

	// Remove dangerous characters
	var sanitized strings.Builder
	for _, r := range filename {
		if unicode.IsLetter(r) || unicode.IsDigit(r) ||
		   r == '-' || r == '_' || r == '.' {
			sanitized.WriteRune(r)
		}
	}

	result := sanitized.String()
	if result == "" {
		return "", fmt.Errorf("%w: filename contains no valid characters", ErrInvalidInput)
	}

	// Ensure it doesn't start or end with a dot
	result = strings.Trim(result, ".")

	return result, nil
}

// SanitizeString removes potentially dangerous characters from a string
func SanitizeString(input string, allowedChars string) string {
	if allowedChars == "" {
		// Default to alphanumeric plus common safe characters
		allowedChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_. "
	}

	var sanitized strings.Builder
	for _, r := range input {
		if strings.ContainsRune(allowedChars, r) {
			sanitized.WriteRune(r)
		}
	}

	return sanitized.String()
}

// ValidateEmail checks if the input is a valid email address
func ValidateEmail(email string) error {
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("%w: invalid email format", ErrInvalidInput)
	}
	return nil
}

// ValidateUUID checks if the input is a valid UUID
func ValidateUUID(uuid string) error {
	if !uuidRegex.MatchString(uuid) {
		return fmt.Errorf("%w: invalid UUID format", ErrInvalidInput)
	}
	return nil
}

// ValidateSHA256 checks if the input is a valid SHA256 hash
func ValidateSHA256(hash string) error {
	if !sha256Regex.MatchString(hash) {
		return fmt.Errorf("%w: invalid SHA256 format", ErrInvalidInput)
	}
	return nil
}

// ValidateAlphanumeric checks if the input contains only alphanumeric characters
func ValidateAlphanumeric(input string) error {
	if !alphanumericRegex.MatchString(input) {
		return fmt.Errorf("%w: input must be alphanumeric", ErrInvalidInput)
	}
	return nil
}

// ContainsSQLInjection checks for potential SQL injection patterns
func ContainsSQLInjection(input string) bool {
	lowerInput := strings.ToLower(input)
	for _, pattern := range sqlPatterns {
		if strings.Contains(lowerInput, pattern) {
			return true
		}
	}
	return false
}

// ValidateRepositoryName validates a repository name
func ValidateRepositoryName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: repository name cannot be empty", ErrInvalidInput)
	}

	if len(name) > 255 {
		return fmt.Errorf("%w: repository name too long", ErrInvalidInput)
	}

	// Allow alphanumeric, dash, underscore, and dot
	validChars := regexp.MustCompile(`^[a-zA-Z0-9\-_.]+$`)
	if !validChars.MatchString(name) {
		return fmt.Errorf("%w: repository name contains invalid characters", ErrInvalidInput)
	}

	// Must not start or end with special characters
	if strings.HasPrefix(name, "-") || strings.HasPrefix(name, ".") ||
	   strings.HasSuffix(name, "-") || strings.HasSuffix(name, ".") {
		return fmt.Errorf("%w: repository name cannot start or end with special characters", ErrInvalidInput)
	}

	return nil
}

// ValidateSnapshotID validates a snapshot identifier
func ValidateSnapshotID(id string) error {
	// Snapshot IDs can be UUIDs or SHA256 hashes
	if err := ValidateUUID(id); err == nil {
		return nil
	}

	if err := ValidateSHA256(id); err == nil {
		return nil
	}

	// Also allow short hashes (minimum 7 characters)
	if len(id) >= 7 && len(id) <= 64 {
		if regexp.MustCompile(`^[a-fA-F0-9]+$`).MatchString(id) {
			return nil
		}
	}

	return fmt.Errorf("%w: invalid snapshot ID format", ErrInvalidInput)
}

// TruncateString safely truncates a string to a maximum length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}

	// Handle UTF-8 properly
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}

	return string(runes[:maxLen])
}

// RemoveNullBytes removes null bytes from a string
func RemoveNullBytes(s string) string {
	return strings.ReplaceAll(s, "\x00", "")
}

// EscapeHTML escapes HTML special characters
func EscapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}

// NormalizeWhitespace replaces all whitespace sequences with a single space
func NormalizeWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}