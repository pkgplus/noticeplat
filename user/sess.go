package user

import (
	"crypto/rand"
	"encoding/hex"
)

func GetRandomID(size int) (string, error) {
	b := make([]byte, size)
	_, err := rand.Read(b)
	if err != nil {
		return string(b), err
	}

	return hex.EncodeToString(b), nil
}
