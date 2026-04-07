---
phase: 07-no-amnen
plan: 02
subsystem: database
tags: [sql, migrations, seed-data, kemi, system-prompts, np-calibrated]

requires:
  - phase: 07-no-amnen
    plan: 01
    provides: Migration scaffold 006 with Kemi and Fysik subjects, 8 topics, 24 exercise stubs

provides:
  - 12 NP-calibrated Kemi system prompts replacing SYSTEM_PROMPT_PLACEHOLDER
  - Full 8-section prompt structure for all 4 Kemi topics (materiens-uppbyggnad, kemiska-reaktioner, kemikalier-i-vardagen, kemi-och-hallbar-utveckling)

affects: [07-03-fysik-prompts]

tech-stack:
  added: []
  patterns:
    - "8-section NP prompt structure: ÄMNE, ÖVNING, NIVÅ, LÄRANDEMÅL, REGLER, NP-KOPPLING, BEDÖMNINGSLEDTRÅDAR, VANLIGA ELEVMISSAR, EXEMPELFRÅGOR"
    - "NP verb hierarchy per difficulty: E=Beskriv/Vad kallas/Vad menas med, C=Förklara sambandet/Förklara varför/Hur påverkas, A=Resonera kring/Ta ställning/Diskutera"

key-files:
  created: []
  modified:
    - backend/db/migrations/006_seed_no_amnen.up.sql

key-decisions:
  - "E-nivå prompts reference Delprov A1 (faktafrågor) — matches Skolverket NP structure"
  - "C and A-nivå prompts both reference Delprov A2 (resonerande frågor) — A goes deeper with independent reasoning requirement"
  - "Topic-specific VANLIGA ELEVMISSAR sourced from NP research (e.g., pH-skalan felriktad, förbränning utan syre, bioackumulering)"
  - "No single quotes used in Swedish prompt text — SQL syntax safety confirmed"

patterns-established:
  - "Kemi prompt quality standard: each prompt has topic-specific common mistakes from NP research, not generic chemistry mistakes"
  - "Difficulty escalation: E=name/recall, C=explain mechanism, A=independent reasoning with position and justification"

requirements-completed: [KEMI-01, KEMI-02, KEMI-03, KEMI-04, KEMI-05]

duration: 2min
completed: 2026-04-07
---

# Phase 7 Plan 02: Kemi NP-Calibrated System Prompts Summary

**12 NP-calibrated Kemi system prompts with 8-section structure, Skolverket verb hierarchy (E/C/A), and topic-specific common mistakes sourced from NP research — all SYSTEM_PROMPT_PLACEHOLDER replaced for Kemi exercises**

## Performance

- **Duration:** 2 min
- **Started:** 2026-04-07T17:15:51Z
- **Completed:** 2026-04-07T17:18:30Z
- **Tasks:** 2
- **Files modified:** 1

## Accomplishments
- Committed 6 pre-written NP-calibrated prompts for materiens-uppbyggnad and kemiska-reaktioner (Task 1)
- Wrote 6 new NP-calibrated prompts for kemikalier-i-vardagen and kemi-och-hallbar-utveckling (Task 2)
- All 12 Kemi prompts follow the canonical 8-section template with correct difficulty verb hierarchy
- SYSTEM_PROMPT_PLACEHOLDER count reduced from 24 to 12 (only Fysik placeholders remain)
- Verified no unescaped single quotes in SQL string literals

## Task Commits

1. **Task 1: Write Kemi Materiens uppbyggnad and Kemiska reaktioner prompts** - `dfc4d51` (feat)
2. **Task 2: Write Kemi Kemikalier i vardagen and Kemi och hållbar utveckling prompts** - `751c7d0` (feat)

## Files Created/Modified
- `backend/db/migrations/006_seed_no_amnen.up.sql` — 12 Kemi exercises with full NP-calibrated system prompts (E/C/A per topic)

## Decisions Made
- E-nivå prompts reference Delprov A1 (faktafrågor), C and A-nivå reference Delprov A2 (resonerande) — matches real Skolverket NP structure
- VANLIGA ELEVMISSAR uses topic-specific mistakes from NP research (pH-skalan riktning, förbränning utan syre, bioackumulering) not generic content
- A-nivå bedömningsledtråd explicitly requires independent reasoning with chemical mechanism — distinguishes from C-svar which explains but does not take position

## Deviations from Plan

None - plan executed exactly as written. The 6 prompts for Task 1 were already present as uncommitted modifications in the working tree; they were committed as Task 1 and verified against all quality criteria before proceeding to Task 2.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- All 12 Kemi system prompts complete — plan 03 (Fysik prompts) can proceed immediately
- SYSTEM_PROMPT_PLACEHOLDER count is exactly 12 (all Fysik topics) — confirmed by grep
- Kemi prompt quality pattern established: topic-specific common mistakes, correct Delprov references, NP verb hierarchy

---
*Phase: 07-no-amnen*
*Completed: 2026-04-07*
