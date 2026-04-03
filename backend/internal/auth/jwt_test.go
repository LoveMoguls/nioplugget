package auth_test

import (
	"testing"
	"time"

	"github.com/trollstaven/nioplugget/backend/internal/auth"
)

func TestJWT_GenerateToken_ParentRole(t *testing.T) {
	auth.NewTokenAuth("test-secret")
	tokenStr, expiry, err := auth.GenerateToken("user-123", "parent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tokenStr == "" {
		t.Error("expected non-empty token string")
	}
	if expiry.IsZero() {
		t.Error("expected non-zero expiry time")
	}
}

func TestJWT_GenerateToken_ChildRole(t *testing.T) {
	auth.NewTokenAuth("test-secret")
	tokenStr, _, err := auth.GenerateToken("child-456", "child")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tokenStr == "" {
		t.Error("expected non-empty token string")
	}
}

func TestJWT_TokenContainsClaims(t *testing.T) {
	auth.NewTokenAuth("test-secret")
	tokenStr, expiry, err := auth.GenerateToken("user-123", "parent")
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	// Decode and verify claims
	token, err := auth.TokenAuth.Decode(tokenStr)
	if err != nil {
		t.Fatalf("failed to decode token: %v", err)
	}

	// Check sub claim
	sub, ok := token.Subject()
	if !ok || sub == "" {
		t.Error("expected non-empty sub claim")
	}
	if sub != "user-123" {
		t.Errorf("expected sub=user-123, got %s", sub)
	}

	// Check role claim
	var role string
	if err := token.Get("role", &role); err != nil {
		t.Errorf("expected role claim, got error: %v", err)
	}
	if role != "parent" {
		t.Errorf("expected role=parent, got %s", role)
	}

	// Check expiry
	exp, expOk := token.Expiration()
	if !expOk || exp.IsZero() {
		t.Error("expected non-zero exp claim")
	}
	// Expiry should be approximately 24 hours from now
	diff := expiry.Sub(time.Now())
	if diff < 23*time.Hour || diff > 25*time.Hour {
		t.Errorf("expected expiry ~24h from now, got %v", diff)
	}
}

func TestJWT_TokenAuth_IsInitialized(t *testing.T) {
	auth.NewTokenAuth("test-secret")
	if auth.TokenAuth == nil {
		t.Error("expected TokenAuth to be initialized")
	}
}

func TestJWT_NewTokenAuth_NotNil(t *testing.T) {
	ta := auth.NewTokenAuth("another-secret")
	if ta == nil {
		t.Error("expected non-nil JWTAuth")
	}
}

func TestJWT_ChildTokenHasChildRole(t *testing.T) {
	auth.NewTokenAuth("test-secret")
	tokenStr, _, err := auth.GenerateToken("child-456", "child")
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	token, err := auth.TokenAuth.Decode(tokenStr)
	if err != nil {
		t.Fatalf("failed to decode token: %v", err)
	}

	var role string
	if err := token.Get("role", &role); err != nil {
		t.Errorf("expected role claim, got error: %v", err)
	}
	if role != "child" {
		t.Errorf("expected role=child, got %s", role)
	}
}

