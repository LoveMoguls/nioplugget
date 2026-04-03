package child

import (
	"testing"

	"github.com/trollstaven/nioplugget/backend/internal/auth"
)

func init() {
	// Initialize the package-level TokenAuth so auth.GenerateToken works in tests.
	auth.NewTokenAuth("test-secret-for-child-handler-tests")
}

// mustHashPIN hashes a PIN using Argon2id and panics on error (test helper).
func mustHashPIN(pin string) string {
	h, err := auth.HashPassword(pin)
	if err != nil {
		panic("mustHashPIN: " + err.Error())
	}
	return h
}

// Ensure testing import is used.
var _ = testing.T{}
