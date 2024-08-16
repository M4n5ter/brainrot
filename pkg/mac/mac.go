// https://www.ietf.org/archive/id/draft-ietf-oauth-v2-http-mac-01.pdf
package mac

import (
	"crypto/rand"
	"encoding/base64"
)

// Generate MAC key identifier and key
func GenerateIDAndKey() (id, key string, err error) {
	id, err = GenerateRandomString()
	if err != nil {
		return id, key, err
	}

	key, err = GenerateRandomString()
	if err != nil {
		return id, key, err
	}

	return id, key, err
}

func GenerateRandomString(opts ...Option) (string, error) {
	length := 32

	for _, opt := range opts {
		length = opt.Length
	}

	randomBytes := make([]byte, length)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(randomBytes), nil
}

type Option struct {
	// Length of the random bytes
	Length int
}

type Request struct {
	ID    string
	TS    string
	Nonce string
	EXT   string
	MAC   string
}

func NewRequest(id, ts, nonce, ext, mac string) Request {
	return Request{
		ID:    id,
		TS:    ts,
		Nonce: nonce,
		EXT:   ext,
		MAC:   mac,
	}
}
