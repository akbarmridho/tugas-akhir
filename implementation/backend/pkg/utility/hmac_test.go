package utility

import (
	"testing"
)

// LLM generated, of course
// still funny how the test code are longer than the wrapper code

func TestComputeHMACSHA256(t *testing.T) {
	tests := []struct {
		name     string
		secret   string
		payload  string
		expected string
	}{
		{
			name:     "Basic HMAC-SHA256",
			secret:   "secret-key",
			payload:  "hello world",
			expected: "095d5a21fe6d0646db223fdf3de6436bb8dfb2fab0b51677ecf6441fcf5f2a67",
		},
		{
			name:     "Special characters",
			secret:   "!@#$%^&*()",
			payload:  "こんにちは世界",
			expected: "7eb301b008e1c6dfb0b40b83aa401294a57caa81a1ee8802613ab86102af2191",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeHMACSHA256(tt.secret, tt.payload)
			if result != tt.expected {
				t.Errorf("ComputeHMACSHA256() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestKnownValues tests against known values from other implementations
func TestKnownValues(t *testing.T) {
	// Test case from RFC 4231 (Test Case 1)
	secret := string([]byte{0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b, 0x0b})
	payload := "Hi There"
	expected := "b0344c61d8db38535ca8afceaf0bf12b881dc200c9833da726e9376c2e32cff7"

	result := ComputeHMACSHA256(secret, payload)
	if result != expected {
		t.Errorf("RFC 4231 Test Case failed: got %v, want %v", result, expected)
	}
}

// TestConsistency verifies that multiple calls with the same input produce the same output
func TestConsistency(t *testing.T) {
	secret := "test-secret"
	payload := "test-payload"

	result1 := ComputeHMACSHA256(secret, payload)
	result2 := ComputeHMACSHA256(secret, payload)

	if result1 != result2 {
		t.Errorf("Inconsistent results: %v != %v", result1, result2)
	}
}

// TestDifferentInputs verifies that different inputs produce different outputs
func TestDifferentInputs(t *testing.T) {
	secret := "same-secret"
	payload1 := "payload1"
	payload2 := "payload2"

	result1 := ComputeHMACSHA256(secret, payload1)
	result2 := ComputeHMACSHA256(secret, payload2)

	if result1 == result2 {
		t.Errorf("Different inputs produced same output: %v", result1)
	}
}
