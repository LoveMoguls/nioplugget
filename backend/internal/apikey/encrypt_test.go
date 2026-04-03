package apikey_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/trollstaven/nioplugget/backend/internal/apikey"
)

func TestNewEncryptionService_InvalidKey(t *testing.T) {
	// Too short key (not 32 bytes)
	shortKey := hex.EncodeToString([]byte("tooshort"))
	_, err := apikey.NewEncryptionService(shortKey)
	if err == nil {
		t.Fatal("expected error for key != 32 bytes, got nil")
	}
}

func TestNewEncryptionService_ValidKey(t *testing.T) {
	key := make([]byte, 32)
	hexKey := hex.EncodeToString(key)
	svc, err := apikey.NewEncryptionService(hexKey)
	if err != nil {
		t.Fatalf("expected no error for valid 32-byte key, got: %v", err)
	}
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestEncrypt_DiffersFromPlaintext(t *testing.T) {
	svc := newTestEncryptionService(t)
	plaintext := []byte("my-secret-api-key")

	ciphertext, err := svc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}
	if bytes.Equal(ciphertext, plaintext) {
		t.Fatal("ciphertext should differ from plaintext")
	}
}

func TestDecrypt_ReturnsOriginalPlaintext(t *testing.T) {
	svc := newTestEncryptionService(t)
	plaintext := []byte("my-secret-api-key")

	ciphertext, err := svc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt returned error: %v", err)
	}

	decrypted, err := svc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt returned error: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("decrypted %q does not match original %q", decrypted, plaintext)
	}
}

func TestEncrypt_UniqueNonces(t *testing.T) {
	svc := newTestEncryptionService(t)
	plaintext := []byte("same-plaintext")

	ct1, err := svc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("first Encrypt error: %v", err)
	}
	ct2, err := svc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("second Encrypt error: %v", err)
	}

	if bytes.Equal(ct1, ct2) {
		t.Fatal("expected different ciphertexts for same plaintext (unique nonces)")
	}
}

func TestDecrypt_WrongKey(t *testing.T) {
	svc1 := newTestEncryptionService(t)

	key2 := make([]byte, 32)
	key2[0] = 0xFF
	svc2, err := apikey.NewEncryptionService(hex.EncodeToString(key2))
	if err != nil {
		t.Fatalf("NewEncryptionService error: %v", err)
	}

	ciphertext, err := svc1.Encrypt([]byte("secret"))
	if err != nil {
		t.Fatalf("Encrypt error: %v", err)
	}

	_, err = svc2.Decrypt(ciphertext)
	if err == nil {
		t.Fatal("expected error when decrypting with wrong key")
	}
}

func TestDecrypt_CorruptedCiphertext(t *testing.T) {
	svc := newTestEncryptionService(t)

	ciphertext, err := svc.Encrypt([]byte("secret"))
	if err != nil {
		t.Fatalf("Encrypt error: %v", err)
	}

	// Corrupt the ciphertext
	corrupted := make([]byte, len(ciphertext))
	copy(corrupted, ciphertext)
	if len(corrupted) > 12 {
		corrupted[12] ^= 0xFF
	}

	_, err = svc.Decrypt(corrupted)
	if err == nil {
		t.Fatal("expected error when decrypting corrupted ciphertext")
	}
}

// helpers

func newTestEncryptionService(t *testing.T) *apikey.EncryptionService {
	t.Helper()
	key := make([]byte, 32)
	svc, err := apikey.NewEncryptionService(hex.EncodeToString(key))
	if err != nil {
		t.Fatalf("failed to create test encryption service: %v", err)
	}
	return svc
}
