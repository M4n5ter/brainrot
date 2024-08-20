package apikey

import "brainrot/pkg/util"

func GenerateAPIKey() (string, error) {
	return util.GenerateRandomHexString()
}
