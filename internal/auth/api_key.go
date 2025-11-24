package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// GenerateUserAPIKey generates a new user API key and its hash. The key
// format is "sk-" followed by a 32-character random hex string.
func GenerateUserAPIKey() (plain string, hash string, err error) {
	b := make([]byte, 16) // 16 bytes -> 32 hex characters
	if _, err = rand.Read(b); err != nil {
		return "", "", err
	}
	random := hex.EncodeToString(b)
	plain = "sk-" + random
	hash = HashAPIKey(plain)
	return plain, hash, nil
}

// HashAPIKey computes a SHA-256 hex-encoded hash of the provided key.
func HashAPIKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}
