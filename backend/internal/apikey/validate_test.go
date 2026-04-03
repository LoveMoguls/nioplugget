package apikey_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/trollstaven/nioplugget/backend/internal/apikey"
)

func TestValidateAPIKey_EmptyKey(t *testing.T) {
	err := apikey.ValidateAPIKey("")
	if err == nil {
		t.Fatal("expected error for empty key, got nil")
	}
}

func TestValidateAPIKey_ValidKey(t *testing.T) {
	// Mock server that returns 200 for valid key
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") == "" {
			http.Error(w, "missing key", http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	err := apikey.ValidateAPIKeyWithURL("sk-ant-valid-key", srv.URL+"/v1/models")
	if err != nil {
		t.Fatalf("expected no error for valid key response, got: %v", err)
	}
}

func TestValidateAPIKey_InvalidKey(t *testing.T) {
	// Mock server that returns 401 for invalid key
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer srv.Close()

	err := apikey.ValidateAPIKeyWithURL("sk-ant-invalid-key", srv.URL+"/v1/models")
	if err == nil {
		t.Fatal("expected error for 401 response, got nil")
	}
}

func TestValidateAPIKey_ServerError(t *testing.T) {
	// Mock server that returns 500
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	err := apikey.ValidateAPIKeyWithURL("sk-ant-key", srv.URL+"/v1/models")
	if err == nil {
		t.Fatal("expected error for server error response, got nil")
	}
}

func TestValidateAPIKey_ConnectionError(t *testing.T) {
	// Use a URL that will fail to connect
	err := apikey.ValidateAPIKeyWithURL("sk-ant-key", "http://127.0.0.1:1/v1/models")
	if err == nil {
		t.Fatal("expected error for connection failure, got nil")
	}
}

func TestMaskAPIKey_Normal(t *testing.T) {
	key := "sk-ant-api03-abcdefghijklmnop"
	masked := apikey.MaskAPIKey(key)
	expected := "sk-ant...****"
	if masked != expected {
		t.Fatalf("expected %q, got %q", expected, masked)
	}
}

func TestMaskAPIKey_Short(t *testing.T) {
	key := "abc"
	masked := apikey.MaskAPIKey(key)
	if masked != "****" {
		t.Fatalf("expected %q for short key, got %q", "****", masked)
	}
}

func TestMaskAPIKey_ExactlyEightChars(t *testing.T) {
	key := "sk-ant-x"
	masked := apikey.MaskAPIKey(key)
	// len >= 8, should return first 6 + "...****"
	expected := "sk-ant...****"
	if masked != expected {
		t.Fatalf("expected %q, got %q", expected, masked)
	}
}
