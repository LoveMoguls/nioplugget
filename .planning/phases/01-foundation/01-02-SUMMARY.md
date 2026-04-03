---
phase: 01-foundation
plan: 02
subsystem: auth
tags: [go, jwt, argon2id, jwtauth, chi, httponly-cookie, tdd]

# Dependency graph
requires:
  - phase: 01-01-foundation
    provides: "Go backend scaffold, parents table, sqlc CreateParent/GetParentByEmail queries"
provides:
  - Argon2id password hashing (HashPassword, ComparePassword)
  - JWT generation with sub, role, exp claims (24h expiry)
  - ParentOnly and ChildOnly role-based middleware
  - Register, Login, Logout HTTP handlers with httpOnly JWT cookies
  - ParentQuerier interface for testable handler injection
  - GetUserIDFromContext and GetRoleFromContext helpers
affects:
  - 01-03-foundation (API key handler uses auth.ParentOnly middleware and GetUserIDFromContext)
  - 01-04-foundation (child invite/login needs ChildOnly middleware)
  - all subsequent phases (protected routes use ParentOnly/ChildOnly)

# Tech tracking
tech-stack:
  added:
    - github.com/alexedwards/argon2id v1.0.0 (Argon2id password hashing)
    - github.com/go-chi/jwtauth/v5 v5.4.0 (JWT auth for Chi)
    - github.com/lestrrat-go/jwx/v3 v3.0.2 (JWT implementation dependency)
  patterns:
    - TDD: failing test commit -> implementation commit for each module
    - ParentQuerier interface enables mock injection in tests (no real DB needed)
    - Same 401 error for unknown email and wrong password (prevents user enumeration)
    - httpOnly + Secure + SameSite=Lax JWT cookie for XSS/CSRF resistance
    - jwtauth.Verifier + jwtauth.Authenticator + ParentOnly stacked in Chi route groups

key-files:
  created:
    - backend/internal/auth/hash.go
    - backend/internal/auth/hash_test.go
    - backend/internal/auth/jwt.go
    - backend/internal/auth/jwt_test.go
    - backend/internal/auth/middleware.go
    - backend/internal/auth/middleware_test.go
    - backend/internal/auth/handler.go
    - backend/internal/auth/handler_test.go
  modified:
    - backend/cmd/server/main.go (wired auth routes, initialized TokenAuth)
    - backend/go.mod (added argon2id, jwtauth/v5)

key-decisions:
  - "ParentQuerier interface over concrete *queries.Queries — allows test mocking without a live database"
  - "Same 401 error for unknown email and wrong password — prevents user enumeration attacks"
  - "httpOnly + Secure + SameSite=Lax JWT cookie — XSS cannot steal token; CSRF mitigated via Lax same-site policy"
  - "jwtauth/v5 with lestrrat-go/jwx/v3 — Chi-native JWT library, consistent with Chi ecosystem"

patterns-established:
  - "Auth middleware stack: jwtauth.Verifier -> jwtauth.Authenticator -> ParentOnly/ChildOnly"
  - "ParentQuerier interface pattern: testable handler injection via interface, real queries.Queries implements it at runtime"
  - "Cookie pattern: setAuthCookie(w, token, expiry) / clearAuthCookie(w) helpers used in both Register and Login"

requirements-completed: [AUTH-01, AUTH-02, AUTH-03, SEC-03]

# Metrics
duration: 8min
completed: 2026-04-03
---

# Phase 1 Plan 02: Parent Authentication Summary

**Argon2id password hashing + HS256 JWT session management with httpOnly cookies, role-based middleware, and register/login/logout handlers — all implemented via TDD with 24 automated tests**

## Performance

- **Duration:** ~8 min
- **Started:** 2026-04-03T07:58:26Z
- **Completed:** 2026-04-03T08:06:45Z
- **Tasks:** 3
- **Files modified:** 10 (8 created, 2 modified)

## Accomplishments

- Argon2id password hashing with unique salts per hash, explicit empty password rejection
- JWT generation with sub/role/exp claims, 24-hour expiry, package-level TokenAuth for Chi middleware
- ParentOnly and ChildOnly middleware for role-based route protection (403 for wrong role, 401 for no token)
- Register/Login/Logout handlers with httpOnly JWT cookies, GDPR consent timestamp stored at registration
- ParentQuerier interface enables test mocking without a live database

## Task Commits

Each task was committed atomically:

1. **Task 1: Argon2id password hashing service** - `a4d5b26` (feat)
2. **Task 2: JWT generation and auth middleware** - `920cf5c` (feat)
3. **Task 3: Register, Login, Logout HTTP handlers** - `9b3cf27` (feat)

**Plan metadata:** committed with state updates

## Files Created/Modified

- `backend/internal/auth/hash.go` - HashPassword (argon2id) and ComparePassword
- `backend/internal/auth/hash_test.go` - 5 tests: hash returns non-empty, empty rejects, correct matches, wrong doesn't match, unique salts
- `backend/internal/auth/jwt.go` - NewTokenAuth, TokenAuth var, GenerateToken with 24h expiry
- `backend/internal/auth/jwt_test.go` - 6 tests: parent/child role tokens, claim verification (sub/role/exp)
- `backend/internal/auth/middleware.go` - ParentOnly, ChildOnly, GetUserIDFromContext, GetRoleFromContext
- `backend/internal/auth/middleware_test.go` - 5 tests: no token=401, wrong role=403, correct role=200
- `backend/internal/auth/handler.go` - AuthHandler, ParentQuerier interface, Register/Login/Logout
- `backend/internal/auth/handler_test.go` - 8 tests: register 201/409/400, login 200/401, logout 200
- `backend/cmd/server/main.go` - Wired auth routes with JWT middleware stack, initialized TokenAuth

## Decisions Made

- **ParentQuerier interface** over injecting `*queries.Queries` directly — enables test mocking without a live database, keeping handler tests fast and hermetic
- **Same 401 for unknown email and wrong password** — prevents user enumeration (attacker cannot distinguish "email not found" from "wrong password")
- **httpOnly + Secure + SameSite=Lax cookie** — httpOnly prevents XSS token theft; SameSite=Lax mitigates CSRF for navigation requests while allowing normal browser use
- **jwtauth/v5 with lestrrat-go/jwx/v3 API** — the `Token` interface in jwx/v3 exposes claims via `Get(name, &dst)` not `AsMap()` (breaking change from v2)

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] lestrrat-go/jwx/v3 Token interface incompatibility**
- **Found during:** Task 2 (JWT tests)
- **Issue:** Test code used `token.AsMap(nil)` and `token.Expiration()` (single return) from jwx/v2 API, but go-chi/jwtauth/v5 uses jwx/v3 which changed the interface: `Expiration()` returns `(time.Time, bool)` and the `Token` interface no longer exposes `PrivateClaims()` or `AsMap()` — custom claims must be accessed via `token.Get("role", &dst)`
- **Fix:** Updated test to use `token.Get("role", &role)` and `exp, expOk := token.Expiration()`
- **Files modified:** backend/internal/auth/jwt_test.go
- **Verification:** All 6 JWT tests pass
- **Committed in:** 920cf5c (Task 2 commit)

---

**Total deviations:** 1 auto-fixed (API version incompatibility in test code)
**Impact on plan:** Tests correctly reflect the actual jwx/v3 API used at runtime. No behavior change.

## Issues Encountered

- go-chi/jwtauth/v5 upgraded golang.org/x/crypto to v0.38.0 and golang.org/x/sys to v0.33.0 — handled automatically by `go get`

## User Setup Required

None - no external service configuration required. JWT secret is read from JWT_SECRET environment variable at runtime.

## Next Phase Readiness

- Ready for 01-03-PLAN.md: API key management — auth.ParentOnly middleware and GetUserIDFromContext are in place
- Ready for 01-04-PLAN.md: Child invite/login — auth.ChildOnly middleware is in place
- All 24 auth tests pass (hash: 5, JWT: 6, middleware: 5, handler: 8)

---
*Phase: 01-foundation*
*Completed: 2026-04-03*
