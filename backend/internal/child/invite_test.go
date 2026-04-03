package child

import (
	"regexp"
	"testing"
)

func TestInviteTokenLength(t *testing.T) {
	token, err := GenerateInviteToken()
	if err != nil {
		t.Fatalf("GenerateInviteToken() error = %v", err)
	}
	if len(token) < 43 {
		t.Errorf("token length = %d, want >= 43", len(token))
	}
}

func TestInviteTokenUniqueness(t *testing.T) {
	token1, err := GenerateInviteToken()
	if err != nil {
		t.Fatalf("first GenerateInviteToken() error = %v", err)
	}
	token2, err := GenerateInviteToken()
	if err != nil {
		t.Fatalf("second GenerateInviteToken() error = %v", err)
	}
	if token1 == token2 {
		t.Error("two calls produced identical tokens, expected unique")
	}
}

func TestInviteTokenURLSafe(t *testing.T) {
	token, err := GenerateInviteToken()
	if err != nil {
		t.Fatalf("GenerateInviteToken() error = %v", err)
	}
	// base64url characters: A-Z a-z 0-9 - _  (with optional = padding)
	re := regexp.MustCompile(`^[A-Za-z0-9\-_=]+$`)
	if !re.MatchString(token) {
		t.Errorf("token %q contains non-URL-safe characters", token)
	}
}
