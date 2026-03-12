package messagetemplate

import (
	"fmt"
	"os"
	"strings"

	"interface-api/pkg/logger"
)

func GetOTPTemplate() string {
	template := os.Getenv("OTP_MESSAGE_TEMPLATE")
	if template == "" {
		template = "Your OTP code is: {{code}}. It will expire at {{expiry}}."
	}
	return template
}

func FormatOTPMessage(code, expiry string) string {
	template := GetOTPTemplate()
	message := template
	message = strings.ReplaceAll(message, "{{code}}", code)
	message = strings.ReplaceAll(message, "{{expiry}}", expiry)

	logger.Debug(fmt.Sprintf("Formatted OTP message: %s", message))

	return message
}
