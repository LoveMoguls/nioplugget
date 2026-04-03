package auth

import (
	"errors"

	"github.com/alexedwards/argon2id"
)

// HashPassword hashes the given password using Argon2id.
// Returns an error if the password is empty.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password must not be empty")
	}
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

// ComparePassword compares a plaintext password against an Argon2id hash.
// Returns true if they match, false otherwise.
func ComparePassword(hash, password string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hash)
}
