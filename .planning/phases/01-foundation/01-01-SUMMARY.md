---
phase: 01-foundation
plan: 01
subsystem: database
tags: [go, chi, pgx, postgres, sqlc, zerolog, cors, migrations]

# Dependency graph
requires: []
provides:
  - Go backend with Chi router and HTTP server entry point
  - pgxpool connection pool configured for PostgreSQL
  - Database migration schema: parents, students, api_keys, pin_attempts tables
  - sqlc-generated type-safe query functions for all tables
  - Redacting logger middleware (never logs Authorization, x-api-key, or sensitive query params)
  - CORS middleware configured via FRONTEND_URL env
affects:
  - 01-02-foundation (parent auth builds on this schema and server)
  - 01-03-foundation (API key CRUD uses api_keys table and queries)
  - 01-04-foundation (child invite/login uses students table and queries)
  - all subsequent phases

# Tech tracking
tech-stack:
  added:
    - github.com/go-chi/chi/v5 v5.2.5
    - github.com/go-chi/cors v1.2.2
    - github.com/jackc/pgx/v5 v5.9.1 (pgxpool)
    - github.com/rs/zerolog v1.35.0
    - sqlc v1.30.0 (code generation tool)
  patterns:
    - Chi router with middleware stack (RedactingLogger, Recoverer, CORS)
    - pgxpool.NewWithConfig for tunable pool parameters
    - sqlc query annotations for type-safe DB access
    - Sensitive field redaction at the middleware layer (not in handler code)

key-files:
  created:
    - backend/cmd/server/main.go
    - backend/internal/database/pool.go
    - backend/internal/middleware/logger.go
    - backend/internal/middleware/cors.go
    - backend/db/migrations/001_initial_schema.up.sql
    - backend/db/migrations/001_initial_schema.down.sql
    - backend/db/queries/parents.sql
    - backend/db/queries/students.sql
    - backend/db/queries/api_keys.sql
    - backend/sqlc.yaml
    - backend/internal/database/queries/db.go
    - backend/internal/database/queries/models.go
    - backend/internal/database/queries/parents.sql.go
    - backend/internal/database/queries/students.sql.go
    - backend/internal/database/queries/api_keys.sql.go
  modified: []

key-decisions:
  - "Logger redacts at middleware level using header name denylist — no risk of accidentally logging secrets in handler code"
  - "pgxpool configured with max 25 conns, min 5, 30s health check — tuned for moderate load"
  - "sqlc with pgx/v5 driver and JSON tags — type-safe queries without ORM overhead"
  - "CORS configured from FRONTEND_URL env (default localhost:5173) — allows easy local dev and production config"

patterns-established:
  - "Sensitive field redaction via middleware denylist: [authorization, x-api-key, api_key, password, pin, token, secret, encrypted_key]"
  - "pool.go wraps pgxpool.NewWithConfig — single place to tune pool parameters"
  - "mustEnv() helper for required env vars — fail-fast at startup instead of panicking later"

requirements-completed: [SEC-01]

# Metrics
duration: 2min
completed: 2026-04-03
---

# Phase 1 Plan 01: Foundation Scaffold Summary

**Chi router + pgxpool backend with parents/students/api_keys schema, sqlc type-safe queries, and zerolog-based redacting logger that never outputs Authorization headers or API keys**

## Performance

- **Duration:** ~2 min
- **Started:** 2026-04-03T07:53:28Z
- **Completed:** 2026-04-03T07:55:43Z
- **Tasks:** 3
- **Files modified:** 15 created

## Accomplishments

- Go backend scaffolded with Chi router, all placeholder route groups (/api/auth, /api/apikey, /api/children, /api/child, /api/invite), and pgxpool connection
- PostgreSQL schema with four tables: parents, students, api_keys, pin_attempts — includes pgcrypto extension for UUID generation
- sqlc generates 5 Go files with type-safe functions for all CRUD operations using pgx/v5 driver
- Redacting logger middleware prevents sensitive data leakage at the HTTP layer for SEC-01 compliance

## Task Commits

Each task was committed atomically:

1. **Task 1: Scaffold Go project with Chi router and PostgreSQL pool** - `ef4167e` (feat)
2. **Task 2: Create database schema migration and sqlc configuration** - `f8c84bc` (feat)
3. **Task 3: Create redacting logger middleware** - included in `ef4167e` (required for compilation)

## Files Created/Modified

- `backend/cmd/server/main.go` - HTTP server entry point with Chi router, env loading, pgxpool init, placeholder routes
- `backend/go.mod` - Go module github.com/trollstaven/nioplugget/backend
- `backend/go.sum` - Dependency checksums
- `backend/internal/database/pool.go` - pgxpool.NewWithConfig with tuned parameters
- `backend/internal/middleware/cors.go` - CORS middleware from FRONTEND_URL env
- `backend/internal/middleware/logger.go` - Redacting request logger using zerolog
- `backend/db/migrations/001_initial_schema.up.sql` - Schema creation with pgcrypto extension
- `backend/db/migrations/001_initial_schema.down.sql` - Schema rollback
- `backend/db/queries/parents.sql` - CreateParent, GetParentByEmail, GetParentByID
- `backend/db/queries/students.sql` - CreateStudent, GetStudentsByParentID, ActivateStudent, GetStudentByNameAndParent, GetStudentByID, ListStudentNamesByParentID
- `backend/db/queries/api_keys.sql` - UpsertAPIKey, GetAPIKeyByParentID, DeleteAPIKeyByParentID
- `backend/sqlc.yaml` - sqlc configuration for pgx/v5
- `backend/internal/database/queries/*.go` - 5 generated files (db.go, models.go, parents.sql.go, students.sql.go, api_keys.sql.go)

## Decisions Made

- Logger middleware included in Task 1 commit because it was required for main.go compilation (it's wired via r.Use())
- pgx/v5 required upgrading Go to 1.25 (pgx/v5 v5.9.1 requirement) — handled automatically by go get

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 3 - Blocking] Logger created before Chi scaffold**
- **Found during:** Task 1 (Go project scaffold)
- **Issue:** main.go references `appMiddleware.RedactingLogger` so the logger file must exist before compilation
- **Fix:** Created logger.go during Task 1 instead of Task 3, committed together
- **Files modified:** backend/internal/middleware/logger.go
- **Verification:** `go build ./cmd/server/` passes
- **Committed in:** ef4167e (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (1 blocking dependency — logger needed before compilation)
**Impact on plan:** Logger was already planned for Task 3; creating it in Task 1 is a commit ordering adjustment, not scope change.

## Issues Encountered

- pgx/v5 v5.9.1 requires go 1.25.0+; go get automatically upgraded the go directive in go.mod from 1.24.1 to 1.25.0
- Missing pgxpool puddle dependency required explicit `go get github.com/jackc/pgx/v5/pgxpool@v5.9.1` followed by `go mod tidy`

## User Setup Required

None - no external service configuration required beyond setting environment variables at runtime (DATABASE_URL, JWT_SECRET, ENCRYPTION_KEY, PORT, FRONTEND_URL).

## Next Phase Readiness

- Ready for 01-02-PLAN.md: Parent auth (Argon2id, JWT, register/login/logout) — schema and query scaffolding are in place
- ActivateStudent query in students.sql already implements the atomic invite-code activation logic needed for 01-04
- All sqlc-generated types available for handler development in subsequent plans

---
*Phase: 01-foundation*
*Completed: 2026-04-03*
