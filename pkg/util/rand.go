package util

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
)

func GenerateRandomBase64String() (string, error) {
	length := 32 // 256 bits

	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(randomBytes), nil
}

func GenerateRandomHexString() (string, error) {
	length := 32 // 256 bits

	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}
