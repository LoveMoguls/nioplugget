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
	parents  []queries.ListParentsRow
	students []queries.ListAllStudentsRow
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

func (f *fakeStore) ListParents(_ context.Context) ([]queries.ListParentsRow, error) {
	return f.parents, nil
}
func (f *fakeStore) ListAllStudents(_ context.Context) ([]queries.ListAllStudentsRow, error) {
	return f.students, nil
}
func (f *fakeStore) GetParentByID(_ context.Context, id pgtype.UUID) (queries.Parent, error) {
	for _, p := range f.parents {
		if p.ID == id {
			return queries.Parent{ID: p.ID, Email: p.Email}, nil
		}
	}
	return queries.Parent{}, errors.New("not found")
}
func (f *fakeStore) GetStudentByID(_ context.Context, id pgtype.UUID) (queries.Student, error) {
	for _, s := range f.students {
		if s.ID == id {
			return queries.Student{ID: s.ID, Name: s.Name}, nil
		}
	}
	return queries.Student{}, errors.New("not found")
}

func deviceRequest(t *testing.T, method, path, body string, epoch int32) *http.Request {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	tokenStr, _, err := GenerateDeviceToken(epoch)
	if err != nil {
		t.Fatal(err)
	}
	r.AddCookie(&http.Cookie{Name: "device", Value: tokenStr})
	return r
}

func testUUID(t *testing.T, s string) pgtype.UUID {
	t.Helper()
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		t.Fatal(err)
	}
	return u
}

func TestProfilesRequiresDeviceCookie(t *testing.T) {
	h := newTestHandler(&fakeStore{settings: settingsWithCode(t, "hemlig1", 1)})
	rr := httptest.NewRecorder()
	h.Profiles(rr, httptest.NewRequest(http.MethodGet, "/api/profiles", nil))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status %d, want 401", rr.Code)
	}
}

func TestProfilesRejectsStaleEpoch(t *testing.T) {
	h := newTestHandler(&fakeStore{settings: settingsWithCode(t, "hemlig1", 5)})
	rr := httptest.NewRecorder()
	h.Profiles(rr, deviceRequest(t, http.MethodGet, "/api/profiles", "", 4))
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("status %d, want 401 on stale epoch", rr.Code)
	}
}

func TestProfilesListsFamily(t *testing.T) {
	store := &fakeStore{
		settings: settingsWithCode(t, "hemlig1", 1),
		parents:  []queries.ListParentsRow{{ID: testUUID(t, "11111111-1111-1111-1111-111111111111"), Email: "pappa@example.com"}},
		students: []queries.ListAllStudentsRow{{ID: testUUID(t, "22222222-2222-2222-2222-222222222222"), Name: "Nio"}},
	}
	h := newTestHandler(store)
	rr := httptest.NewRecorder()
	h.Profiles(rr, deviceRequest(t, http.MethodGet, "/api/profiles", "", 1))
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}
	body := rr.Body.String()
	if !strings.Contains(body, `"name":"pappa"`) || !strings.Contains(body, `"name":"Nio"`) {
		t.Errorf("body %s", body)
	}
}

func TestProfileLoginSetsJWT(t *testing.T) {
	childID := testUUID(t, "22222222-2222-2222-2222-222222222222")
	store := &fakeStore{
		settings: settingsWithCode(t, "hemlig1", 1),
		students: []queries.ListAllStudentsRow{{ID: childID, Name: "Nio"}},
	}
	h := newTestHandler(store)
	rr := httptest.NewRecorder()
	h.ProfileLogin(rr, deviceRequest(t, http.MethodPost, "/api/profile/login",
		`{"id":"22222222-2222-2222-2222-222222222222","role":"child"}`, 1))
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}
	var jwtCookie *http.Cookie
	for _, c := range rr.Result().Cookies() {
		if c.Name == "jwt" {
			jwtCookie = c
		}
	}
	if jwtCookie == nil || jwtCookie.Value == "" {
		t.Fatal("expected jwt cookie")
	}
	tok, err := auth.TokenAuth.Decode(jwtCookie.Value)
	if err != nil {
		t.Fatal(err)
	}
	var role string
	if err := tok.Get("role", &role); err != nil {
		t.Fatalf("get role: %v", err)
	}
	if role != "child" {
		t.Errorf("role %v", role)
	}
}

func TestProfileLoginUnknownID(t *testing.T) {
	h := newTestHandler(&fakeStore{settings: settingsWithCode(t, "hemlig1", 1)})
	rr := httptest.NewRecorder()
	h.ProfileLogin(rr, deviceRequest(t, http.MethodPost, "/api/profile/login",
		`{"id":"99999999-9999-9999-9999-999999999999","role":"child"}`, 1))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status %d, want 404", rr.Code)
	}
}

func TestSetCodeFirstTime(t *testing.T) {
	store := &fakeStore{}
	h := newTestHandler(store)
	rr := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/api/device/set-code", strings.NewReader(`{"newCode":"hemlig1"}`))
	h.SetCode(rr, r)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}
	if store.upserted == "" {
		t.Fatal("expected code hash upserted")
	}
}

func TestSetCodeChangeRequiresCurrent(t *testing.T) {
	store := &fakeStore{settings: settingsWithCode(t, "gammal1", 1)}
	h := newTestHandler(store)

	rr := httptest.NewRecorder()
	h.SetCode(rr, httptest.NewRequest(http.MethodPost, "/api/device/set-code",
		strings.NewReader(`{"newCode":"nyhemlig"}`)))
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status %d, want 403 without currentCode", rr.Code)
	}

	rr = httptest.NewRecorder()
	h.SetCode(rr, httptest.NewRequest(http.MethodPost, "/api/device/set-code",
		strings.NewReader(`{"newCode":"nyhemlig","currentCode":"gammal1"}`)))
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}
	if store.settings.DeviceEpoch != 2 {
		t.Errorf("epoch %d, want 2 (bumped)", store.settings.DeviceEpoch)
	}
}

func TestSetCodeTooShort(t *testing.T) {
	h := newTestHandler(&fakeStore{})
	rr := httptest.NewRecorder()
	h.SetCode(rr, httptest.NewRequest(http.MethodPost, "/api/device/set-code",
		strings.NewReader(`{"newCode":"kort"}`)))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d, want 400", rr.Code)
	}
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
