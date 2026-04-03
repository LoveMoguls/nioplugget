---
gsd_state_version: 1.0
milestone: v1.0
milestone_name: milestone
status: unknown
last_updated: "2026-04-03T09:17:42.369Z"
progress:
  total_phases: 5
  completed_phases: 3
  total_plans: 16
  completed_plans: 15
---

# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-04-03)

**Core value:** Eleven lär sig genom dialog — AI:n ger aldrig direkta svar utan guidar eleven med ledande frågor
**Current focus:** Phase 5 — Polish + UI

## Current Position

Phase: 4 of 5 (Progress Views) — COMPLETE
Plan: 3 of 3 in current phase — COMPLETE
Status: Phase 4 complete — Progress views for students and parents
Last activity: 2026-04-03 — Completed Phase 4 (progress API, student progress page, parent child view)

Progress: [████████░░] 80% (Phase 4 complete)

## Performance Metrics

**Velocity:**
- Total plans completed: 0
- Average duration: —
- Total execution time: —

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| - | - | - | - |

**Recent Trend:**
- Last 5 plans: —
- Trend: —

*Updated after each plan completion*
| Phase 01-foundation P03 | 12min | 3 tasks | 8 files |
| Phase 01-foundation P04 | 5min | 2 tasks | 9 files |
| Phase 01-foundation P05 | 10 | 4 tasks | 20 files |
| Phase 03-spaced-repetition P02 | 2 | 2 tasks | 5 files |
| Phase 03-spaced-repetition P03 | 2 | 2 tasks | 7 files |

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- Init: BYOK model — parent stores own Claude API key (AES-256-GCM encrypted at rest)
- Init: Parent-child account model — parent manages API key and oversight, child studies
- Init: SM-2 spaced repetition for learning scheduling (SM-2 vs SM-2+ to be decided in Phase 3 planning)
- Init: Socratic output filter approach (second Claude call vs regex) to be decided in Phase 2 planning
- 01-01: Logger redacts at middleware level using header name denylist — no risk of accidentally logging secrets in handler code
- 01-01: pgxpool configured with max 25 conns, min 5, 30s health check — tuned for moderate load
- 01-01: sqlc with pgx/v5 driver and JSON tags — type-safe queries without ORM overhead
- 01-02: ParentQuerier interface over *queries.Queries — enables test mocking without a live database
- 01-02: Same 401 for unknown email and wrong password — prevents user enumeration attacks
- 01-02: httpOnly + Secure + SameSite=Lax JWT cookie — XSS cannot steal token, CSRF mitigated via Lax same-site
- [Phase 01-foundation]: APIKeyStore interface decouples handler from sqlc — enables unit testing without DB
- [Phase 01-foundation]: Raw key never logged or returned — only masked format (sk-ant...****) in responses
- [Phase 01-foundation]: QueriesStore adapter in store.go keeps handler independent of sqlc implementation details
- [Phase 01-foundation]: ChildQuerier interface over concrete *queries.Queries — enables unit tests without live database, consistent with ParentQuerier pattern
- [Phase 01-foundation]: UpdateStudentInvite via raw SQL in QueriesStore — sqlc lacks this query; passing queries.DBTX avoids touching generated files
- [Phase 01-foundation]: In-memory PINRateLimiter with sync.Mutex — DB-backed rate limiting not needed for MVP, simpler and fast for single-process deployment
- [Phase 01-foundation]: Flat SvelteKit routes instead of route groups — (parent)/login and (child)/login both resolve to /login causing fatal conflict
- [Phase 01-foundation]: GET /api/auth/me endpoint added to backend for session restoration from httpOnly cookie on frontend page load
- [Phase 01-foundation]: onMount auth guard in dashboard page instead of SvelteKit load function — SSR load runs before cookie-based store is hydrated
- [Phase 03-spaced-repetition 03-01]: Pure function with now time.Time parameter for SM-2 Calculate — enables deterministic testing without time.Now() side effects
- [Phase 03-spaced-repetition 03-01]: math.Max for EF floor enforcement (1.3) — simple, unambiguous, applied to both success and failure cases
- [Phase 03-spaced-repetition]: ListDueReviews joins exercises/topics/subjects for enriched due-item rows - avoids N+1 in due-item listing
- [Phase 03-spaced-repetition]: UpsertReviewSchedule uses ON CONFLICT DO UPDATE - idempotent after every session, no separate insert/update paths
- [Phase 03-spaced-repetition 03-03]: Goroutine for SM-2 update after EndSession — captures context+vars before launch, logs failure, never blocks response
- [Phase 03-spaced-repetition 03-03]: ChatStore extended with UpsertReviewSchedule/GetReviewSchedule — avoids circular import between srs and chat packages
- [Phase 03-spaced-repetition 03-03]: writeJSON/uuidToString helpers duplicated in srs/handler.go — circular import prevention over DRY

### Pending Todos

None yet.

### Blockers/Concerns

- Phase 2: Per-exercise system prompt design is the highest-value, hardest-to-get-right work. Needs deliberate prompt review before seeding ~50-60 exercises.
- Phase 2: Socratic guardrail implementation approach (output filter) needs a concrete decision before chat code is written.
- Phase 3: SM-2 vs SM-2+ (Blue Raja variant) decision needed before SRS implementation begins.
- General: GDPR data retention policy (e.g., "session messages deleted after 12 months") must be defined before launch.

## Session Continuity

Last session: 2026-04-03
Stopped at: Completed 03-03-PLAN.md — SRS integration (SM-2 hook, /api/reviews/due, review cards)
Resume file: None
