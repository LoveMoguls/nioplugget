---
phase: 01-foundation
plan: 04
subsystem: auth
tags: [go, chi, argon2id, jwt, pin-auth, invite-link, rate-limiting, tdd]

# Dependency graph
requires:
  - phase: 01-02-foundation
    provides: "Argon2id HashPassword/ComparePassword, GenerateToken, ParentOnly/ChildOnly middleware, GetUserIDFromContext"
  - phase: 01-01-foundation
    provides: "sqlc-generated students queries (CreateStudent, ActivateStudent, GetStudentByNameAndParent, GetStudentsByParentID, ListStudentNamesByParentID, GetStudentByID)"
provides:
  - GenerateInviteToken: 32-byte crypto/rand token, base64url encoded (44 chars)
  - PINRateLimiter: in-memory, 5-attempt limit with 15-min lockout, per-studentID counters
  - ChildHandler: Create, List, GenerateInvite, Activate, PINLogin, ListNames HTTP handlers
  - ChildQuerier interface for test mocking without a live database
  - QueriesStore wrapping sqlc queries + raw UpdateStudentInvite query
  - Routes: POST/GET /api/children (parent-protected), POST /api/children/{id}/invite, POST /api/child/login, GET /api/child/names, POST /api/invite/{token}/activate
affects:
  - 01-05-foundation (final plan in phase — all child routes now functional)
  - all subsequent phases using child sessions (ChildOnly middleware now has a session source)

# Tech tracking
tech-stack:
  added: []
  patterns:
    - TDD: failing test commit (RED) -> implementation commit (GREEN) for each module
    - ChildQuerier interface: testable handler injection, same pattern as ParentQuerier in auth package
    - QueriesStore: wraps *queries.Queries + queries.DBTX for raw SQL when sqlc lacks a query
    - Rate limiter: in-memory sync.Mutex map, independent per-ID counters, auto-reset on expiry
    - PIN auth: same Argon2id as parent passwords, pinRegex validation before hashing

key-files:
  created:
    - backend/internal/child/invite.go
    - backend/internal/child/invite_test.go
    - backend/internal/child/ratelimit.go
    - backend/internal/child/ratelimit_test.go
    - backend/internal/child/handler.go
    - backend/internal/child/handler_test.go
    - backend/internal/child/store.go
    - backend/internal/child/testhelpers_test.go
  modified:
    - backend/cmd/server/main.go (wired all child routes, removed stub route groups)

key-decisions:
  - "ChildQuerier interface over *queries.Queries — enables unit tests without a live database, consistent with ParentQuerier pattern from 01-02"
  - "UpdateStudentInvite via raw SQL in QueriesStore — sqlc-generated code lacks this query; passing queries.DBTX to store avoids touching generated files"
  - "In-memory PINRateLimiter (not DB-backed) — sufficient for MVP; survives process restarts only (acceptable trade-off for simplicity)"
  - "auth.TokenAuth init() in testhelpers_test.go — GenerateToken uses package-level var; must be initialized before any handler test that calls Activate or PINLogin"

patterns-established:
  - "ChildQuerier interface pattern: handler depends on interface, QueriesStore implements at runtime, mockQuerier in tests"
  - "Rate limiter usage: Check(id) before hash compare; Reset(id) only on successful login"
  - "Token injection in tests: injectParentJWT helper sets JWT claims in context via jwtauth.NewContext"

requirements-completed: [AUTH-04, AUTH-05, AUTH-06, AUTH-07, AUTH-08, SEC-02]

# Metrics
duration: 5min
completed: 2026-04-03
---

# Phase 1 Plan 04: Child Profile Management Summary

**Invite-link child onboarding with 4-digit PIN auth — parent creates child, system generates single-use 72h invite, child activates with PIN (Argon2id), PIN login with 5-attempt rate limiting and 15-min lockout**

## Performance

- **Duration:** ~5 min
- **Started:** 2026-04-03T08:06:09Z
- **Completed:** 2026-04-03T08:11:29Z
- **Tasks:** 2
- **Files modified:** 9 (8 created, 1 modified)

## Accomplishments

- Cryptographically random invite tokens: 32 bytes from crypto/rand, base64url encoded, 44 characters
- In-memory PIN rate limiter with sync.Mutex, 5-attempt limit, 15-minute lockout, automatic reset on expiry
- Full child lifecycle: parent creates child -> 72h invite link generated -> child activates with 4-digit PIN (Argon2id hashed) -> child logs in by name + PIN
- Invite atomically invalidated on activation (single UPDATE sets pin_hash, activated_at, clears invite_code and invite_expires_at)
- All routes wired in main.go replacing placeholder stubs

## Task Commits

Each task was committed atomically:

1. **Task 1: Invite token generation and PIN rate limiter (RED)** - `8f89425` (test)
2. **Task 1: Invite token generation and PIN rate limiter (GREEN)** - `e302e87` (feat)
3. **Task 2: Child management and invite/activation handlers (RED)** - `ff45569` (test)
4. **Task 2: Child management and invite/activation handlers (GREEN)** - `940c379` (feat)

**Plan metadata:** committed with state updates

## Files Created/Modified

- `backend/internal/child/invite.go` - GenerateInviteToken: 32 bytes crypto/rand, base64url
- `backend/internal/child/invite_test.go` - 3 tests: length, uniqueness, URL-safe chars
- `backend/internal/child/ratelimit.go` - PINRateLimiter with sync.Mutex map, Check/Reset
- `backend/internal/child/ratelimit_test.go` - 6 tests: first attempt, lockout, locked, expiry reset, Reset(), independence
- `backend/internal/child/handler.go` - ChildHandler, ChildQuerier interface, UpdateStudentInviteParams, all 6 handlers
- `backend/internal/child/handler_test.go` - 12 tests covering full child lifecycle
- `backend/internal/child/store.go` - QueriesStore wrapping sqlc + raw UpdateStudentInvite SQL
- `backend/internal/child/testhelpers_test.go` - mustHashPIN helper, init() for auth.TokenAuth
- `backend/cmd/server/main.go` - Wired child routes: /api/children, /api/child, /api/invite

## Decisions Made

- **ChildQuerier interface** over injecting `*queries.Queries` — consistent with ParentQuerier pattern from 01-02, enables hermetic handler tests
- **UpdateStudentInvite via raw SQL** — sqlc-generated code only has ActivateStudent (which atomically clears invite fields); refreshing an invite for an unactivated child required a separate UPDATE not in the schema. Rather than regenerating sqlc, the raw query is executed via the `queries.DBTX` interface passed directly to QueriesStore
- **In-memory PINRateLimiter** — DB-backed rate limiting not needed for MVP; in-memory is simpler, fast, and sufficient for single-process deployment
- **auth.TokenAuth init() in test helpers** — GenerateToken relies on the package-level `auth.TokenAuth` variable; calling `auth.NewTokenAuth("test-secret")` in `init()` ensures it's initialized before any handler test that calls Activate or PINLogin

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] auth.TokenAuth nil in handler tests**
- **Found during:** Task 2 (TestHandlerActivateSuccess)
- **Issue:** `auth.GenerateToken` uses package-level `auth.TokenAuth` which is nil until `auth.NewTokenAuth()` is called. Handler tests calling Activate or PINLogin panic with nil pointer dereference.
- **Fix:** Added `init()` in testhelpers_test.go calling `auth.NewTokenAuth("test-secret-for-child-handler-tests")` to initialize the package-level variable before tests run.
- **Files modified:** backend/internal/child/testhelpers_test.go
- **Verification:** TestHandlerActivateSuccess and TestHandlerPINLoginSuccess pass
- **Committed in:** 940c379 (Task 2 feat commit)

---

**Total deviations:** 1 auto-fixed (nil pointer in test setup)
**Impact on plan:** Fix was necessary to run handler tests. No behavior change in production code.

## Issues Encountered

- `queries.Queries.db` field is unexported (by sqlc design), so QueriesStore cannot use it for raw queries. Resolved by passing `queries.DBTX` (the pool) separately to `child.NewQueriesStore(q, pool)` in main.go.

## User Setup Required

None - no external service configuration required. All new endpoints use the existing JWT_SECRET and database connection.

## Next Phase Readiness

- Ready for 01-05-PLAN.md (final foundation plan)
- All 21 child tests pass (invite: 3, rate limiter: 6, handler: 12)
- Full backend test suite passes: auth (24), apikey (implicit), child (21)
- Child JWT sessions are now functional — ChildOnly middleware has a real session source

## Self-Check: PASSED

- backend/internal/child/handler.go: FOUND
- backend/internal/child/ratelimit.go: FOUND
- backend/internal/child/invite.go: FOUND
- backend/internal/child/store.go: FOUND
- .planning/phases/01-foundation/01-04-SUMMARY.md: FOUND
- Commit 8f89425 (test RED): FOUND
- Commit e302e87 (feat GREEN): FOUND
- Commit ff45569 (test RED): FOUND
- Commit 940c379 (feat GREEN): FOUND

---
*Phase: 01-foundation*
*Completed: 2026-04-03*
