package util

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
)

// D 必须能被 encoding/json 序列化。
type AES[D any] struct {
	secret []byte
}

func NewAES[D any](secret string) *AES[D] {
	return &AES[D]{secret: ToMakabakaBytes(secret)}
}

func (a *AES[D]) Encrypt(data D) (string, error) {
	block, err := aes.NewCipher(a.secret)
	if err != nil {
		return "", err
	}

	plaintext, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (a *AES[D]) Decrypt(token string) (D, error) {
	var data D

	ciphertext, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return data, err
	}

	block, err := aes.NewCipher(a.secret)
	if err != nil {
		return data, err
	}

	if len(ciphertext) < aes.BlockSize {
		return data, fmt.Errorf("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	if err := json.Unmarshal(ciphertext, &data); err != nil {
		return data, err
	}

	return data, nil
}
