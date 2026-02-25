package crypto

import (
	"os"
	"strconv"

	"interface-api/pkg/logger"

	"github.com/alexedwards/argon2id"
)

func getArgon2Params() *argon2id.Params {
	params := &argon2id.Params{
		Memory:      65536,
		Iterations:  3,
		Parallelism: 4,
		SaltLength:  16,
		KeyLength:   32,
	}

	if val := os.Getenv("ARGON2_MEMORY"); val != "" {
		if memory, err := strconv.ParseUint(val, 10, 32); err == nil {
			params.Memory = uint32(memory)
		}
	}

	if val := os.Getenv("ARGON2_ITERATIONS"); val != "" {
		if iterations, err := strconv.ParseUint(val, 10, 32); err == nil {
			params.Iterations = uint32(iterations)
		}
	}

	if val := os.Getenv("ARGON2_PARALLELISM"); val != "" {
		if parallelism, err := strconv.ParseUint(val, 10, 8); err == nil {
			params.Parallelism = uint8(parallelism)
		}
	}

	if val := os.Getenv("ARGON2_SALT_LENGTH"); val != "" {
		if saltLength, err := strconv.ParseUint(val, 10, 32); err == nil {
			params.SaltLength = uint32(saltLength)
		}
	}

	if val := os.Getenv("ARGON2_KEY_LENGTH"); val != "" {
		if keyLength, err := strconv.ParseUint(val, 10, 32); err == nil {
			params.KeyLength = uint32(keyLength)
		}
	}

	logger.Log.Debugf(
		"Argon2 params: mem=%dKB iter=%d par=%d salt=%dB key=%dB",
		params.Memory,
		params.Iterations,
		params.Parallelism,
		params.SaltLength,
		params.KeyLength,
	)
	return params
}

func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, getArgon2Params())
}

func VerifyPassword(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}
