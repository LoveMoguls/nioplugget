package device

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trollstaven/nioplugget/backend/internal/auth"
	"github.com/trollstaven/nioplugget/backend/internal/child"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

type fakeStore struct {
	Store // panics on unimplemented
	settings *queries.FamilySetting
	upserted string
}

func (f *fakeStore) GetFamilySettings(_ context.Context) (queries.FamilySetting, error) {
	if f.settings == nil {
		return queries.FamilySetting{}, errors.New("no rows")
	}
	return *f.settings, nil
}

func (f *fakeStore) UpsertFamilyCode(_ context.Context, codeHash string) (queries.FamilySetting, error) {
	f.upserted = codeHash
	epoch := int32(1)
	if f.settings != nil {
		epoch = f.settings.DeviceEpoch + 1
	}
	fs := queries.FamilySetting{ID: 1, CodeHash: codeHash, DeviceEpoch: epoch}
	f.settings = &fs
	return fs, nil
}

func settingsWithCode(t *testing.T, code string, epoch int32) *queries.FamilySetting {
	t.Helper()
	hash, err := auth.HashPassword(code)
	if err != nil {
		t.Fatal(err)
	}
	return &queries.FamilySetting{ID: 1, CodeHash: hash, DeviceEpoch: epoch}
}

func newTestHandler(store Store) *Handler {
	auth.NewTokenAuth("test-secret")
	return NewHandler(store, child.NewPINRateLimiter(5, 15*time.Minute))
}

func TestUnlockCorrectCodeSetsDeviceCookie(t *testing.T) {
	h := newTestHandler(&fakeStore{settings: settingsWithCode(t, "hemlig1", 2)})
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/device/unlock", strings.NewReader(`{"code":"hemlig1"}`))
	h.Unlock(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}
	var deviceCookie *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == "device" {
			deviceCookie = c
		}
	}
	if deviceCookie == nil || deviceCookie.Value == "" || !deviceCookie.HttpOnly {
		t.Fatalf("expected HttpOnly device cookie, got %+v", deviceCookie)
	}
}

func TestUnlockWrongCodeRejects(t *testing.T) {
	h := newTestHandler(&fakeStore{settings: settingsWithCode(t, "hemlig1", 1)})
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/device/unlock", strings.NewReader(`{"code":"fel"}`))
	h.Unlock(rr, r)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Fel familjekod") {
		t.Errorf("body %s", rr.Body.String())
	}
}

func TestUnlockNoCodeSet(t *testing.T) {
	h := newTestHandler(&fakeStore{})
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/device/unlock", strings.NewReader(`{"code":"nagot"}`))
	h.Unlock(rr, r)
	if rr.Code != http.StatusUnauthorized || !strings.Contains(rr.Body.String(), "Ingen familjekod") {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
}

func TestUnlockRateLimited(t *testing.T) {
	h := newTestHandler(&fakeStore{settings: settingsWithCode(t, "hemlig1", 1)})
	var last *httptest.ResponseRecorder
	for i := 0; i < 6; i++ {
		last = httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/api/device/unlock", strings.NewReader(`{"code":"fel"}`))
		r.RemoteAddr = "10.0.0.9:1234"
		h.Unlock(last, r)
	}
	if last.Code != http.StatusTooManyRequests {
		t.Fatalf("status %d, want 429", last.Code)
	}
}

var _ = pgtype.UUID{} // keep import if unused in this file after edits
