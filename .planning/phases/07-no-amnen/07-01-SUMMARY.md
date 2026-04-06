---
phase: 07-no-amnen
plan: 01
subsystem: database
tags: [sql, migrations, seed-data, kemi, fysik]

requires:
  - phase: 06-matematik
    provides: Established subject+topic+exercise migration pattern (CROSS JOIN VALUES)
provides:
  - Migration scaffold 006 with Kemi and Fysik subjects, 8 topics, 24 exercise stubs
affects: [07-02-kemi-prompts, 07-03-fysik-prompts]

tech-stack:
  added: []
  patterns: ["CROSS JOIN VALUES for multi-row inserts keyed by subject slug"]

key-files:
  created:
    - backend/db/migrations/006_seed_no_amnen.up.sql
    - backend/db/migrations/006_seed_no_amnen.down.sql
  modified: []

key-decisions:
  - "Used SYSTEM_PROMPT_PLACEHOLDER for all 24 exercises — plans 02 and 03 fill in NP-calibrated content"
  - "display_order=4 for Kemi, display_order=5 for Fysik — continues existing sequence"
  - "All slugs URL-safe: ä→a, å→a, ö→o, space→- (e.g. kemi-och-hallbar-utveckling, krafter-och-rorelse)"

patterns-established:
  - "Migration scaffold pattern: subjects + topics + exercise stubs in one file, prompts filled later"

requirements-completed: [KEMI-01, KEMI-02, KEMI-03, KEMI-04, FYS-01, FYS-02, FYS-03, FYS-04]

duration: 1min
completed: 2026-04-06
---

# Phase 7 Plan 01: Migration Scaffold for Kemi and Fysik Summary

**SQL migration scaffold with Kemi (display_order=4) and Fysik (display_order=5) subjects, 8 URL-safe-slugged topics, and 24 exercise stubs awaiting NP-calibrated system prompts in plans 02 and 03**

## Performance

- **Duration:** 1 min
- **Started:** 2026-04-06T18:53:19Z
- **Completed:** 2026-04-06T18:54:24Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Created `006_seed_no_amnen.down.sql` — CASCADE delete for kemi and fysik subjects
- Created `006_seed_no_amnen.up.sql` — 133 lines with 2 subjects, 8 topics (4 Kemi + 4 Fysik), 24 exercise stubs
- Verified 24 × `SYSTEM_PROMPT_PLACEHOLDER` entries (one per exercise row) — confirmed by grep -c

## Task Commits

1. **Task 1: Write down migration** - `203abe5` (feat)
2. **Task 2: Write up migration scaffold** - `50b9cde` (feat)

## Files Created/Modified
- `backend/db/migrations/006_seed_no_amnen.down.sql` — DELETE kemi+fysik, CASCADE removes topics+exercises
- `backend/db/migrations/006_seed_no_amnen.up.sql` — Full migration scaffold: subjects, topics, 24 exercise stubs

## Decisions Made
- SYSTEM_PROMPT_PLACEHOLDER used for all 24 exercises so plans 02 and 03 can focus solely on NP-calibrated prompt quality without structural changes
- Slug transliteration confirmed: `kemi-och-hallbar-utveckling` (not hållbar), `krafter-och-rorelse` (not rörelse)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Migration scaffold complete — plans 02 (Kemi prompts) and 03 (Fysik prompts) can proceed
- Both plans only need to replace `SYSTEM_PROMPT_PLACEHOLDER` with NP-calibrated content for their respective subjects

---
*Phase: 07-no-amnen*
*Completed: 2026-04-06*
