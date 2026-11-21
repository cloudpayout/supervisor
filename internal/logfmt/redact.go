package logfmt

import "strings"

// Redact replaces each non-empty secret substring with [REDACTED].
func Redact(msg string, secrets []string) string {
	for _, secret := range secrets {
		if secret == "" {
			continue
		}
		msg = strings.ReplaceAll(msg, secret, "[REDACTED]")
	}
	return msg
}
