# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-04-03)

**Core value:** Eleven lär sig genom dialog — AI:n ger aldrig direkta svar utan guidar eleven med ledande frågor
**Current focus:** Phase 1 — Foundation

## Current Position

Phase: 1 of 5 (Foundation)
Plan: 1 of 5 in current phase
Status: In progress
Last activity: 2026-04-03 — Completed 01-01 (Go scaffold, DB schema, sqlc, redacting logger)

Progress: [█░░░░░░░░░] 4%

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

### Pending Todos

None yet.

### Blockers/Concerns

- Phase 2: Per-exercise system prompt design is the highest-value, hardest-to-get-right work. Needs deliberate prompt review before seeding ~50-60 exercises.
- Phase 2: Socratic guardrail implementation approach (output filter) needs a concrete decision before chat code is written.
- Phase 3: SM-2 vs SM-2+ (Blue Raja variant) decision needed before SRS implementation begins.
- General: GDPR data retention policy (e.g., "session messages deleted after 12 months") must be defined before launch.

## Session Continuity

Last session: 2026-04-03
Stopped at: Completed 01-01-PLAN.md — Go backend scaffold, DB schema, sqlc, redacting logger
Resume file: None
