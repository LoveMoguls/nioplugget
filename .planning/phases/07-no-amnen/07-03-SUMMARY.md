---
phase: 07-no-amnen
plan: 03
subsystem: database
tags: [sql, migrations, seed-data, fysik, system-prompts, np-calibrated]

requires:
  - phase: 07-no-amnen
    plan: 02
    provides: 12 NP-calibrated Kemi system prompts replacing SYSTEM_PROMPT_PLACEHOLDER

provides:
  - 12 NP-calibrated Fysik system prompts replacing SYSTEM_PROMPT_PLACEHOLDER
  - Full 8-section prompt structure for all 4 Fysik topics (krafter-och-rorelse, elektricitet-och-magnetism, energi-och-energiomvandlingar, astronomi-och-universum)
  - Complete migration 006 applied to database — 2 subjects, 8 topics, 24 exercises

affects: []

tech-stack:
  added: []
  patterns:
    - "8-section NP prompt structure: ÄMNE, ÖVNING, NIVÅ, LÄRANDEMÅL, REGLER, NP-KOPPLING, BEDÖMNINGSLEDTRÅDAR, VANLIGA ELEVMISSAR, EXEMPELFRÅGOR"
    - "NP verb hierarchy per difficulty: E=Beskriv/Vad menas med/Nämn, C=Förklara sambandet/Förklara varför/Hur påverkas, A=Resonera kring/Ta ställning/Diskutera"

key-files:
  created: []
  modified:
    - backend/db/migrations/006_seed_no_amnen.up.sql

key-decisions:
  - "E-nivå Fysik prompts reference Delprov A1 (faktafrågor), C and A-nivå reference Delprov A2 (resonerande) — consistent with Kemi pattern"
  - "No apostrophes used in Swedish prompt text — Ohms lag written without apostrophe to avoid SQL escaping issues"
  - "Migration run confirms complete SQL file is syntactically valid — golang-migrate applied migration 006 successfully"
  - "Fysik VANLIGA ELEVMISSAR sourced from NP research: energy conservation misconception (energin försvinner), Newtons 3:e lag misunderstanding, massa vs vikt confusion"

patterns-established:
  - "Fysik prompt quality matches Kemi standard: topic-specific NP research-backed common mistakes, not generic physics errors"
  - "All 4 Fysik topics follow difficulty escalation: E=name/recall, C=explain mechanism/relationship, A=independent reasoning with position and justification"

requirements-completed: [FYS-01, FYS-02, FYS-03, FYS-04, FYS-05]

duration: 4min
completed: 2026-04-07
---

# Phase 7 Plan 03: Fysik NP-Calibrated System Prompts and Migration Run Summary

**12 NP-calibrated Fysik system prompts with 8-section structure, Skolverket verb hierarchy (E/C/A), and topic-specific common mistakes — all SYSTEM_PROMPT_PLACEHOLDER replaced; migration 006 applied with 2 subjects, 8 topics, 24 exercises**

## Performance

- **Duration:** 4 min
- **Started:** 2026-04-07T17:21:00Z
- **Completed:** 2026-04-07T17:25:58Z
- **Tasks:** 2
- **Files modified:** 1

## Accomplishments
- Wrote 6 NP-calibrated prompts for krafter-och-rorelse and elektricitet-och-magnetism (Task 1)
- Wrote 6 NP-calibrated prompts for energi-och-energiomvandlingar and astronomi-och-universum (Task 2)
- All 12 Fysik prompts follow the canonical 8-section template with correct difficulty verb hierarchy
- SYSTEM_PROMPT_PLACEHOLDER count reduced from 12 to 0 — migration file complete
- Migration 006 successfully applied: `6/u seed_no_amnen (100.052828ms)`
- Verified all 24 prompts contain all 8 required sections (each section count = 24)
- NP verb hierarchy confirmed: E=Beskriv/Vad menas med (48 instances), C=Förklara sambandet/Hur påverkas (21 instances), A=Resonera kring/Ta ställning/Diskutera (44 instances)

## Task Commits

1. **Task 1: Write Fysik Krafter och rörelse and Elektricitet och magnetism prompts** - `171808a` (feat)
2. **Task 2: Write Fysik Energi och Astronomi prompts; run migration 006** - `3c58b4e` (feat)

## Files Created/Modified
- `backend/db/migrations/006_seed_no_amnen.up.sql` — all 24 exercises have full NP-calibrated system prompts; migration applied to database

## Decisions Made
- E-nivå prompts reference Delprov A1 (faktafrågor), C and A-nivå reference Delprov A2 (resonerande) — consistent with Kemi pattern from plan 02
- No apostrophes in any prompt text — Ohms lag written without possessive to avoid SQL escaping issues
- VANLIGA ELEVMISSAR uses topic-specific physics mistakes from NP research: energy conservation confusion, Newtons 3rd law action-reaction misunderstanding, massa vs vikt (mass vs weight in N), Big Bang as explosion vs expansion

## Deviations from Plan

None - plan executed exactly as written. All 12 Fysik prompts written across 2 tasks following the exact NP template structure. Migration ran successfully on first attempt.

## Issues Encountered

None. psql is not installed on this system, but the golang-migrate CLI confirmed successful migration via exit code 0 and output `6/u seed_no_amnen`.

## User Setup Required

None - migration has been applied to the local development database.

## Next Phase Readiness
- Phase 7 complete — all 24 NP-calibrated exercises for Kemi and Fysik are in the database
- A student can navigate to Kemi or Fysik, select any of the 8 topic areas, and start an exercise session
- Phase 8 (SO-ämnen: Geografi and Historia) can proceed using the same migration and prompt patterns

---
*Phase: 07-no-amnen*
*Completed: 2026-04-07*
