package random

import (
	"crypto/rand"
	"encoding/hex"
)

func NewRandomString(length int) string {
	byteLen := (length + 1) / 2
	bytes := make([]byte, byteLen)
	_, _ = rand.Read(bytes)

	return hex.EncodeToString(bytes)[:length]
}