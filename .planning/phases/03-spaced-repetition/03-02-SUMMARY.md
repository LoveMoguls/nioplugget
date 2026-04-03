---
phase: 03-spaced-repetition
plan: 02
subsystem: database
tags: [postgresql, sqlc, pgx, migrations, spaced-repetition, sm2]

# Dependency graph
requires:
  - phase: 02-content-and-sessions
    provides: exercises, topics, subjects, students tables referenced by review_schedule
provides:
  - review_schedule table DDL (migration 004)
  - sqlc queries: UpsertReviewSchedule, ListDueReviews, GetReviewSchedule
  - ReviewSchedule struct in models.go
affects: [03-spaced-repetition, integration handlers that schedule reviews after sessions]

# Tech tracking
tech-stack:
  added: []
  patterns: [sqlc upsert with ON CONFLICT DO UPDATE for idempotent schedule writes]

key-files:
  created:
    - backend/db/migrations/004_review_schedule.up.sql
    - backend/db/migrations/004_review_schedule.down.sql
    - backend/db/queries/review_schedule.sql
    - backend/internal/database/queries/review_schedule.sql.go
  modified:
    - backend/internal/database/queries/models.go

key-decisions:
  - "ListDueReviews joins exercises/topics/subjects to return enriched rows - avoids N+1 in due-item listing"
  - "UNIQUE(student_id, exercise_id) enables idempotent upsert - safe to call after every session"

patterns-established:
  - "Upsert pattern: ON CONFLICT DO UPDATE for SM-2 state - avoids separate insert/update paths"

requirements-completed: [SRS-02]

# Metrics
duration: 2min
completed: 2026-04-03
---

# Phase 03 Plan 02: Review Schedule DB Summary

**PostgreSQL review_schedule table with UNIQUE constraint, composite index, and sqlc queries for SM-2 persistence (upsert, list-due with joins, get-by-student-exercise)**

## Performance

- **Duration:** 2 min
- **Started:** 2026-04-03T09:13:59Z
- **Completed:** 2026-04-03T09:15:50Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- Migration 004 creates review_schedule with all SM-2 state columns (ease_factor, interval_days, repetition_count, next_review, last_reviewed_at)
- UNIQUE(student_id, exercise_id) constraint enables safe idempotent upsert after every session
- Composite index idx_review_schedule_due on (student_id, next_review) enables efficient due-item queries
- sqlc generates type-safe Go functions for all three query operations; ListDueReviews enriches rows with exercise/topic/subject data via joins

## Task Commits

Each task was committed atomically:

1. **Task 1: Create review_schedule migration** - `c0f6795` (feat)
2. **Task 2: Create sqlc queries and regenerate** - `7e4e1aa` (feat)

**Plan metadata:** (docs commit follows)

## Files Created/Modified
- `backend/db/migrations/004_review_schedule.up.sql` - CREATE TABLE review_schedule with SM-2 columns, UNIQUE constraint, and composite index
- `backend/db/migrations/004_review_schedule.down.sql` - DROP TABLE IF EXISTS review_schedule
- `backend/db/queries/review_schedule.sql` - UpsertReviewSchedule, ListDueReviews, GetReviewSchedule query definitions
- `backend/internal/database/queries/review_schedule.sql.go` - Generated Go functions for all three queries
- `backend/internal/database/queries/models.go` - Updated with ReviewSchedule struct

## Decisions Made
- ListDueReviews joins exercises, topics, and subjects in a single query to return a rich row type (ListDueReviewsRow) — avoids N+1 lookups when rendering the due-items list in the frontend.
- UpsertReviewSchedule uses ON CONFLICT DO UPDATE — a single call handles both first-time scheduling and subsequent updates, making it safe to call unconditionally after each session ends.

## Deviations from Plan

None - plan executed exactly as written.

Migration could not be applied (PostgreSQL not running in this environment), but migration files are correct SQL and will be applied when the database is available.

## Issues Encountered
- PostgreSQL is not installed/running in this environment. Migration files were created and committed correctly; they will be applied at deployment time via the migrate CLI. All other artifacts (sqlc queries, generated Go code) were created and verified with `go build ./...`.

## User Setup Required
None - no external service configuration required beyond running the migration against the database.

## Next Phase Readiness
- review_schedule table DDL is ready for migration
- ReviewSchedule struct and query functions are available for use in Wave 2 integration (03-03 onward)
- SM-2 algorithm (03-01) + DB layer (03-02) complete — both Wave 1 artifacts ready for integration

---
*Phase: 03-spaced-repetition*
*Completed: 2026-04-03*
