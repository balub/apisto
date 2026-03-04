package auth

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateToken generates a cryptographically random hex token of the given byte length.
func GenerateToken(byteLength int) (string, error) {
	b := make([]byte, byteLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
