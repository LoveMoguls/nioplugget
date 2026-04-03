package auth

import (
	"time"

	"github.com/go-chi/jwtauth/v5"
)

// TokenAuth is the global JWT authenticator initialized by NewTokenAuth.
var TokenAuth *jwtauth.JWTAuth

// NewTokenAuth initializes TokenAuth with HS256 using the given secret.
// Returns the created JWTAuth for convenience (also sets the package-level variable).
func NewTokenAuth(secret string) *jwtauth.JWTAuth {
	TokenAuth = jwtauth.New("HS256", []byte(secret), nil)
	return TokenAuth
}

// GenerateToken creates a JWT with sub, role, and exp claims.
// Expiry is set to 24 hours from now.
// Returns the token string, the expiry time, and any error.
func GenerateToken(userID string, role string) (string, time.Time, error) {
	expiry := time.Now().Add(24 * time.Hour)
	claims := map[string]any{
		"sub":  userID,
		"role": role,
		"exp":  expiry.Unix(),
	}
	_, tokenStr, err := TokenAuth.Encode(claims)
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenStr, expiry, nil
}
