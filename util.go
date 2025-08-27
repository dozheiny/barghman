package main

import (
	"crypto/rand"
	"encoding/hex"
)

// generateBoundary creates a random MIME boundary string.
func generateBoundary() string {
	b := make([]byte, 16) // 16 random bytes
	if _, err := rand.Read(b); err != nil {
		// fallback
		return "simpleboundary123456"
	}
	return "boundary_" + hex.EncodeToString(b)
}
