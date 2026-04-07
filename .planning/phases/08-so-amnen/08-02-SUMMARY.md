---
phase: 08-so-amnen
plan: 02
subsystem: database
tags: [postgres, sql, migrations, seed-data, geografi, np-prompts]

# Dependency graph
requires:
  - phase: 08-so-amnen
    provides: 08-01 — 007_seed_so_amnen.up.sql scaffold with 24 SYSTEM_PROMPT_PLACEHOLDER entries for Geografi and Historia
  - phase: 07-no-amnen
    provides: Phase 6/7 canonical 8-section NP prompt template validated against real NP scoring

provides:
  - 12 NP-calibrated Geografi system prompts in 007_seed_so_amnen.up.sql replacing SYSTEM_PROMPT_PLACEHOLDER for all 4 Geografi topics
  - E/C/A verb hierarchy correctly applied across all 12 Geografi exercises
  - Topic-specific VANLIGA ELEVMISSAR and EXEMPELFRÅGOR for each of the 4 Geografi topics
affects: [08-03-so-amnen]

# Tech tracking
tech-stack:
  added: []
  patterns:
    - 8-section NP prompt template applied to Geografi subject area (same structure as Kemi/Fysik in Phase 7)
    - Delprov A1/A2 framing for SO subjects (same as NO subjects — E=A1 faktafrågor, C/A=A2 resonerande)

key-files:
  created: []
  modified:
    - backend/db/migrations/007_seed_so_amnen.up.sql

key-decisions:
  - "Delprov A1/A2 framing used for Geografi — same as Kemi/Fysik in Phase 7 (consistent NP structure across all subjects)"
  - "No apostrophes in any Geografi prompt text — FNs not FN's, SDG-målen not SDG's — avoids SQL escaping issues"
  - "12 Historia SYSTEM_PROMPT_PLACEHOLDER entries intentionally left for Plan 03"

patterns-established:
  - "Pattern: scaffold-then-fill across plans — Plan 01 scaffold, Plan 02 Geografi prompts, Plan 03 Historia prompts + migration run"

requirements-completed: [GEO-01, GEO-02, GEO-03, GEO-04, GEO-05]

# Metrics
duration: 12min
completed: 2026-04-07
---

# Phase 8 Plan 02: Geografi NP-Calibrated System Prompts Summary

**12 NP-calibrated Geografi system prompts written across 4 topics (Befolkning, Klimat, Naturresurser, Geopolitik) using E/C/A verb hierarchy and Delprov A1/A2 structure**

## Performance

- **Duration:** 12 min
- **Started:** 2026-04-07T20:35:00Z
- **Completed:** 2026-04-07T20:47:00Z
- **Tasks:** 2
- **Files modified:** 1

## Accomplishments
- Replaced 12 SYSTEM_PROMPT_PLACEHOLDER entries for all 4 Geografi topics (3 exercises each)
- All 12 prompts follow the canonical 8-section template: ÄMNE, ÖVNING, NIVÅ, LÄRANDEMÅL, REGLER, NP-KOPPLING, BEDÖMNINGSLEDTRÅDAR, VANLIGA ELEVMISSAR, EXEMPELFRÅGOR
- E-nivå prompts (4) correctly use Beskriv/Vad kallas/Vad menas med verbs and reference Delprov A1 faktafrågor
- C-nivå prompts (4) correctly use Förklara sambandet/Förklara varför/Hur påverkas and reference Delprov A2 resonerande
- A-nivå prompts (4) correctly use Resonera kring/Ta ställning/Diskutera and reference Delprov A2 resonerande
- No apostrophes in any SQL string content; 12 Historia placeholders remain for Plan 03

## Task Commits

Each task was committed atomically:

1. **Task 1: Befolkning och urbanisering + Klimat och klimatförändringar prompts** - `f86a8a6` (feat)
2. **Task 2: Naturresurser och hållbarhet + Geopolitik och handel prompts** - `d903b92` (feat)

**Plan metadata:** (docs commit follows)

## Files Created/Modified
- `backend/db/migrations/007_seed_so_amnen.up.sql` - 12 SYSTEM_PROMPT_PLACEHOLDER entries replaced with full NP-calibrated prompts for all Geografi exercises

## Decisions Made
- Delprov A1/A2 framing applied to Geografi (same as Kemi/Fysik) — consistent NP structure across all subjects in the system
- No apostrophes in prompt text — possessive forms written without apostrophes (FNs, SDG-målen, Brundtland-definitionen) to avoid SQL escaping issues
- 12 Historia placeholders left intentionally — Plan 03 handles Historia prompts and migration execution

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None.

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- All 12 Geografi prompts are NP-calibrated and complete
- Plan 03 can immediately begin writing Historia prompts (12 exercises across 4 topics)
- CRITICAL: Do not run migration until `grep -c SYSTEM_PROMPT_PLACEHOLDER 007_seed_so_amnen.up.sql` returns 0 (currently 12 — Historia placeholders remain)

## Self-Check: PASSED

- `backend/db/migrations/007_seed_so_amnen.up.sql` confirmed present and modified
- `grep -c SYSTEM_PROMPT_PLACEHOLDER` = 12 (correct: all Geografi done, Historia pending)
- All 12 NP-KOPPLING, BEDÖMNINGSLEDTRÅDAR, VANLIGA ELEVMISSAR, EXEMPELFRÅGOR sections verified present
- Commits f86a8a6 and d903b92 confirmed in git log

---
*Phase: 08-so-amnen*
*Completed: 2026-04-07*
