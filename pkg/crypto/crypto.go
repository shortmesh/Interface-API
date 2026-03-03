package crypto

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"os"
)

var ErrMissingHashKey = errors.New("HASH_KEY not configured")

func Hash(data string) ([]byte, error) {
	hashKey := os.Getenv("HASH_KEY")
	if hashKey == "" {
		return nil, ErrMissingHashKey
	}

	key, err := base64.StdEncoding.DecodeString(hashKey)
	if err != nil {
		return nil, err
	}

	h := hmac.New(sha256.New, key)
	h.Write([]byte(data))

	return h.Sum(nil), nil
}

func HashToBase64(data string) (string, error) {
	hash, err := Hash(data)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(hash), nil
}

func GenerateSecureToken(length int) (string, error) {
	if length <= 0 {
		length = 32
	}

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(bytes), nil
}
