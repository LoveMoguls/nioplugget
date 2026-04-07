---
phase: 08-so-amnen
plan: 03
subsystem: database
tags: [postgres, sql, migrations, seed-data, historia, np-prompts]

# Dependency graph
requires:
  - phase: 08-so-amnen
    provides: 08-02 — 12 Geografi NP-calibrated prompts in 007_seed_so_amnen.up.sql; 12 Historia SYSTEM_PROMPT_PLACEHOLDER entries remaining
  - phase: 07-no-amnen
    provides: Phase 6/7 canonical 8-section NP prompt template validated against real NP scoring

provides:
  - 12 NP-calibrated Historia system prompts across 4 topics (Industrialismens tid, De två världskrigen, Kalla kriget, 1900-talets politiska rörelser)
  - Complete 007_seed_so_amnen.up.sql with all 24 prompts — zero SYSTEM_PROMPT_PLACEHOLDER remaining
  - Migration 007 applied to database — 2 subjects, 8 topics, 24 exercises live
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns:
    - 8-section NP prompt template applied to all 4 Historia topics (same structure as Geografi/Kemi/Fysik)
    - Delprov A1/A2 framing for Historia (same as all prior subjects — E=A1 faktafrågor, C/A=A2 resonerande)
    - Scaffold-then-fill approach completed: Plan 01 scaffold, Plan 02 Geografi, Plan 03 Historia + migration

key-files:
  created: []
  modified:
    - backend/db/migrations/007_seed_so_amnen.up.sql

key-decisions:
  - "No apostrophes in Historia prompt text — Hitlers uppgång not Hitler's, Versaillesfredensmekanismer as compound noun to avoid possessive — consistent with established Phase 7 pattern"
  - "Migration run only after zero-placeholder confirmation (grep -c = 0) — anti-pattern from research avoided"

patterns-established:
  - "Pattern: scaffold-then-fill across 3 plans fully validated — Plan 01 scaffold, Plan 02 first-subject prompts, Plan 03 second-subject prompts + migration run"

requirements-completed: [HIST-01, HIST-02, HIST-03, HIST-04, HIST-05]

# Metrics
duration: 15min
completed: 2026-04-07
---

# Phase 8 Plan 03: Historia NP-Calibrated System Prompts + Migration Summary

**12 NP-calibrated Historia system prompts written across 4 topics using E/C/A verb hierarchy; migration 007 applied — 24 SO exercises live in database**

## Performance

- **Duration:** 15 min
- **Started:** 2026-04-07T20:50:00Z
- **Completed:** 2026-04-07T21:05:00Z
- **Tasks:** 2
- **Files modified:** 1

## Accomplishments
- Replaced all 12 remaining SYSTEM_PROMPT_PLACEHOLDER entries for 4 Historia topics (3 exercises each)
- All 12 Historia prompts follow the canonical 8-section template: ÄMNE, ÖVNING, NIVÅ, LÄRANDEMÅL, REGLER, NP-KOPPLING, BEDÖMNINGSLEDTRÅDAR, VANLIGA ELEVMISSAR, EXEMPELFRÅGOR
- E-nivå prompts (4) correctly use Beskriv/Vad kallas/Vad menas med verbs and reference Delprov A1 faktafrågor
- C-nivå prompts (4) correctly use Förklara sambandet/Förklara varför/Hur påverkas and reference Delprov A2 resonerande
- A-nivå prompts (4) correctly use Resonera kring/Ta ställning/Diskutera and reference Delprov A2 resonerande
- No apostrophes in any SQL string content — possessive forms written without apostrophes throughout
- Migration 007 applied successfully: `7/u seed_so_amnen (57ms)` — database now has 2 SO subjects live

## Task Commits

Each task was committed atomically:

1. **Task 1: Historia Industrialismens tid + De två världskrigen prompts** - `bf0eabe` (feat)
2. **Task 2: Historia Kalla kriget + Politiska rörelser prompts + migration run** - `bf3fa67` (feat)

**Plan metadata:** (docs commit follows)

## Files Created/Modified
- `backend/db/migrations/007_seed_so_amnen.up.sql` - All 24 NP-calibrated prompts complete; migration applied at version 7

## Decisions Made
- No apostrophes in Historia prompt text — "Hitlers uppgång" not "Hitler's uppgång", "Versaillesfredensmekanismer" as compound noun — avoids SQL escaping issues and follows established Phase 7 pattern
- Migration run after zero-placeholder confirmation (grep count = 0) — followed the anti-pattern rule from research exactly
- psql not available in environment — migration verification done via migrate version (7) and section count grep (24 per section) instead of psql SELECT queries

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- psql binary not installed on machine — post-migration database verification adapted: used `migrate version` to confirm version 7 applied, and grep counts to verify all 24 sections of each type present (NP-KOPPLING=24, EXEMPELFRÅGOR=24, BEDÖMNINGSLEDTRÅDAR=24, VANLIGA ELEVMISSAR=24). Migration ran cleanly with no SQL errors.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- Phase 8 fully complete: all 24 SO exercises (12 Geografi + 12 Historia) live in database with NP-calibrated prompts
- Migration 007 applied — students can navigate to Geografi or Historia, select any of 8 topic areas, and start exercise sessions
- No further seeding phases planned — v1.1 milestone complete with 6 subjects total (Biologi, Samhällskunskap, Matematik, Kemi, Fysik, Geografi, Historia = 7 subjects, 28 topics, 84 exercises)

## Self-Check: PASSED

- `backend/db/migrations/007_seed_so_amnen.up.sql` confirmed present and modified
- `grep -c SYSTEM_PROMPT_PLACEHOLDER` = 0 (all 24 exercises have full prompts)
- `grep -c "NP-KOPPLING"` = 24, `grep -c "EXEMPELFRÅGOR ATT STÄLLA"` = 24
- `grep -c "E-nivå (Delprov A1"` = 8, `grep -c "C-nivå (Delprov A2"` = 8, `grep -c "A-nivå (Delprov A2"` = 8
- Migration version confirmed at 7 via `migrate version`
- Commits bf0eabe and bf3fa67 confirmed in git log

---
*Phase: 08-so-amnen*
*Completed: 2026-04-07*
