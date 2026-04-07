---
phase: 08-so-amnen
plan: 01
subsystem: database
tags: [postgres, sql, migrations, seed-data, geografi, historia]

# Dependency graph
requires:
  - phase: 07-no-amnen
    provides: Migration 006 pattern — CROSS JOIN VALUES scaffold for subjects, topics, and exercise stubs
provides:
  - 007_seed_so_amnen.up.sql — subjects Geografi (display_order=6) and Historia (display_order=7), 8 topics, 24 exercise stubs with SYSTEM_PROMPT_PLACEHOLDER
  - 007_seed_so_amnen.down.sql — CASCADE delete for geografi and historia slugs
affects: [08-02-so-amnen, 08-03-so-amnen]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - CROSS JOIN VALUES scaffold pattern for multi-topic multi-exercise seed migrations
    - Scaffold-then-fill: SYSTEM_PROMPT_PLACEHOLDER stubs allow structural commit before content work

key-files:
  created:
    - backend/db/migrations/007_seed_so_amnen.up.sql
    - backend/db/migrations/007_seed_so_amnen.down.sql
  modified: []

key-decisions:
  - "Geografi assigned display_order=6, Historia display_order=7 — continues established sequence (Biologi=1, Samhällskunskap=2, Matematik=3, Kemi=4, Fysik=5)"
  - "All 8 topic slugs transliterated to URL-safe forms (å→a, ä→a, ö→o, space→hyphen)"
  - "24 SYSTEM_PROMPT_PLACEHOLDER entries — one per exercise — to be replaced in plans 02 and 03"

patterns-established:
  - "Pattern: scaffold-then-fill — structural SQL committed separately from prompt content for cleaner review"

requirements-completed: [GEO-01, GEO-02, GEO-03, GEO-04, HIST-01, HIST-02, HIST-03, HIST-04]

# Metrics
duration: 5min
completed: 2026-04-07
---

# Phase 8 Plan 01: SO-ämnen Migration Scaffold Summary

**SQL seed migration scaffold for Geografi and Historia — 2 subjects, 8 topics, 24 exercise stubs with SYSTEM_PROMPT_PLACEHOLDER ready for prompt fill-in**

## Performance

- **Duration:** 5 min
- **Started:** 2026-04-07T17:29:31Z
- **Completed:** 2026-04-07T17:34:00Z
- **Tasks:** 2
- **Files modified:** 2

## Accomplishments
- Created 007_seed_so_amnen.down.sql with CASCADE delete for 'geografi' and 'historia' slugs
- Created 007_seed_so_amnen.up.sql inserting 2 subjects (display_order=6,7), 8 topics, and 24 exercise stubs
- All topic slugs verified URL-safe (no å/ä/ö in slug values); exactly 24 SYSTEM_PROMPT_PLACEHOLDER entries confirmed

## Task Commits

Each task was committed atomically:

1. **Task 1: Write down migration** - `3820ebb` (feat)
2. **Task 2: Write up migration scaffold** - `f748b25` (feat)

**Plan metadata:** (docs commit follows)

## Files Created/Modified
- `backend/db/migrations/007_seed_so_amnen.down.sql` - DELETE statement for 'geografi' and 'historia' with CASCADE
- `backend/db/migrations/007_seed_so_amnen.up.sql` - Full scaffold: subjects, 8 topics, 24 exercise stubs

## Decisions Made
- Geografi=6, Historia=7 for display_order — continues established subject sequence from Phase 7
- All Swedish special characters transliterated in slugs per established convention
- SYSTEM_PROMPT_PLACEHOLDER pattern identical to Phase 7 approach — plans 02 and 03 will replace all 24 entries before migration is run

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Migration scaffold structure is complete and committed
- Plan 02 can immediately begin filling in Geografi system prompts (12 exercises)
- Plan 03 will fill Historia prompts (12 exercises) and run the migration
- CRITICAL: Do not run migration until grep -c SYSTEM_PROMPT_PLACEHOLDER 007_seed_so_amnen.up.sql returns 0

## Self-Check: PASSED

All files confirmed present and commits verified.

---
*Phase: 08-so-amnen*
*Completed: 2026-04-07*
