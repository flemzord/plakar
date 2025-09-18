package security

import (
	"bytes"
	"testing"
)

func TestNewSecureString(t *testing.T) {
	data := []byte("sensitive-data-123")

	ss, err := NewSecureString(data)
	if err != nil {
		t.Fatalf("Failed to create SecureString: %v", err)
	}
	defer ss.Clear()

	if ss == nil {
		t.Fatal("NewSecureString returned nil")
	}

	if ss.Length() != len(data) {
		t.Errorf("Expected length %d, got %d", len(data), ss.Length())
	}
}

func TestSecureStringGet(t *testing.T) {
	original := []byte("test-secret")

	ss, err := NewSecureString(original)
	if err != nil {
		t.Fatalf("Failed to create SecureString: %v", err)
	}
	defer ss.Clear()

	retrieved, err := ss.Get()
	if err != nil {
		t.Fatalf("Failed to get data: %v", err)
	}

	if !bytes.Equal(original, retrieved) {
		t.Error("Retrieved data doesn't match original")
	}

	// Clear retrieved data
	for i := range retrieved {
		retrieved[i] = 0
	}
}

func TestSecureStringFromString(t *testing.T) {
	str := "sensitive-string"

	ss, err := NewSecureStringFromString(str)
	if err != nil {
		t.Fatalf("Failed to create SecureString from string: %v", err)
	}
	defer ss.Clear()

	retrieved, err := ss.GetString()
	if err != nil {
		t.Fatalf("Failed to get string: %v", err)
	}

	if retrieved != str {
		t.Errorf("Expected '%s', got '%s'", str, retrieved)
	}
}

func TestSecureStringEquals(t *testing.T) {
	data1 := []byte("secret-123")
	data2 := []byte("secret-123")
	data3 := []byte("different")

	ss1, _ := NewSecureString(data1)
	defer ss1.Clear()

	ss2, _ := NewSecureString(data2)
	defer ss2.Clear()

	ss3, _ := NewSecureString(data3)
	defer ss3.Clear()

	// Test equal SecureStrings
	equal, err := ss1.Equals(ss2)
	if err != nil {
		t.Fatalf("Failed to compare: %v", err)
	}
	if !equal {
		t.Error("Expected SecureStrings to be equal")
	}

	// Test unequal SecureStrings
	equal, err = ss1.Equals(ss3)
	if err != nil {
		t.Fatalf("Failed to compare: %v", err)
	}
	if equal {
		t.Error("Expected SecureStrings to be unequal")
	}
}

func TestSecureStringEqualsBytes(t *testing.T) {
	data := []byte("test-data")

	ss, _ := NewSecureString(data)
	defer ss.Clear()

	// Test with equal data
	equal, err := ss.EqualsBytes(data)
	if err != nil {
		t.Fatalf("Failed to compare: %v", err)
	}
	if !equal {
		t.Error("Expected data to be equal")
	}

	// Test with different data
	different := []byte("different")
	equal, err = ss.EqualsBytes(different)
	if err != nil {
		t.Fatalf("Failed to compare: %v", err)
	}
	if equal {
		t.Error("Expected data to be unequal")
	}
}

func TestSecureStringClear(t *testing.T) {
	data := []byte("to-be-cleared")

	ss, _ := NewSecureString(data)

	// Clear the SecureString
	ss.Clear()

	// Verify it's cleared
	if !ss.IsCleared() {
		t.Error("SecureString should be marked as cleared")
	}

	// Try to get data from cleared SecureString
	_, err := ss.Get()
	if err == nil {
		t.Error("Expected error when getting from cleared SecureString")
	}

	// Length should be 0 after clearing
	if ss.Length() != 0 {
		t.Error("Length should be 0 after clearing")
	}
}

func TestSecureStringEmptyInput(t *testing.T) {
	// Test with empty byte slice
	_, err := NewSecureString([]byte{})
	if err == nil {
		t.Error("Expected error for empty data")
	}

	// Test with empty string
	_, err = NewSecureStringFromString("")
	if err == nil {
		t.Error("Expected error for empty string")
	}
}

func TestSecureStringConcurrency(t *testing.T) {
	data := []byte("concurrent-test")
	ss, _ := NewSecureString(data)
	defer ss.Clear()

	done := make(chan bool)

	// Multiple goroutines reading
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				ss.Get()
				ss.Length()
				ss.IsCleared()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// If we get here without deadlock or panic, concurrency is working
}