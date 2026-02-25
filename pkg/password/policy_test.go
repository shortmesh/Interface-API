package password

import (
	"os"
	"testing"
)

func TestValidatePassword(t *testing.T) {
	originalEnabled := os.Getenv("PASSWORD_POLICY_ENABLED")
	originalCheckPwned := os.Getenv("PASSWORD_CHECK_PWNED")
	defer func() {
		os.Setenv("PASSWORD_POLICY_ENABLED", originalEnabled)
		os.Setenv("PASSWORD_CHECK_PWNED", originalCheckPwned)
	}()

	tests := []struct {
		name        string
		password    string
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "valid password",
			password: "MySecureP@ssw0rd123!",
			wantErr:  false,
		},
		{
			name:        "too short",
			password:    "short",
			wantErr:     true,
			expectedErr: ErrPasswordTooShort,
		},
		{
			name:        "too long",
			password:    string(make([]byte, 65)),
			wantErr:     true,
			expectedErr: ErrPasswordTooLong,
		},
		{
			name:        "leading space",
			password:    " validpassword123",
			wantErr:     true,
			expectedErr: ErrPasswordContainsSpaces,
		},
		{
			name:        "trailing space",
			password:    "validpassword123 ",
			wantErr:     true,
			expectedErr: ErrPasswordContainsSpaces,
		},
	}

	os.Setenv("PASSWORD_POLICY_ENABLED", "true")
	os.Setenv("PASSWORD_CHECK_PWNED", "false")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.expectedErr != nil && err != tt.expectedErr {
				t.Errorf("ValidatePassword() error = %v, expectedErr %v", err, tt.expectedErr)
			}
		})
	}
}

func TestValidatePasswordDisabled(t *testing.T) {
	originalEnabled := os.Getenv("PASSWORD_POLICY_ENABLED")
	defer os.Setenv("PASSWORD_POLICY_ENABLED", originalEnabled)

	os.Setenv("PASSWORD_POLICY_ENABLED", "false")

	err := ValidatePassword("123")
	if err != nil {
		t.Errorf("Expected no error when policy disabled, got: %v", err)
	}
}

func TestIsPwned(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{
			name:     "known pwned password",
			password: "password",
			want:     true,
		},
		{
			name:     "likely secure password",
			password: "MyVerySecureP@ssw0rd!2024#UniqueString",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsPwned(tt.password, GetPolicyConfig().PwnedTimeout)
			if err != nil {
				t.Logf("Warning: API call failed: %v", err)
				t.Skip("Skipping due to API error")
			}
			if got != tt.want {
				t.Errorf("IsPwned() = %v, want %v", got, tt.want)
			}
		})
	}
}
