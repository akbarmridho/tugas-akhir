package utility

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// ComputeHMACSHA256 generates a SHA-256 HMAC hash using a secret and payload
func ComputeHMACSHA256(secret, payload string) string {
	// Create a new HMAC hash using SHA-256
	mac := hmac.New(sha256.New, []byte(secret))

	// Write the payload to the hash
	mac.Write([]byte(payload))

	// Get the final hash and convert to hex string
	return hex.EncodeToString(mac.Sum(nil))
}
