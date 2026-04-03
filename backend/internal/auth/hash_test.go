package auth_test

import (
	"testing"

	"github.com/trollstaven/nioplugget/backend/internal/auth"
)

func TestHashPassword_ReturnsHash(t *testing.T) {
	hash, err := auth.HashPassword("mypassword")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hash == "" {
		t.Error("expected non-empty hash")
	}
}

func TestHashPassword_EmptyPasswordReturnsError(t *testing.T) {
	_, err := auth.HashPassword("")
	if err == nil {
		t.Error("expected error for empty password, got nil")
	}
}

func TestComparePassword_CorrectPassword(t *testing.T) {
	hash, err := auth.HashPassword("mypassword")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	match, err := auth.ComparePassword(hash, "mypassword")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !match {
		t.Error("expected match for correct password")
	}
}

func TestComparePassword_WrongPassword(t *testing.T) {
	hash, err := auth.HashPassword("mypassword")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	match, err := auth.ComparePassword(hash, "wrongpassword")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if match {
		t.Error("expected no match for wrong password")
	}
}

func TestHashPassword_UniqueSalts(t *testing.T) {
	hash1, err := auth.HashPassword("mypassword")
	if err != nil {
		t.Fatalf("failed to hash password (1): %v", err)
	}
	hash2, err := auth.HashPassword("mypassword")
	if err != nil {
		t.Fatalf("failed to hash password (2): %v", err)
	}
	if hash1 == hash2 {
		t.Error("expected different hashes for same password (unique salts)")
	}
}
