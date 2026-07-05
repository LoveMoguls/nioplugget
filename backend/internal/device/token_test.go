package device

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/trollstaven/nioplugget/backend/internal/auth"
)

func TestGenerateDeviceTokenRoundTrip(t *testing.T) {
	auth.NewTokenAuth("test-secret")
	tokenStr, expiry, err := GenerateDeviceToken(3)
	if err != nil {
		t.Fatal(err)
	}
	if until := time.Until(expiry); until < 364*24*time.Hour || until > 366*24*time.Hour {
		t.Errorf("expiry %v not ~365d away", expiry)
	}
	tok, err := auth.TokenAuth.Decode(tokenStr)
	if err != nil {
		t.Fatal(err)
	}
	var role string
	if err := tok.Get("role", &role); err != nil {
		t.Fatalf("get role: %v", err)
	}
	if role != "device" {
		t.Errorf("role = %v, want device", role)
	}
	var epoch float64
	if err := tok.Get("epoch", &epoch); err != nil {
		t.Fatalf("get epoch: %v", err)
	}
	if int32(epoch) != 3 {
		t.Errorf("epoch = %v, want 3", epoch)
	}
}

func TestClientIP(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	r.RemoteAddr = "192.168.1.10:5555"
	if got := clientIP(r); got != "192.168.1.10" {
		t.Errorf("got %q", got)
	}
	r.Header.Set("CF-Connecting-IP", "203.0.113.7")
	if got := clientIP(r); got != "203.0.113.7" {
		t.Errorf("got %q", got)
	}
}
