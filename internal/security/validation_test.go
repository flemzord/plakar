package security

import (
	"strings"
	"testing"
)

func TestValidator(t *testing.T) {
	tests := []struct {
		name      string
		validator *Validator
		input     string
		wantErr   bool
	}{
		{
			name:      "valid input",
			validator: NewValidator().WithMinLength(5).WithMaxLength(20),
			input:     "validinput",
			wantErr:   false,
		},
		{
			name:      "too short",
			validator: NewValidator().WithMinLength(10),
			input:     "short",
			wantErr:   true,
		},
		{
			name:      "too long",
			validator: NewValidator().WithMaxLength(5),
			input:     "toolongstring",
			wantErr:   true,
		},
		{
			name:      "required but empty",
			validator: NewValidator().WithRequired(),
			input:     "",
			wantErr:   true,
		},
		{
			name:      "pattern mismatch",
			validator: NewValidator().WithPattern(alphanumericRegex),
			input:     "invalid-chars!",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.validator.Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSanitizePath(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name:    "valid path",
			path:    "dir/subdir/file.txt",
			want:    "dir/subdir/file.txt",
			wantErr: false,
		},
		{
			name:    "path traversal attempt",
			path:    "../../../etc/passwd",
			want:    "",
			wantErr: true,
		},
		{
			name:    "null byte",
			path:    "file\x00.txt",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty path",
			path:    "",
			want:    "",
			wantErr: true,
		},
		{
			name:    "tilde expansion attempt",
			path:    "~/secret",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SanitizePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:    "valid filename",
			input:   "document.pdf",
			want:    "document.pdf",
			wantErr: false,
		},
		{
			name:    "filename with path",
			input:   "/path/to/file.txt",
			want:    "file.txt",
			wantErr: false,
		},
		{
			name:    "special chars removed",
			input:   "file@#$%.txt",
			want:    "file.txt",
			wantErr: false,
		},
		{
			name:    "dots at edges",
			input:   ".hidden.",
			want:    "hidden",
			wantErr: false,
		},
		{
			name:    "parent directory",
			input:   "..",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty",
			input:   "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SanitizeFilename(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SanitizeFilename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("SanitizeFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		allowedChars string
		want         string
	}{
		{
			name:         "default allowed chars",
			input:        "Hello-World_123!@#",
			allowedChars: "",
			want:         "Hello-World_123",
		},
		{
			name:         "custom allowed chars",
			input:        "abc123xyz",
			allowedChars: "abc",
			want:         "abc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeString(tt.input, tt.allowedChars)
			if got != tt.want {
				t.Errorf("SanitizeString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email   string
		wantErr bool
	}{
		{"user@example.com", false},
		{"user.name@example.co.uk", false},
		{"user+tag@example.com", false},
		{"invalid", true},
		{"@example.com", true},
		{"user@", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail(%s) error = %v, wantErr %v", tt.email, err, tt.wantErr)
			}
		})
	}
}

func TestValidateUUID(t *testing.T) {
	tests := []struct {
		uuid    string
		wantErr bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", false},
		{"550E8400-E29B-41D4-A716-446655440000", false},
		{"not-a-uuid", true},
		{"550e8400-e29b-41d4-a716", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.uuid, func(t *testing.T) {
			err := ValidateUUID(tt.uuid)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUUID(%s) error = %v, wantErr %v", tt.uuid, err, tt.wantErr)
			}
		})
	}
}

func TestValidateSHA256(t *testing.T) {
	tests := []struct {
		hash    string
		wantErr bool
	}{
		{"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", false},
		{"E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855", false},
		{"not-a-hash", true},
		{"e3b0c44298fc1c14", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.hash, func(t *testing.T) {
			err := ValidateSHA256(tt.hash)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSHA256(%s) error = %v, wantErr %v", tt.hash, err, tt.wantErr)
			}
		})
	}
}

func TestContainsSQLInjection(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"normal input", false},
		{"SELECT * FROM users", true},
		{"'; DROP TABLE users; --", true},
		{"1' OR '1'='1", true},
		{"/* comment */", true},
		{"exec sp_help", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ContainsSQLInjection(tt.input)
			if got != tt.want {
				t.Errorf("ContainsSQLInjection(%s) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateRepositoryName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid name", "my-repo", false},
		{"valid with underscore", "my_repo", false},
		{"valid with dot", "my.repo", false},
		{"valid alphanumeric", "repo123", false},
		{"empty", "", true},
		{"too long", strings.Repeat("a", 256), true},
		{"invalid chars", "repo@#$", true},
		{"starts with dash", "-repo", true},
		{"ends with dot", "repo.", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRepositoryName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRepositoryName(%s) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestValidateSnapshotID(t *testing.T) {
	tests := []struct {
		id      string
		wantErr bool
	}{
		// Valid UUIDs
		{"550e8400-e29b-41d4-a716-446655440000", false},
		// Valid SHA256
		{"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", false},
		// Valid short hash
		{"a1b2c3d", false},
		// Invalid
		{"xyz", true},
		{"", true},
		{"not-valid-id", true},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			err := ValidateSnapshotID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSnapshotID(%s) error = %v, wantErr %v", tt.id, err, tt.wantErr)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"this is too long", 7, "this is"},
		{"", 5, ""},
		{"exact", 5, "exact"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := TruncateString(tt.input, tt.maxLen)
			if got != tt.want {
				t.Errorf("TruncateString(%s, %d) = %v, want %v", tt.input, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestRemoveNullBytes(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"normal", "normal"},
		{"with\x00null", "withnull"},
		{"\x00start", "start"},
		{"end\x00", "end"},
		{"\x00\x00", ""},
	}

	for _, tt := range tests {
		t.Run("test", func(t *testing.T) {
			got := RemoveNullBytes(tt.input)
			if got != tt.want {
				t.Errorf("RemoveNullBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"normal text", "normal text"},
		{"<script>alert('xss')</script>", "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
		{"a & b", "a &amp; b"},
		{`"quoted"`, "&quot;quoted&quot;"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := EscapeHTML(tt.input)
			if got != tt.want {
				t.Errorf("EscapeHTML(%s) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeWhitespace(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"normal text", "normal text"},
		{"  extra   spaces  ", "extra spaces"},
		{"tabs\tand\nnewlines", "tabs and newlines"},
		{"\n\n\n", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NormalizeWhitespace(tt.input)
			if got != tt.want {
				t.Errorf("NormalizeWhitespace(%s) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}