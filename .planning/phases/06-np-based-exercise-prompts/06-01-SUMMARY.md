# Phase 06 Plan 01 Summary: NP-Calibrated Exercise Prompts

**Status:** Complete
**Duration:** Single execution
**Files modified:** 2

## What Was Built

Rewrote all 37 exercise system prompts to match real Swedish nationella prov (NP) patterns. Created a new database migration (005) that UPDATEs every exercise's `system_prompt` value.

## Key Changes

### Migration 005 (up)
- 37 UPDATE statements covering all exercises across 3 subjects (Biologi, Samhällskunskap, Matematik), 12 topics
- Each prompt now includes:
  - NP-KOPPLING section explaining which delprov tests this content
  - BEDÖMNINGSLEDTRÅDAR for the AI tutor calibrated to E/C/A level
  - VANLIGA ELEVMISSAR from NP research
  - 3-5 NP-style example questions using authentic verb patterns
  - Preserved Socratic constraint ("Ge ALDRIG direkta svar")

### Migration 005 (down)
- 37 UPDATE statements restoring all original prompts from 003_seed_exercises.up.sql

## NP Verb Hierarchy Implementation

| Difficulty | NP Level | Primary Verbs Used |
|-----------|----------|-------------------|
| 1 | E-nivå | Beskriv, Vad kallas, Vad menas med, Nämn |
| 2 | C-nivå | Förklara sambandet, Förklara varför, Hur påverkas |
| 3+ | A-nivå | Resonera kring, Diskutera, Ta ställning, Motivera |

## Subject-Specific NP Patterns

- **Biologi:** References Delprov A1 (faktafrågor) and A2 (resonerande frågor). Addresses evolution misconceptions, mechanism gaps, atom/material confusion.
- **Samhällskunskap:** "Resonera" as primary verb. Quantity scaffolding ("ge 1-2 exempel"). Source-critical thinking. Fördelar/nackdelar balance requirement.
- **Matematik:** References Delprov B/C/D. "Visa hur du tänker" throughout. PRIM-gruppens bedömningsprincip. Real-world word problems with names (Lena, Ali, Maria).

## Files

| File | Action |
|------|--------|
| `backend/db/migrations/005_update_exercise_prompts.up.sql` | Created — 37 NP-calibrated UPDATE statements |
| `backend/db/migrations/005_update_exercise_prompts.down.sql` | Created — 37 rollback UPDATE statements |

## Decisions Made

- Used UPDATE by title+topic+subject join rather than by UUID (UUIDs are generated at insert time)
- Kept consistent prompt structure across all exercises: ÄMNE, ÖVNING, NIVÅ, LÄRANDEMÅL, REGLER, NP-KOPPLING, BEDÖMNINGSLEDTRÅDAR, VANLIGA ELEVMISSAR, EXEMPELFRÅGOR
- Mapped difficulty_order to NP levels: 1→E, 2→C, 3+→A
- Bio difficulty 4 (Människans påverkan) mapped to A-level since it requires multi-step reasoning
