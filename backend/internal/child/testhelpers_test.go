package child

import (
	"testing"

	"github.com/trollstaven/nioplugget/backend/internal/auth"
)

// mustHashPIN hashes a PIN using Argon2id and panics on error (test helper).
func mustHashPIN(pin string) string {
	h, err := auth.HashPassword(pin)
	if err != nil {
		panic("mustHashPIN: " + err.Error())
	}
	return h
}

// Ensure mustHashPIN compiles.
var _ = testing.T{}
