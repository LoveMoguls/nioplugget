# Family Device Auth (profilväljare) Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Engångsupplåsning per enhet med familjekod → Netflix-lik profilväljare → tryck på avatar = inloggad; lösenord/PIN försvinner ur vardagen. Spec: `docs/superpowers/specs/2026-07-05-family-device-auth-design.md`.

**Architecture:** Nytt paket `backend/internal/device` (unlock/profiles/profile-login/set-code) + tabell `family_settings` (Argon2id-hash av familjekod + `device_epoch`). Device-cookie = JWT `{role:"device", epoch:N}`, 365 d, signerad med samma `JWT_SECRET`. Ordinarie `jwt`-cookie förlängs 24 h → 30 d. Frontend: `/las-upp`, `/profiler`, rot-redirect, "Byt profil", Familjekod-kort i föräldrapanelen; PIN/invite bort ur UI (backend orörd).

**Tech Stack:** Go 1.25, chi + jwtauth/v5, argon2id (`internal/auth`), sqlc/pgx, golang-migrate, SvelteKit (Svelte 5 runes, shadcn-svelte).

## Global Constraints

- Modulnamn: `github.com/trollstaven/nioplugget/backend`.
- All användartext svenska; fel loggas med zerolog; API-fel som `{"error":"..."}`.
- Rollen `device` får ALDRIG passera `ParentOnly`/`ChildOnly` (de kräver exakt "parent"/"child" — verifiera, ändra inte).
- Ingen familjekod satt (tom `family_settings`) ⇒ unlock svarar 401 med `{"error":"Ingen familjekod är satt ännu"}`; apikey-endpoints beter sig som idag.
- Byte av familjekod bumpar `device_epoch` ⇒ gamla device-cookies avvisas.
- Rate limit på unlock: 5 försök / 15 min per klient-IP (CF-Connecting-IP-headern först, annars RemoteAddr-host).
- Kommandon körs från `backend/` resp. `frontend/`. Migrate: `set -a && source .env && set +a && migrate -path db/migrations -database "$DATABASE_URL" up` (DB-containern `nioplugget-db` är igång).
- Servern kör produktionsdrift — deploya INTE (starta inte om tjänster); det sker i sista tasken av controllern.
- Commit efter varje grönt steg, trailer: `Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>`.

---

### Task 1: Migration 013 + queries + sqlc

**Files:**
- Create: `backend/db/migrations/013_family_settings.up.sql`
- Create: `backend/db/migrations/013_family_settings.down.sql`
- Create: `backend/db/queries/family.sql`
- Generated: `backend/internal/database/queries/family.sql.go`

**Interfaces:**
- Produces (Task 3–5 använder): `GetFamilySettings(ctx) (FamilySettings, error)`, `UpsertFamilyCode(ctx, codeHash string) (FamilySettings, error)`, `ListParents(ctx) ([]ListParentsRow, error)` (ID, Email), `ListAllStudents(ctx) ([]ListAllStudentsRow, error)` (ID, Name). `FamilySettings{ID int16, CodeHash string, DeviceEpoch int32, UpdatedAt pgtype.Timestamptz}`.

- [ ] **Step 1: up-migration**

`backend/db/migrations/013_family_settings.up.sql`:

```sql
CREATE TABLE family_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    code_hash TEXT NOT NULL,
    device_epoch INT NOT NULL DEFAULT 1,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

- [ ] **Step 2: down-migration**

`backend/db/migrations/013_family_settings.down.sql`:

```sql
DROP TABLE family_settings;
```

- [ ] **Step 3: queries**

`backend/db/queries/family.sql`:

```sql
-- name: GetFamilySettings :one
SELECT * FROM family_settings WHERE id = 1;

-- name: UpsertFamilyCode :one
INSERT INTO family_settings (id, code_hash)
VALUES (1, $1)
ON CONFLICT (id) DO UPDATE SET
    code_hash = $1,
    device_epoch = family_settings.device_epoch + 1,
    updated_at = NOW()
RETURNING *;

-- name: ListParents :many
SELECT id, email FROM parents ORDER BY created_at;

-- name: ListAllStudents :many
SELECT id, name FROM students ORDER BY created_at;
```

- [ ] **Step 4: migrera + generera + bygg**

```bash
set -a && source .env && set +a && migrate -path db/migrations -database "$DATABASE_URL" up
sqlc generate
go build ./...
```
Expected: migration OK, `family.sql.go` skapad, bygge grönt. Verifiera de genererade signaturerna mot Interfaces-blocket ovan (sqlc kan namnge param annorlunda — notera avvikelser i rapporten).

- [ ] **Step 5: Commit** — `git add db/ internal/database/queries/ && git commit -m "feat(device): family_settings schema and queries"`

---

### Task 2: 30-dagars JWT + logout för båda roller

**Files:**
- Modify: `backend/internal/auth/jwt.go:23` (24h → 30d)
- Modify: `backend/cmd/server/main.go` (flytta logout ur ParentOnly-gruppen)
- Test: `backend/internal/auth/jwt_test.go`

**Interfaces:**
- Produces: `auth.GenerateToken(userID, role string) (string, time.Time, error)` — oförändrad signatur, expiry nu 30 dagar. `POST /api/auth/logout` nåbar för både parent och child (fortfarande bakom Verifier+Authenticator).

- [ ] **Step 1: Failande test**

Lägg i `backend/internal/auth/jwt_test.go` (behåll befintliga tester; justera ev. test som antar 24 h):

```go
func TestGenerateTokenExpiry30Days(t *testing.T) {
	NewTokenAuth("test-secret")
	_, expiry, err := GenerateToken("user-1", "parent")
	if err != nil {
		t.Fatal(err)
	}
	want := time.Now().Add(30 * 24 * time.Hour)
	if diff := expiry.Sub(want); diff < -time.Minute || diff > time.Minute {
		t.Errorf("expiry %v, want ~%v", expiry, want)
	}
}
```

- [ ] **Step 2: RED** — `go test ./internal/auth/ -run TestGenerateTokenExpiry30Days -v` → FAIL (expiry ~24 h)

- [ ] **Step 3: Implementera** — i `jwt.go`: `expiry := time.Now().Add(30 * 24 * time.Hour)` och uppdatera kommentaren ("Expiry is set to 30 days from now.").

- [ ] **Step 4: Flytta logout.** I `main.go`: `r.Post("/logout", authHandler.Logout)` ligger idag i den ParentOnly-nästlade gruppen under `/api/auth` (rad ~121-124). Flytta upp den till den yttre skyddade gruppen (samma nivå som `r.Get("/me", ...)`), så child-JWT också kan logga ut. Ta bort den nu tomma ParentOnly-gruppen om inget annat finns kvar i den.

- [ ] **Step 5: GREEN + hela sviten** — `go test ./... && go vet ./...` (justera ev. befintligt test som hårdkodat 24 h). Expected: allt grönt.

- [ ] **Step 6: Commit** — `git commit -m "feat(auth): 30-day sessions, logout for both roles"`

---

### Task 3: device-paketet — token, klient-IP, unlock med rate limit

**Files:**
- Create: `backend/internal/device/token.go`
- Create: `backend/internal/device/handler.go`
- Create: `backend/internal/device/store.go`
- Test: `backend/internal/device/handler_test.go`, `backend/internal/device/token_test.go`

**Interfaces:**
- Consumes: `auth.TokenAuth` (jwtauth), `auth.HashPassword`/`auth.ComparePassword`, `child.NewPINRateLimiter(5, 15*time.Minute)` (generisk försöksräknare, nyckel = IP-sträng; `Check(key) (bool, int)`, `Reset(key)`).
- Produces (Task 4/6): `Store` interface; `NewHandler(store Store, limiter *child.PINRateLimiter) *Handler`; `(*Handler).Unlock(w, r)`; `GenerateDeviceToken(epoch int32) (string, time.Time, error)`; `deviceClaims(r *http.Request) (epoch int32, ok bool)`; `clientIP(r *http.Request) string`; cookie-namn-konstant `deviceCookieName = "device"`; `TokenFromDeviceCookie(r *http.Request) string` (för jwtauth.Verify i Task 6).

- [ ] **Step 1: store.go**

```go
package device

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/trollstaven/nioplugget/backend/internal/database/queries"
)

// Store is the database access the device/profile endpoints need.
type Store interface {
	GetFamilySettings(ctx context.Context) (queries.FamilySettings, error)
	UpsertFamilyCode(ctx context.Context, codeHash string) (queries.FamilySettings, error)
	ListParents(ctx context.Context) ([]queries.ListParentsRow, error)
	ListAllStudents(ctx context.Context) ([]queries.ListAllStudentsRow, error)
	GetParentByID(ctx context.Context, id pgtype.UUID) (queries.Parent, error)
	GetStudentByID(ctx context.Context, id pgtype.UUID) (queries.Student, error)
}

// QueriesStore implements Store with sqlc-generated queries.
type QueriesStore struct{ q *queries.Queries }

func NewQueriesStore(q *queries.Queries) *QueriesStore { return &QueriesStore{q: q} }

func (s *QueriesStore) GetFamilySettings(ctx context.Context) (queries.FamilySettings, error) {
	return s.q.GetFamilySettings(ctx)
}
func (s *QueriesStore) UpsertFamilyCode(ctx context.Context, codeHash string) (queries.FamilySettings, error) {
	return s.q.UpsertFamilyCode(ctx, codeHash)
}
func (s *QueriesStore) ListParents(ctx context.Context) ([]queries.ListParentsRow, error) {
	return s.q.ListParents(ctx)
}
func (s *QueriesStore) ListAllStudents(ctx context.Context) ([]queries.ListAllStudentsRow, error) {
	return s.q.ListAllStudents(ctx)
}
func (s *QueriesStore) GetParentByID(ctx context.Context, id pgtype.UUID) (queries.Parent, error) {
	return s.q.GetParentByID(ctx, id)
}
func (s *QueriesStore) GetStudentByID(ctx context.Context, id pgtype.UUID) (queries.Student, error) {
	return s.q.GetStudentByID(ctx, id)
}
```

(Verifiera genererade metodnamn/typer mot `internal/database/queries/` — sqlc:s UpsertFamilyCode-param kan heta `codeHash` eller vara positionsbaserad; anpassa wrappern, inte interfacet.)

- [ ] **Step 2: Failande tester för token + clientIP**

`backend/internal/device/token_test.go`:

```go
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
	claims, _ := tok.AsMap(nil)
	if claims["role"] != "device" {
		t.Errorf("role = %v, want device", claims["role"])
	}
	if epoch, _ := claims["epoch"].(float64); int32(epoch) != 3 {
		t.Errorf("epoch = %v, want 3", claims["epoch"])
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
```

(OBS: `AsMap`-signaturen beror på jwx-versionen; om `tok.AsMap(nil)` inte kompilerar, använd `tok.AsMap(context.Background())`. Epoch-claimen decodas som `float64` via JSON-numbers — behåll konverteringen.)

- [ ] **Step 3: RED** — `go test ./internal/device/ -v` → kompileringsfel.

- [ ] **Step 4: token.go**

```go
package device

import (
	"net"
	"net/http"
	"time"

	"github.com/trollstaven/nioplugget/backend/internal/auth"
)

const deviceCookieName = "device"

// GenerateDeviceToken creates a 365-day JWT with role "device" and the
// current device epoch. Signed with the same secret as session JWTs.
func GenerateDeviceToken(epoch int32) (string, time.Time, error) {
	expiry := time.Now().Add(365 * 24 * time.Hour)
	claims := map[string]any{
		"role":  "device",
		"epoch": epoch,
		"exp":   expiry.Unix(),
	}
	_, tokenStr, err := auth.TokenAuth.Encode(claims)
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenStr, expiry, nil
}

// TokenFromDeviceCookie extracts the device JWT for jwtauth.Verify.
func TokenFromDeviceCookie(r *http.Request) string {
	c, err := r.Cookie(deviceCookieName)
	if err != nil {
		return ""
	}
	return c.Value
}

// clientIP prefers the Cloudflare header (all public traffic arrives via
// the tunnel from localhost), falling back to RemoteAddr.
func clientIP(r *http.Request) string {
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
```

- [ ] **Step 5: Failande tester för Unlock**

`backend/internal/device/handler_test.go`:

```go
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
	settings *queries.FamilySettings
	upserted string
}

func (f *fakeStore) GetFamilySettings(_ context.Context) (queries.FamilySettings, error) {
	if f.settings == nil {
		return queries.FamilySettings{}, errors.New("no rows")
	}
	return *f.settings, nil
}

func (f *fakeStore) UpsertFamilyCode(_ context.Context, codeHash string) (queries.FamilySettings, error) {
	f.upserted = codeHash
	epoch := int32(1)
	if f.settings != nil {
		epoch = f.settings.DeviceEpoch + 1
	}
	fs := queries.FamilySettings{ID: 1, CodeHash: codeHash, DeviceEpoch: epoch}
	f.settings = &fs
	return fs, nil
}

func settingsWithCode(t *testing.T, code string, epoch int32) *queries.FamilySettings {
	t.Helper()
	hash, err := auth.HashPassword(code)
	if err != nil {
		t.Fatal(err)
	}
	return &queries.FamilySettings{ID: 1, CodeHash: hash, DeviceEpoch: epoch}
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
```

- [ ] **Step 6: RED** — `go test ./internal/device/ -v` → `undefined: Handler` m.m.

- [ ] **Step 7: handler.go (Unlock + hjälpare)**

```go
package device

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/trollstaven/nioplugget/backend/internal/auth"
	"github.com/trollstaven/nioplugget/backend/internal/child"
)

// Handler serves device unlock and profile endpoints.
type Handler struct {
	store   Store
	limiter *child.PINRateLimiter
}

func NewHandler(store Store, limiter *child.PINRateLimiter) *Handler {
	return &Handler{store: store, limiter: limiter}
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Unlock handles POST /api/device/unlock — trades the family code for a
// 365-day trusted-device cookie.
func (h *Handler) Unlock(w http.ResponseWriter, r *http.Request) {
	ip := clientIP(r)
	if allowed, _ := h.limiter.Check(ip); !allowed {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "För många försök. Vänta 15 minuter."})
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Code == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Ogiltigt format"})
		return
	}

	settings, err := h.store.GetFamilySettings(r.Context())
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Ingen familjekod är satt ännu"})
		return
	}

	match, err := auth.ComparePassword(settings.CodeHash, req.Code)
	if err != nil || !match {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Fel familjekod"})
		return
	}
	h.limiter.Reset(ip)

	tokenStr, expiry, err := GenerateDeviceToken(settings.DeviceEpoch)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate device token")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     deviceCookieName,
		Value:    tokenStr,
		Path:     "/",
		Expires:  expiry,
		HttpOnly: true,
		Secure:   auth.IsSecureCookie(),
		SameSite: http.SameSiteLaxMode,
	})
	writeJSON(w, http.StatusOK, map[string]string{"message": "Enheten är upplåst"})
}
```

OBS: `auth.IsSecureCookie` — kolla `internal/auth/handler.go:67`: funktionen heter `isSecureCookie` (opriviligerad). Exportera den (`IsSecureCookie`, uppdatera interna anrop) som del av detta steg — den behövs även för device-cookien. Sök alla anrop: `grep -rn isSecureCookie internal/`.

- [ ] **Step 8: GREEN** — `go test ./internal/device/ -v && go test ./... && go vet ./...` → allt grönt.

- [ ] **Step 9: Commit** — `git commit -m "feat(device): unlock endpoint with family code and trusted-device cookie"`

---

### Task 4: profiles, profile-login, set-code

**Files:**
- Modify: `backend/internal/device/handler.go`
- Test: `backend/internal/device/handler_test.go` (utöka)

**Interfaces:**
- Consumes: `auth.GenerateToken`, `auth.SetAuthCookie`, `auth.GetUserIDFromContext`, `auth.HashPassword`.
- Produces (Task 6 router): `(*Handler).Profiles(w, r)`, `(*Handler).ProfileLogin(w, r)`, `(*Handler).SetCode(w, r)`; hjälpare `(*Handler).requireDevice(w, r) (ok bool)`.
- API-kontrakt (Task 7 frontend): `GET /api/profiles` → `[{"id":"...","name":"...","role":"parent"|"child"}]` (parent-namn = e-postens lokaldel före `@`); `POST /api/profile/login {"id":"...","role":"parent"|"child"}` → 200 + jwt-cookie + `{"role":"...","name":"..."}`; `POST /api/device/set-code {"newCode":"...","currentCode":"..."}` (parent-JWT; currentCode krävs när kod finns; newCode minst 6 tecken) → `{"message":"Familjekoden är sparad"}`.

- [ ] **Step 1: Failande tester** (lägg till i handler_test.go; utöka fakeStore med profiler)

```go
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
```

(lägg fälten `parents []queries.ListParentsRow` och `students []queries.ListAllStudentsRow` i fakeStore-structen)

```go
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
	claims, _ := tok.AsMap(nil)
	if claims["role"] != "child" {
		t.Errorf("role %v", claims["role"])
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
```

SetCode-tester (SetCode monteras bakom parent-JWT-middleware i Task 6; handlern läser bara `auth.GetUserIDFromContext` för logging — testa själva logiken direkt):

```go
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
```

- [ ] **Step 2: RED** — `go test ./internal/device/ -v` → undefined Profiles/ProfileLogin/SetCode.

- [ ] **Step 3: Implementera i handler.go**

```go
// requireDevice validates the device cookie: parses the JWT, checks role
// "device" and that the epoch matches the current family settings.
func (h *Handler) requireDevice(w http.ResponseWriter, r *http.Request) bool {
	tokenStr := TokenFromDeviceCookie(r)
	if tokenStr == "" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Enheten är inte upplåst"})
		return false
	}
	tok, err := auth.TokenAuth.Decode(tokenStr)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Enheten är inte upplåst"})
		return false
	}
	claims, err := tok.AsMap(r.Context())
	if err != nil || claims["role"] != "device" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Enheten är inte upplåst"})
		return false
	}
	epoch, _ := claims["epoch"].(float64)
	settings, err := h.store.GetFamilySettings(r.Context())
	if err != nil || int32(epoch) != settings.DeviceEpoch {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Enheten är inte upplåst"})
		return false
	}
	return true
}
```

(OBS jwtauth.Decode verifierar signatur+exp; `AsMap(ctx)` — samma versionskoll som i Task 3. Decode kan också returnera fel för utgången token — samma 401-väg.)

```go
type profileResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// Profiles handles GET /api/profiles (trusted devices only).
func (h *Handler) Profiles(w http.ResponseWriter, r *http.Request) {
	if !h.requireDevice(w, r) {
		return
	}
	parents, err := h.store.ListParents(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to list parents")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}
	students, err := h.store.ListAllStudents(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to list students")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}
	profiles := make([]profileResponse, 0, len(parents)+len(students))
	for _, p := range parents {
		name := p.Email
		if i := strings.Index(name, "@"); i > 0 {
			name = name[:i]
		}
		profiles = append(profiles, profileResponse{ID: uuidToString(p.ID), Name: name, Role: "parent"})
	}
	for _, s := range students {
		profiles = append(profiles, profileResponse{ID: uuidToString(s.ID), Name: s.Name, Role: "child"})
	}
	writeJSON(w, http.StatusOK, profiles)
}

// ProfileLogin handles POST /api/profile/login (trusted devices only).
func (h *Handler) ProfileLogin(w http.ResponseWriter, r *http.Request) {
	if !h.requireDevice(w, r) {
		return
	}
	var req struct {
		ID   string `json:"id"`
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" || (req.Role != "parent" && req.Role != "child") {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Ogiltigt format"})
		return
	}
	id := parseUUID(req.ID)
	var name string
	if req.Role == "parent" {
		p, err := h.store.GetParentByID(r.Context(), id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Profilen hittades inte"})
			return
		}
		name = p.Email
	} else {
		s, err := h.store.GetStudentByID(r.Context(), id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "Profilen hittades inte"})
			return
		}
		name = s.Name
	}
	tokenStr, expiry, err := auth.GenerateToken(req.ID, req.Role)
	if err != nil {
		log.Error().Err(err).Msg("failed to generate token for profile login")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}
	auth.SetAuthCookie(w, tokenStr, expiry)
	writeJSON(w, http.StatusOK, map[string]string{"role": req.Role, "name": name})
}

// SetCode handles POST /api/device/set-code (parent JWT required — mounted
// behind ParentOnly in the router). Changing an existing code requires the
// current code and bumps the device epoch (logs out all devices).
func (h *Handler) SetCode(w http.ResponseWriter, r *http.Request) {
	var req struct {
		NewCode     string `json:"newCode"`
		CurrentCode string `json:"currentCode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Ogiltigt format"})
		return
	}
	if len(req.NewCode) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Koden måste vara minst 6 tecken"})
		return
	}
	if settings, err := h.store.GetFamilySettings(r.Context()); err == nil {
		match, err := auth.ComparePassword(settings.CodeHash, req.CurrentCode)
		if err != nil || !match {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "Fel nuvarande familjekod"})
			return
		}
	}
	hash, err := auth.HashPassword(req.NewCode)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internt fel"})
		return
	}
	if _, err := h.store.UpsertFamilyCode(r.Context(), hash); err != nil {
		log.Error().Err(err).Msg("failed to save family code")
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Kunde inte spara koden"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "Familjekoden är sparad"})
}
```

Lägg också `uuidToString`/`parseUUID` i `backend/internal/device/util.go` — kopiera mönstret från `internal/telegram/util.go` (samma två funktioner). Lägg till `"strings"` i handler-importerna.

- [ ] **Step 4: GREEN** — `go test ./internal/device/ -v && go test ./... && go vet ./...`

- [ ] **Step 5: Commit** — `git commit -m "feat(device): profile picker endpoints and family-code management"`

---

### Task 5: apikey kräver familjekod

**Files:**
- Modify: `backend/internal/apikey/handler.go`
- Modify: `backend/cmd/server/main.go` (wire verifier)
- Test: `backend/internal/apikey/handler_test.go` (utöka)

**Interfaces:**
- Produces: `APIKeyHandler.SetFamilyCodeVerifier(v func(ctx context.Context, code string) (required bool, ok bool))` — `required=false` ⇒ ingen kod satt ⇒ släpp igenom; `required=true && !ok` ⇒ 403 `{"error":"Fel familjekod"}`. Store/Update/Delete läser fältet `familyCode` ur request-bodyn (Delete får nu valfri JSON-body).
- Consumes (main.go): verifier byggd på device-storen:

```go
verifier := func(ctx context.Context, code string) (bool, bool) {
	settings, err := deviceStore.GetFamilySettings(ctx)
	if err != nil {
		return false, false // ingen kod satt — inget krav
	}
	match, err := auth.ComparePassword(settings.CodeHash, code)
	return true, err == nil && match
}
apiKeyHandler.SetFamilyCodeVerifier(verifier)
```

- [ ] **Step 1: Failande test** (utöka handler_test.go — läs filen först och följ dess befintliga test-setup/hjälpare; testet nedan anpassas till den befintliga konstruktorn och auth-contexten som redan används där)

```go
func TestStoreRequiresFamilyCodeWhenSet(t *testing.T) {
	h := newTestHandlerWithParent(t) // återanvänd filens befintliga setup-hjälpare (namnet kan skilja — läs filen)
	h.SetFamilyCodeVerifier(func(_ context.Context, code string) (bool, bool) {
		return true, code == "hemlig1"
	})

	// utan kod → 403
	rr := doStoreRequest(t, h, `{"apiKey":"sk-ant-api03-test"}`)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("status %d, want 403", rr.Code)
	}
	// med rätt kod → som tidigare (201)
	rr = doStoreRequest(t, h, `{"apiKey":"sk-ant-api03-test","familyCode":"hemlig1"}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}
}

func TestStoreNoVerifierUnchanged(t *testing.T) {
	h := newTestHandlerWithParent(t)
	rr := doStoreRequest(t, h, `{"apiKey":"sk-ant-api03-test"}`)
	if rr.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", rr.Code, rr.Body.String())
	}
}
```

(`doStoreRequest` = filens befintliga sätt att göra en autentiserad Store-request; om ingen sådan hjälpare finns, skapa en minimal utifrån hur befintliga tester bygger requests. API-nyckelvalideringen kan kräva ett visst format — läs `validate()` och använd en nyckel som passerar den, samma som befintliga tester använder.)

- [ ] **Step 2: RED** — `go test ./internal/apikey/ -v` → undefined SetFamilyCodeVerifier.

- [ ] **Step 3: Implementera.** I handler.go:

```go
// FamilyCodeVerifier reports whether a family code is required and whether
// the given code is correct. Nil verifier = no requirement (bootstrap).
type FamilyCodeVerifier func(ctx context.Context, code string) (required bool, ok bool)

func (h *APIKeyHandler) SetFamilyCodeVerifier(v FamilyCodeVerifier) { h.verifyFamilyCode = v }

// checkFamilyCode enforces the family-code requirement on mutating calls.
func (h *APIKeyHandler) checkFamilyCode(w http.ResponseWriter, r *http.Request, code string) bool {
	if h.verifyFamilyCode == nil {
		return true
	}
	required, ok := h.verifyFamilyCode(r.Context(), code)
	if required && !ok {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "Fel familjekod"})
		return false
	}
	return true
}
```

Lägg fältet `verifyFamilyCode FamilyCodeVerifier` i `APIKeyHandler`-structen (+ `context`-import). I `Store` och `Update`: lägg `FamilyCode string \`json:"familyCode"\`` i request-structen och anropa `if !h.checkFamilyCode(w, r, req.FamilyCode) { return }` direkt efter JSON-decoden (före validate). I `Delete`: decoda en valfri body `{familyCode}` (ignorera decode-fel för tom body ⇒ tom kod) och gör samma check först.

- [ ] **Step 4: main.go wiring** — skapa `deviceStore := device.NewQueriesStore(q)` (före apikey-initieringen; återanvänds i Task 6) och sätt verifiern enligt Interfaces-blocket. Import av `device`-paketet.

- [ ] **Step 5: GREEN** — `go test ./... && go vet ./... && go build ./...`

- [ ] **Step 6: Commit** — `git commit -m "feat(apikey): require family code for key changes when code is set"`

---

### Task 6: routes i main.go

**Files:**
- Modify: `backend/cmd/server/main.go`

**Interfaces:**
- Consumes: allt från Task 3–5. `deviceStore` finns redan (Task 5). Rate limiter: `child.NewPINRateLimiter(5, 15*time.Minute)` (egen instans för unlock, separat från PIN-loginens).

- [ ] **Step 1: Wiring.** Efter apikey-initieringen:

```go
	// Device unlock + profile picker
	deviceLimiter := child.NewPINRateLimiter(5, 15*time.Minute)
	deviceHandler := device.NewHandler(deviceStore, deviceLimiter)
```

Routes (lägg efter `/api/auth`-blocket):

```go
	// Device trust routes (public unlock; profiles gated by device cookie inside handlers)
	r.Route("/api/device", func(r chi.Router) {
		r.Post("/unlock", deviceHandler.Unlock)

		// Parent-only: set/change the family code
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(tokenAuth))
			r.Use(jwtauth.Authenticator(tokenAuth))
			r.Use(auth.ParentOnly)
			r.Post("/set-code", deviceHandler.SetCode)
		})
	})
	r.Get("/api/profiles", deviceHandler.Profiles)
	r.Post("/api/profile/login", deviceHandler.ProfileLogin)
```

- [ ] **Step 2: Verifiera** — `go build ./... && go test ./...` grönt. Runtime-röktest utan att störa driften (DB kör):

```bash
set -a && source .env && set +a && PORT=8082 go run ./cmd/server/main.go &
sleep 3
curl -s -X POST localhost:8082/api/device/unlock -d '{"code":"x"}' # → {"error":"Ingen familjekod är satt ännu"} ELLER "Fel familjekod" om kod redan satt
curl -s -o /dev/null -w '%{http_code}\n' localhost:8082/api/profiles # → 401
kill %1
```

- [ ] **Step 3: Commit** — `git commit -m "feat(device): wire unlock and profile routes"`

---

### Task 7: frontend — /las-upp, /profiler, rotflöde, api.ts

**Files:**
- Modify: `frontend/src/lib/api.ts`
- Create: `frontend/src/routes/las-upp/+page.svelte`
- Create: `frontend/src/routes/profiler/+page.svelte`
- Modify: `frontend/src/routes/+page.svelte` (rot-redirect)

Läs först `frontend/src/lib/api.ts`, `frontend/src/routes/+page.svelte`, `frontend/src/routes/login/+page.svelte` och `frontend/src/routes/child/login/+page.svelte` — följ deras mönster (Svelte 5 runes, `onclick`, shadcn-komponenter, `goto` från `$app/navigation`).

**Interfaces (api.ts-tillägg):**

```ts
export const device = {
	unlock: (code: string) =>
		apiFetch('/api/device/unlock', { method: 'POST', body: JSON.stringify({ code }) }),
	profiles: () =>
		apiFetch('/api/profiles') as Promise<{ id: string; name: string; role: 'parent' | 'child' }[]>,
	profileLogin: (id: string, role: 'parent' | 'child') =>
		apiFetch('/api/profile/login', { method: 'POST', body: JSON.stringify({ id, role }) }) as Promise<{
			role: string;
			name: string;
		}>,
	setCode: (newCode: string, currentCode?: string) =>
		apiFetch('/api/device/set-code', {
			method: 'POST',
			body: JSON.stringify({ newCode, currentCode })
		})
};
```

- [ ] **Step 1: api.ts** — lägg till `device`-exporten ovan (följ filens exportstil).

- [ ] **Step 2: `/las-upp`** — centrerat kort: rubrik "Nioplugget", text "Ange familjekoden för att låsa upp den här enheten.", ett lösenordsfält (`type="password"`, `autocomplete="off"`), knapp "Lås upp". Vid succé → `goto('/profiler')`. Felmeddelande från API:t visas under fältet (t.ex. "Fel familjekod", "För många försök. Vänta 15 minuter."). Diskret länk längst ner: `<a href="/login">Logga in med lösenord</a>`.

- [ ] **Step 3: `/profiler`** — hämta `device.profiles()` i `onMount`; vid 401 → `goto('/las-upp')`. Rendera rutnät (flex/grid, centrerat): per profil en knapp med rund avatar (64–80 px cirkel, initial i vitt, bakgrundsfärg deterministisk ur namnet) och namnet under. Färghash:

```ts
const AVATAR_COLORS = ['#e11d48', '#7c3aed', '#0891b2', '#16a34a', '#d97706', '#db2777'];
function avatarColor(name: string): string {
	let hash = 0;
	for (const ch of name) hash = (hash * 31 + ch.charCodeAt(0)) | 0;
	return AVATAR_COLORS[Math.abs(hash) % AVATAR_COLORS.length];
}
```

Tryck → `device.profileLogin(id, role)` → parent: `goto('/dashboard')`, child: `goto('/study')`. Rubrik: "Vem pluggar?".

- [ ] **Step 4: Rotflödet i `routes/+page.svelte`** — ersätt sidans nuvarande innehåll med en redirect-logik i `onMount`:

```ts
onMount(async () => {
	try {
		const me = (await auth.me()) as { role: string };
		goto(me.role === 'parent' ? '/dashboard' : '/study', { replaceState: true });
		return;
	} catch {}
	try {
		await device.profiles();
		goto('/profiler', { replaceState: true });
	} catch {
		goto('/las-upp', { replaceState: true });
	}
});
```

(`auth.me` — kontrollera exakt namn i api.ts; sidan visar en tom laddyta under omdirigeringen. Behåll ev. metadata/head. Om landningssidans marketing-innehåll är omfattande: flytta INTE innehållet någonstans — det utgår enligt spec.)

- [ ] **Step 5: Verifiera** — `npx svelte-check --threshold error` (0 nya fel) `&& npm run build`.

- [ ] **Step 6: Commit** — `git commit -m "feat(auth): unlock page, profile picker and root redirect flow"`

---

### Task 8: frontend — Byt profil, Familjekod-kort, städa bort PIN/invite ur UI

**Files:**
- Modify: `frontend/src/routes/+layout.svelte` (eller där utloggning/nav bor — läs filen)
- Modify: `frontend/src/routes/dashboard/+page.svelte`
- Modify: `frontend/src/routes/study/+page.svelte` (om PIN/logout-länk finns där — läs)
- Delete/ersätt länkar till `/child/login` och invite-flödet i UI:t

Läs först: `+layout.svelte`, `dashboard/+page.svelte`, `study/+page.svelte`, sök `grep -rn "child/login\|invite\|Logga ut\|logout" frontend/src/routes frontend/src/lib` för alla förekomster.

- [ ] **Step 1: "Byt profil".** Där "Logga ut"-knappen finns (nav/layout/dashboard): ersätt/komplettera med "Byt profil" som gör `POST /api/auth/logout` (befintlig `auth.logout()` i api.ts — kontrollera namnet) och sedan `goto('/profiler')`. Den ska synas för både förälder och barn. (Om `logout()` idag bara anropas för parent — den fungerar nu för båda rollerna efter Task 2.)

- [ ] **Step 2: Familjekod-kort i dashboard.** Nytt kort "Familjekod" i föräldrapanelen (följ befintliga korts struktur, t.ex. API-nyckelkortet): tre fält — "Ny kod" (password), "Upprepa ny kod" (password), "Nuvarande kod" (password, visas bara om en kod redan finns — enklast: visa alltid med hjälptext "Lämna tom om ingen kod är satt"). Knapp "Spara". Klientvalidering: ny kod ≥ 6 tecken och båda fälten lika, annars felmeddelande utan API-anrop. Anropa `device.setCode(newCode, currentCode || undefined)`. Vid succé: grön bekräftelse "Familjekoden är sparad. Alla enheter behöver låsas upp igen." Vid fel: visa API:ts felmeddelande.

- [ ] **Step 3: API-nyckelkortet.** apikey-anropen i api.ts utökas:

```ts
export const apiKey = {
	get: () => apiFetch('/api/apikey'),
	store: (key: string, familyCode?: string) =>
		apiFetch('/api/apikey', { method: 'POST', body: JSON.stringify({ apiKey: key, familyCode }) }),
	update: (key: string, familyCode?: string) =>
		apiFetch('/api/apikey', { method: 'PUT', body: JSON.stringify({ apiKey: key, familyCode }) }),
	delete: (familyCode?: string) =>
		apiFetch('/api/apikey', { method: 'DELETE', body: JSON.stringify({ familyCode }) })
};
```

I dashboardens API-nyckelkort (och `setup`-sidan om den anropar samma — grep!): lägg ett "Familjekod"-fält (password) intill spara/radera och skicka med värdet. Visa API:ts 403-fel ("Fel familjekod") vid fel kod.

- [ ] **Step 4: Bort ur UI.** Ta bort: länkar/knappar till `/child/login` (t.ex. från login-sidan/landning), "Skapa inbjudningslänk"-UI:t i dashboard (knappar + `inviteLinks`-state + `generateInvite`-anrop; låt `children.create` vara kvar). Route-filerna `child/login` och `invite` får ligga kvar (nås bara via direkt-URL) — bara navigeringen dit tas bort. `/login`-sidan behålls som dold fallback; ta bort ev. länk till registrering/child-login från den om det stökar, annars lämna.

- [ ] **Step 5: Verifiera** — `npx svelte-check --threshold error && npm run build` grönt. Grep-kontroll: `grep -rn "child/login" frontend/src/routes frontend/src/lib` ska inte träffa någon navigationslänk (route-filen själv får finnas).

- [ ] **Step 6: Commit** — `git commit -m "feat(auth): switch-profile flow, family code card, remove PIN/invite from UI"`

---

### Task 9: Deploy + live-verifiering (körs av controllern på servern)

- [ ] Merge till master (efter slutreview), `go build -o server ./cmd/server`, `npm run build`
- [ ] Migration 013 mot driftdatabasen (redan körd i Task 1 om samma DB — verifiera med `SELECT * FROM family_settings;` att tabellen finns)
- [ ] `systemctl --user restart nioplugget-backend nioplugget-frontend`
- [ ] Verifiera publikt: `https://nioplugget.vic-corp.net/` → redirect till `/las-upp`; sätt familjekod via `/login` + dashboard; lås upp; profilväljaren visar familjen; tryck på barnprofil → `/study`
- [ ] Uppdatera minnesanteckningar

---

## Self-review (utförd vid planskrivning)

- Spec-täckning: migration/queries (1), 30d-JWT + logout båda roller (2), unlock+ratelimit+cookie (3), profiles/login/set-code+epoch (4), apikey-familjekod (5), routes (6), frontend-flöden (7), Byt profil/kort/städning (8), deploy (9). Rate limit per IP med CF-header: Task 3. Epoch-utkastning: Task 1 (query) + 4 (check + bump).
- Kontrollpunkter i stället för placeholders: sqlc-genererade namn (T1/T3), jwx `AsMap`-signatur (T3/T4), apikey-testets befintliga hjälpare (T5), api.ts/logout-namn och kortstruktur (T7/T8).
- Typkonsistens: `FamilyCodeVerifier(ctx, code) (required, ok)` samma i T5-interfaces och main-wiring; `GenerateDeviceToken(epoch int32)` matchar `DeviceEpoch int32`; API-kontrakten i T4 matchar T7:s api.ts.
