package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"

	"interface-api/internal/logger"

	_ "github.com/joho/godotenv/autoload"
)

var (
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	ErrMissingEncryptKey = errors.New("ENCRYPTION_KEY not configured")
	ErrMissingHashKey    = errors.New("HASH_KEY not configured")
	ErrInvalidKeyFormat  = errors.New("key must be base64 encoded")
	ErrInvalidKeySize    = errors.New("key must be exactly 32 bytes when decoded")
)

var (
	encryptionKey []byte
	hashKey       []byte
)

func init() {
	var err error

	encryptionKey, err = getKey("ENCRYPTION_KEY")
	if err != nil {
		logger.Log.Errorf("Failed to load ENCRYPTION_KEY: %v", err)
	}

	hashKey, err = getKey("HASH_KEY")
	if err != nil {
		logger.Log.Errorf("Failed to load HASH_KEY: %v", err)
	}
}

func getKey(envVar string) ([]byte, error) {
	keyStr := os.Getenv(envVar)
	if keyStr == "" {
		return nil, nil
	}

	key, err := base64.StdEncoding.DecodeString(keyStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w: %v", envVar, ErrInvalidKeyFormat, err)
	}

	if len(key) != 32 {
		return nil, fmt.Errorf("%s: %w (got %d bytes)", envVar, ErrInvalidKeySize, len(key))
	}

	return key, nil
}

func Encrypt(plaintext string) ([]byte, error) {
	if encryptionKey == nil {
		return nil, ErrMissingEncryptKey
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)

	return ciphertext, nil
}

func Decrypt(ciphertext []byte) (string, error) {
	if encryptionKey == nil {
		return "", ErrMissingEncryptKey
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	return string(plaintext), nil
}

func EncryptToBase64(plaintext string) (string, error) {
	ciphertext, err := Encrypt(plaintext)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func DecryptFromBase64(ciphertextB64 string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextB64)
	if err != nil {
		return "", err
	}
	return Decrypt(ciphertext)
}

func Hash(data string) ([]byte, error) {
	if hashKey == nil {
		return nil, ErrMissingHashKey
	}

	h := hmac.New(sha256.New, hashKey)
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
