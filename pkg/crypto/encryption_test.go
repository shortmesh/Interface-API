package crypto

import (
	"encoding/base64"
	"os"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	testKey := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	os.Setenv("ENCRYPTION_KEY", testKey)
	encryptionKey, _ = getKey("ENCRYPTION_KEY")

	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "simple text",
			plaintext: "test@example.com",
		},
		{
			name:      "empty string",
			plaintext: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := Encrypt(tt.plaintext)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			if string(encrypted) == tt.plaintext {
				t.Error("Encrypted bytes should be different from plaintext")
			}

			decrypted, err := Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			if decrypted != tt.plaintext {
				t.Errorf("Decrypted text = %v, want %v", decrypted, tt.plaintext)
			}
		})
	}
}

func TestEncryptToBase64(t *testing.T) {
	testKey := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	os.Setenv("ENCRYPTION_KEY", testKey)
	encryptionKey, _ = getKey("ENCRYPTION_KEY")

	plaintext := "test@example.com"

	encrypted, err := EncryptToBase64(plaintext)
	if err != nil {
		t.Fatalf("EncryptToBase64() error = %v", err)
	}

	_, err = base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		t.Errorf("Encrypted result is not valid base64: %v", err)
	}

	decrypted, err := DecryptFromBase64(encrypted)
	if err != nil {
		t.Fatalf("DecryptFromBase64() error = %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypted text = %v, want %v", decrypted, plaintext)
	}
}

func TestEncryptWithoutKey(t *testing.T) {
	os.Setenv("ENCRYPTION_KEY", "")
	encryptionKey = nil

	_, err := Encrypt("test")
	if err != ErrMissingEncryptKey {
		t.Errorf("Expected ErrMissingEncryptKey, got %v", err)
	}

	_, err = Decrypt([]byte("test"))
	if err != ErrMissingEncryptKey {
		t.Errorf("Expected ErrMissingEncryptKey, got %v", err)
	}
}

func TestHash(t *testing.T) {
	testKey := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	os.Setenv("HASH_KEY", testKey)
	hashKey, _ = getKey("HASH_KEY")

	tests := []struct {
		name  string
		data1 string
		data2 string
		same  bool
	}{
		{
			name:  "same data produces same hash",
			data1: "test@example.com",
			data2: "test@example.com",
			same:  true,
		},
		{
			name:  "different data produces different hash",
			data1: "test@example.com",
			data2: "other@example.com",
			same:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1, err := Hash(tt.data1)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}

			hash2, err := Hash(tt.data2)
			if err != nil {
				t.Fatalf("Hash() error = %v", err)
			}

			if tt.same && string(hash1) != string(hash2) {
				t.Errorf("Expected same hash for identical data")
			}

			if !tt.same && string(hash1) == string(hash2) {
				t.Errorf("Expected different hash for different data")
			}
		})
	}
}

func TestHashToBase64(t *testing.T) {
	testKey := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	os.Setenv("HASH_KEY", testKey)
	hashKey, _ = getKey("HASH_KEY")

	data := "test@example.com"

	hash1, err := HashToBase64(data)
	if err != nil {
		t.Fatalf("HashToBase64() error = %v", err)
	}

	hash2, err := HashToBase64(data)
	if err != nil {
		t.Fatalf("HashToBase64() error = %v", err)
	}

	// Should produce same hash for same data
	if hash1 != hash2 {
		t.Error("Expected same hash for identical data")
	}

	// Should be valid base64
	_, err = base64.StdEncoding.DecodeString(hash1)
	if err != nil {
		t.Errorf("Hash result is not valid base64: %v", err)
	}
}

func TestInvalidKeys(t *testing.T) {
	tests := []struct {
		name      string
		keyValue  string
		shouldErr bool
	}{
		{
			name:      "valid 32 byte key",
			keyValue:  "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=",
			shouldErr: false,
		},
		{
			name:      "invalid base64",
			keyValue:  "not-valid-base64!!!",
			shouldErr: true,
		},
		{
			name:      "wrong size key (16 bytes)",
			keyValue:  "AAAAAAAAAAAAAAAAAAAAAA==",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("ENCRYPTION_KEY", tt.keyValue)
			key, err := getKey("ENCRYPTION_KEY")

			if tt.shouldErr && err == nil {
				t.Error("Expected error for invalid key, got none")
			}

			if !tt.shouldErr && err != nil {
				t.Errorf("Expected no error for valid key, got: %v", err)
			}

			if !tt.shouldErr && len(key) != 32 {
				t.Errorf("Expected 32 byte key, got %d bytes", len(key))
			}
		})
	}
}
