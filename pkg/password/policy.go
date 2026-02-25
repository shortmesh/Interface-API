package password

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrPasswordTooShort       = errors.New("password must be at least %d characters")
	ErrPasswordTooLong        = errors.New("password exceeds maximum length of %d characters")
	ErrPasswordPwned          = errors.New("password has been exposed in a data breach")
	ErrPasswordContainsSpaces = errors.New("password cannot contain leading or trailing spaces")
)

const (
	MinPasswordLength    = 8
	MaxPasswordLength    = 64
	DefaultPolicyEnabled = true
	PwnedPasswordsAPI    = "https://api.pwnedpasswords.com/range/"
)

type PolicyConfig struct {
	Enabled        bool
	MinLength      int
	MaxLength      int
	CheckPwned     bool
	CheckSpaces    bool
	PwnedTimeout   time.Duration
	SkipPwnedOnErr bool
}

func GetPolicyConfig() *PolicyConfig {
	config := &PolicyConfig{
		Enabled:        DefaultPolicyEnabled,
		MinLength:      MinPasswordLength,
		MaxLength:      MaxPasswordLength,
		CheckPwned:     true,
		CheckSpaces:    true,
		PwnedTimeout:   5 * time.Second,
		SkipPwnedOnErr: true,
	}

	if val := os.Getenv("PASSWORD_POLICY_ENABLED"); val != "" {
		if enabled, err := strconv.ParseBool(val); err == nil {
			config.Enabled = enabled
		}
	}

	if val := os.Getenv("PASSWORD_MIN_LENGTH"); val != "" {
		if minLen, err := strconv.Atoi(val); err == nil && minLen >= 8 {
			config.MinLength = minLen
		}
	}

	if val := os.Getenv("PASSWORD_MAX_LENGTH"); val != "" {
		if maxLen, err := strconv.Atoi(val); err == nil && maxLen <= 128 {
			config.MaxLength = maxLen
		}
	}

	if val := os.Getenv("PASSWORD_CHECK_PWNED"); val != "" {
		if check, err := strconv.ParseBool(val); err == nil {
			config.CheckPwned = check
		}
	}

	if val := os.Getenv("PASSWORD_CHECK_SPACES"); val != "" {
		if check, err := strconv.ParseBool(val); err == nil {
			config.CheckSpaces = check
		}
	}

	if val := os.Getenv("PASSWORD_PWNED_TIMEOUT"); val != "" {
		if timeout, err := strconv.Atoi(val); err == nil && timeout > 0 {
			config.PwnedTimeout = time.Duration(timeout) * time.Second
		}
	}

	if val := os.Getenv("PASSWORD_SKIP_PWNED_ON_ERROR"); val != "" {
		if skip, err := strconv.ParseBool(val); err == nil {
			config.SkipPwnedOnErr = skip
		}
	}

	return config
}

// ValidatePassword validates password according to NIST SP 800-63B guidelines
func ValidatePassword(password string) error {
	config := GetPolicyConfig()

	if !config.Enabled {
		return nil
	}

	length := utf8.RuneCountInString(password)
	if length < config.MinLength {
		return fmt.Errorf(ErrPasswordTooShort.Error(), config.MinLength)
	}

	if length > config.MaxLength {
		return fmt.Errorf(ErrPasswordTooLong.Error(), config.MaxLength)
	}

	if config.CheckSpaces {
		if strings.TrimSpace(password) != password {
			return ErrPasswordContainsSpaces
		}
	}

	if config.CheckPwned {
		isPwned, err := IsPwned(password, config.PwnedTimeout)
		if err != nil && !config.SkipPwnedOnErr {
			return fmt.Errorf("password breach check failed: %w", err)
		}
		if isPwned {
			return ErrPasswordPwned
		}
	}

	return nil
}

var pwnedHTTPClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 2,
		IdleConnTimeout:     30 * time.Second,
	},
}

// IsPwned checks if a password has been exposed in a data breach using the Pwned Passwords API with k-Anonymity
func IsPwned(password string, timeout time.Duration) (bool, error) {
	h := sha1.New()
	h.Write([]byte(password))
	hashBytes := h.Sum(nil)
	hashStr := strings.ToUpper(hex.EncodeToString(hashBytes))

	prefix := hashStr[:5]
	suffix := hashStr[5:]

	pwnedHTTPClient.Timeout = timeout

	resp, err := pwnedHTTPClient.Get(PwnedPasswordsAPI + prefix)
	if err != nil {
		return false, fmt.Errorf("pwned passwords API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("pwned passwords API returned status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("pwned passwords API response read failed: %w", err)
	}

	lines := strings.SplitSeq(string(body), "\n")
	for line := range lines {
		parts := strings.Split(line, ":")
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == suffix {
			return true, nil
		}
	}

	return false, nil
}
