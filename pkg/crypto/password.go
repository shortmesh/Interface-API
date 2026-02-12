package crypto

import (
	"os"
	"strconv"

	"github.com/alexedwards/argon2id"
	_ "github.com/joho/godotenv/autoload"
)

var argon2Params *argon2id.Params

func init() {
	argon2Params = getArgon2Params()
}

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

	return params
}

func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2Params)
}

func VerifyPassword(password, hash string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}
