package child

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateInviteToken returns a URL-safe random token of at least 43 characters.
// It uses 32 bytes (256 bits) from crypto/rand encoded with base64 URL encoding.
func GenerateInviteToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
