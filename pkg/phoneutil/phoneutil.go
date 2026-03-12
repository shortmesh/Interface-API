package phoneutil

import (
	"errors"
	"strings"

	"github.com/nyaruka/phonenumbers"
)

var ErrInvalidE164Format = errors.New("identifier must be in E.164 format (e.g., +1234567890)")

// ValidateE164 validates that the given phone number is in E.164 format
func ValidateE164(identifier string) error {
	if !strings.HasPrefix(identifier, "+") {
		return ErrInvalidE164Format
	}

	num, err := phonenumbers.Parse(identifier, "")
	if err != nil {
		return ErrInvalidE164Format
	}

	if !phonenumbers.IsValidNumber(num) {
		return ErrInvalidE164Format
	}

	formatted := phonenumbers.Format(num, phonenumbers.E164)
	if formatted != identifier {
		return ErrInvalidE164Format
	}

	return nil
}
