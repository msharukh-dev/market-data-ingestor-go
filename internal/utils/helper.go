package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashAPIKey(apiKey string) string {
	sum := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(sum[:])
}
