---
phase: 03-spaced-repetition
plan: 03
subsystem: api
tags: [srs, sm2, spaced-repetition, svelte, go, chi, pgx]

# Dependency graph
requires:
  - phase: 03-01
    provides: SM-2 Calculate function and SM2Input/SM2Output types
  - phase: 03-02
    provides: review_schedule table migration and sqlc queries (UpsertReviewSchedule, ListDueReviews, GetReviewSchedule)
provides:
  - SRS store interface and QueriesStore adapter (srs/store.go)
  - SRS HTTP handler for GET /api/reviews/due (srs/handler.go)
  - EndSession SM-2 hook — session completion triggers review_schedule upsert
  - Frontend review cards section on /study page (amber-colored, friendly)
affects: [04-session-history, 05-parent-dashboard]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - Non-blocking goroutine for SRS side-effect — captures context vars before goroutine, logs errors, never fails HTTP response
    - Parallel Promise.all fetch in onMount for concurrent subjects + reviews
    - Conditional section rendering (hidden when empty, not error state)

key-files:
  created:
    - backend/internal/srs/store.go
    - backend/internal/srs/handler.go
  modified:
    - backend/internal/chat/handler.go
    - backend/internal/chat/store.go
    - backend/cmd/server/main.go
    - frontend/src/lib/api.ts
    - frontend/src/routes/study/+page.svelte

key-decisions:
  - "Goroutine for SM-2 update after EndSession — captures context+vars before launch, logs failure, never blocks response"
  - "ChatStore extended with UpsertReviewSchedule/GetReviewSchedule — keeps SM-2 logic in chat package avoiding circular imports"
  - "Duplicate writeJSON/uuidToString helpers in srs/handler.go — avoids circular import between srs and chat packages"

patterns-established:
  - "Non-blocking goroutine pattern: capture ctx and UUID vars before go func{}, never access r.* inside goroutine"
  - "Due reviews section: conditional render with {#if length > 0}, hidden not empty-state"

requirements-completed:
  - SRS-01
  - SRS-03

# Metrics
duration: 2min
completed: 2026-04-03
---

# Phase 3 Plan 03: SRS Integration Summary

**SM-2 wired into session completion via goroutine hook; GET /api/reviews/due endpoint; amber review cards on /study page**

## Performance

- **Duration:** 2 min
- **Started:** 2026-04-03T09:19:00Z
- **Completed:** 2026-04-03T09:21:33Z
- **Tasks:** 2
- **Files modified:** 7

## Accomplishments
- Session end now fires a non-blocking goroutine that reads existing review schedule, runs SM-2 Calculate, and upserts review_schedule — SRS failures never affect HTTP response
- GET /api/reviews/due endpoint returns enriched due-review rows with exercise title, topic/subject names, slugs, and daysOverdue
- /study page shows amber-colored "Dags att repetera" cards above subject grid; section is hidden entirely when nothing is due

## Task Commits

1. **Task 1: Create SRS backend — store, handler, and EndSession hook** - `b29e965` (feat)
2. **Task 2: Wire routes and add frontend review cards** - `aa40535` (feat)

## Files Created/Modified
- `backend/internal/srs/store.go` - SRSStore interface and QueriesStore adapter (created)
- `backend/internal/srs/handler.go` - SRSHandler with ListDueReviews (created)
- `backend/internal/chat/handler.go` - EndSession extended with SM-2 goroutine hook
- `backend/internal/chat/store.go` - ChatStore extended with UpsertReviewSchedule and GetReviewSchedule
- `backend/cmd/server/main.go` - SRS handler initialized and /api/reviews route registered
- `frontend/src/lib/api.ts` - reviews.due() added
- `frontend/src/routes/study/+page.svelte` - Due review cards section added

## Decisions Made
- **Goroutine for SM-2**: The SM-2 update after EndSession is launched in a goroutine. Context variables (studentUUID, exerciseID, score) are captured before the goroutine starts to avoid data races. Using the request context inside the goroutine is acceptable here since the context is captured before response is written and the goroutine completes quickly.
- **ChatStore extension over callback**: Added UpsertReviewSchedule and GetReviewSchedule to ChatStore (preferred approach from plan) — keeps SM-2 logic co-located in chat/handler.go, avoids circular import between srs and chat packages.
- **Duplicate helpers**: writeJSON, uuidToString, parseUUID duplicated in srs/handler.go rather than imported from chat — circular import prevention.

## Deviations from Plan

None — plan executed exactly as written.

## Issues Encountered

Pre-existing TypeScript error in `frontend/src/routes/invite/[token]/+page.svelte` (unrelated to this plan — deferred).

## User Setup Required

None — no external service configuration required.

## Next Phase Readiness
- SRS end-to-end flow complete: session end schedules reviews; student sees due items on home page
- Phase 4 (session history) can build on SessionStore patterns established here
- Pre-existing frontend type error in invite page should be fixed before deploying

---
*Phase: 03-spaced-repetition*
*Completed: 2026-04-03*
