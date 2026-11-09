package instance

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"strings"
)

// NewID returns an opaque, URL-safe instance identifier (FR-002).
func NewID() (string, error) {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate instance id: %w", err)
	}
	enc := base32.StdEncoding.WithPadding(base32.NoPadding)
	return strings.ToLower(enc.EncodeToString(buf)), nil
}
