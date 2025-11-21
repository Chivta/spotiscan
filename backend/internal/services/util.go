package services

import (
	"crypto/rand"
	"encoding/hex"
)

func generateRandomString() string {
	b := make([]byte, 48)
	rand.Read(b)
	return hex.EncodeToString(b)
}
