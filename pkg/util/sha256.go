package util

import "crypto/sha256"

func HashWithSalt(data []byte, salt []byte) []byte {
	hash := sha256.New()
	if salt != nil {
		hash.Write(salt)
	}

	hash.Write(data)
	return hash.Sum(nil)
}
