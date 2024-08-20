// https://www.ietf.org/archive/id/draft-ietf-oauth-v2-http-mac-01.pdf
package mac

import (
	"brainrot/pkg/util"
)

// Generate MAC key identifier and key
func GenerateIDAndKey() (id, key string, err error) {
	id, err = util.GenerateRandomBase64String()
	if err != nil {
		return id, key, err
	}

	key, err = util.GenerateRandomBase64String()
	if err != nil {
		return id, key, err
	}

	return id, key, err
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
