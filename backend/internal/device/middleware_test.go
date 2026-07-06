package device

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/trollstaven/nioplugget/backend/internal/auth"
)

func passThroughHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestEpochGuard_NoCookie_PassesThrough(t *testing.T) {
	h := newTestHandler(&fakeStore{settings: settingsWithCode(t, "hemlig1", 3)})
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/subjects", nil)
	h.EpochGuard(passThroughHandler()).ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d, want 200 (pass through when no jwt cookie)", rr.Code)
	}
}

func TestEpochGuard_EpochlessToken_PassesThrough(t *testing.T) {
	h := newTestHandler(&fakeStore{settings: settingsWithCode(t, "hemlig1", 3)})
	// Bootstrap login (password/PIN) — no epoch claim.
	tokenStr, _, err := auth.GenerateToken("parent-id", "parent")
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/subjects", nil)
	r.AddCookie(&http.Cookie{Name: "jwt", Value: tokenStr})
	h.EpochGuard(passThroughHandler()).ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d, want 200 (pass through for epoch-less/bootstrap token)", rr.Code)
	}
}

func TestEpochGuard_StaleEpoch_ClearsAndRejects(t *testing.T) {
	h := newTestHandler(&fakeStore{settings: settingsWithCode(t, "hemlig1", 5)})
	tokenStr, _, err := generateProfileToken("child-id", "child", 4)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/subjects", nil)
	r.AddCookie(&http.Cookie{Name: "jwt", Value: tokenStr})
	h.EpochGuard(passThroughHandler()).ServeHTTP(rr, r)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status %d, want 401 on stale epoch", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Utloggad") {
		t.Errorf("body %s", rr.Body.String())
	}
	var cleared *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == "jwt" {
			cleared = c
		}
	}
	if cleared == nil || cleared.MaxAge >= 0 {
		t.Fatalf("expected jwt cookie to be cleared, got %+v", cleared)
	}
}

func TestEpochGuard_CurrentEpoch_PassesThrough(t *testing.T) {
	h := newTestHandler(&fakeStore{settings: settingsWithCode(t, "hemlig1", 5)})
	tokenStr, _, err := generateProfileToken("child-id", "child", 5)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/subjects", nil)
	r.AddCookie(&http.Cookie{Name: "jwt", Value: tokenStr})
	h.EpochGuard(passThroughHandler()).ServeHTTP(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d, want 200 for current epoch", rr.Code)
	}
}
