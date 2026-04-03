---
phase: 01-foundation
plan: 03
subsystem: api
tags: [go, aes-256-gcm, encryption, claude-api, jwt, chi]

# Dependency graph
requires:
  - phase: 01-foundation/01-01
    provides: sqlc-generated queries (UpsertAPIKey, GetAPIKeyByParentID, DeleteAPIKeyByParentID), api_keys table schema
  - phase: 01-foundation/01-02
    provides: auth.GetUserIDFromContext, auth.ParentOnly middleware, jwtauth token context
provides:
  - AES-256-GCM EncryptionService (Encrypt/Decrypt with random nonces)
  - Claude API key validation (live /v1/models check with Swedish error messages)
  - MaskAPIKey function returning "sk-ant...****" format
  - APIKeyHandler with Store/Get/Update/Delete methods
  - APIKeyStore interface for testable storage abstraction
  - QueriesStore adapter bridging sqlc queries to APIKeyStore interface
  - /api/apikey routes wired under ParentOnly JWT middleware in main.go
affects:
  - 01-04-foundation (invite/child login may need ENCRYPTION_KEY setup awareness)
  - 02-chat (will use EncryptionService to decrypt stored key for Claude API proxying)
  - all subsequent phases that need the Claude API key

# Tech tracking
tech-stack:
  added:
    - crypto/aes + crypto/cipher (stdlib AES-256-GCM)
    - crypto/rand (stdlib nonce generation)
    - net/http (stdlib HTTP client for Anthropic validation)
  patterns:
    - Interface-based store abstraction (APIKeyStore) for handler testability
    - TDD with httptest.Server for mocking external APIs (Anthropic)
    - Nonce-prepended ciphertext format for AES-256-GCM
    - validateURL injection pattern for testable external HTTP calls
    - Swedish error messages for user-facing validation errors

key-files:
  created:
    - backend/internal/apikey/encrypt.go
    - backend/internal/apikey/encrypt_test.go
    - backend/internal/apikey/validate.go
    - backend/internal/apikey/validate_test.go
    - backend/internal/apikey/handler.go
    - backend/internal/apikey/handler_test.go
    - backend/internal/apikey/store.go
  modified:
    - backend/cmd/server/main.go

key-decisions:
  - "APIKeyStore interface decouples handler from sqlc — enables unit testing without DB"
  - "validateURL parameter on NewAPIKeyHandler allows test injection without global state"
  - "QueriesStore adapter in store.go keeps handler independent of sqlc implementation details"
  - "Raw key never logged or returned — only masked format (sk-ant...****) in responses"
  - "ErrNotFound sentinel value for DB not-found distinguishes 404 from 500 in Get handler"

patterns-established:
  - "Store interface + adapter pattern: define interface in handler package, implement adapter separately"
  - "External URL injection for testable HTTP clients: pass URL string rather than http.Client"
  - "AES-256-GCM nonce prepend: nonce || ciphertext for self-contained ciphertext blobs"

requirements-completed: [KEY-01, KEY-02, KEY-03, KEY-04]

# Metrics
duration: 12min
completed: 2026-04-03
---

# Phase 1 Plan 03: API Key Management Summary

**AES-256-GCM encryption service and Claude API key CRUD (store/get/update/delete) with live validation against Anthropic API before storage and masked output**

## Performance

- **Duration:** ~12 min
- **Started:** 2026-04-03T08:10:00Z
- **Completed:** 2026-04-03T08:22:00Z
- **Tasks:** 3
- **Files modified:** 8 (7 created, 1 modified)

## Accomplishments

- AES-256-GCM EncryptionService with random nonces, round-trip verified, wrong-key and corrupted-ciphertext errors handled
- ValidateAPIKey calls Anthropic /v1/models with 10-second timeout; httptest mock used in unit tests so no live API calls during testing
- API key CRUD handler: Store validates live before encrypting, Get decrypts only to mask (raw key never in response), Update re-validates and re-encrypts, Delete removes with confirmation
- All routes wired into main.go under ParentOnly JWT middleware; ENCRYPTION_KEY read from env at startup

## Task Commits

Each task was committed atomically:

1. **Task 1: AES-256-GCM encryption service** - `9fecb12` (feat)
2. **Task 2: Claude API key validation and masking** - `51e2e18` (feat)
3. **Task 3: API key CRUD handlers** - `6609114` (feat)

## Files Created/Modified

- `backend/internal/apikey/encrypt.go` - EncryptionService with Encrypt/Decrypt using AES-256-GCM
- `backend/internal/apikey/encrypt_test.go` - 7 tests: nonce uniqueness, round-trip, wrong key, corrupted ciphertext
- `backend/internal/apikey/validate.go` - ValidateAPIKey, ValidateAPIKeyWithURL, MaskAPIKey
- `backend/internal/apikey/validate_test.go` - 8 tests: empty key, valid/invalid/server-error/connection-error, masking variants
- `backend/internal/apikey/handler.go` - APIKeyHandler (Store/Get/Update/Delete), APIKeyStore interface, ErrNotFound
- `backend/internal/apikey/handler_test.go` - 6 tests covering all CRUD paths and raw-key exposure check
- `backend/internal/apikey/store.go` - QueriesStore adapting sqlc queries.Queries to APIKeyStore
- `backend/cmd/server/main.go` - Added apikey import, EncryptionService init, /api/apikey routes under ParentOnly middleware

## Decisions Made

- **APIKeyStore interface** defined in handler package with QueriesStore adapter in store.go — clean separation, unit-testable without DB
- **validateURL injection** via NewAPIKeyHandler parameter instead of global var — enables httptest mocking in tests cleanly
- **ErrNotFound sentinel** used to distinguish "no key stored" (404) from "DB error" (500) in Get handler
- **Raw key never leaves handler** — Get decrypts only to derive masked format; decrypted bytes not stored anywhere

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Discovered auth package already complete from 01-02**
- **Found during:** Task 3 (API key CRUD handlers)
- **Issue:** Handler needs auth.GetUserIDFromContext; initially created context.go stub but auth package from 01-02 was already present with full implementation
- **Fix:** Removed redundant context.go; used existing middleware.go which already exports GetUserIDFromContext
- **Files modified:** None (no stub committed)
- **Verification:** go build ./internal/auth/ succeeds
- **Committed in:** N/A (removed before commit)

**2. [Rule 2 - Missing Critical] Added QueriesStore adapter**
- **Found during:** Task 3 (wiring in main.go)
- **Issue:** Plan specified handler uses queries.Queries directly but APIKeyStore interface required an adapter to bridge to sqlc-generated type (queries.UpsertAPIKeyParams etc.)
- **Fix:** Created store.go with QueriesStore wrapping queries.Queries, converting pgtype.Timestamptz to time.Time
- **Files modified:** backend/internal/apikey/store.go
- **Verification:** go build ./cmd/server/ succeeds
- **Committed in:** 6609114 (Task 3 commit)

---

**Total deviations:** 2 auto-fixed (1 unnecessary stub removed, 1 missing adapter added)
**Impact on plan:** Both handled correctly. Store adapter was architecturally necessary to keep handler interface clean.

## Issues Encountered

None — all three TDD cycles completed cleanly with tests failing (RED) before implementation and passing (GREEN) after.

## User Setup Required

None - ENCRYPTION_KEY must be a hex-encoded 32-byte value (64 hex chars). Startup fails fast if not set.

## Next Phase Readiness

- Ready for 01-04-PLAN.md: Child invite/login — schema and query scaffolding already present
- EncryptionService in `internal/apikey/encrypt.go` is ready for Phase 2 use (decrypt stored key for Claude API proxying)
- All 21 apikey tests pass: `cd backend && go test ./internal/apikey/ -v`

---
*Phase: 01-foundation*
*Completed: 2026-04-03*

## Self-Check: PASSED

All created files verified present. All task commits (9fecb12, 51e2e18, 6609114) verified in git history.
