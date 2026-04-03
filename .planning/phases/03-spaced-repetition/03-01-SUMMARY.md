---
phase: 03-spaced-repetition
plan: 01
subsystem: algorithm
tags: [go, sm2, spaced-repetition, scheduling, tdd]

# Dependency graph
requires: []
provides:
  - Pure SM-2 algorithm as Calculate(SM2Input, time.Time) SM2Output
  - SM2Input and SM2Output types in package srs
  - Comprehensive table-driven tests covering all score values and edge cases
affects: [03-spaced-repetition]

# Tech tracking
tech-stack:
  added: []
  patterns: [pure-function algorithm, table-driven tests, TDD red-green]

key-files:
  created:
    - backend/internal/srs/sm2.go
    - backend/internal/srs/sm2_test.go
  modified: []

key-decisions:
  - "Standard SM-2 algorithm with ease factor floor at 1.3 per spec"
  - "Pure function with no DB dependencies - easy to test all edge cases before wiring to database"
  - "math.Round for interval calculation and math.Max for EF floor enforcement"

patterns-established:
  - "Pure algorithm functions in backend/internal/srs/ package"
  - "Table-driven Go tests with fixed 'now' time for deterministic NextReview assertions"

requirements-completed: [SRS-02, SRS-04]

# Metrics
duration: 8min
completed: 2026-04-03
---

# Phase 3 Plan 01: SM-2 Spaced Repetition Algorithm Summary

**Pure SM-2 scheduling algorithm in Go with 13 tests covering interval progression (1->6->n*EF), EF floor at 1.3, and failure reset behavior**

## Performance

- **Duration:** 8 min
- **Started:** 2026-04-03T11:10:00Z
- **Completed:** 2026-04-03T11:18:00Z
- **Tasks:** 2 (TDD: RED + GREEN)
- **Files modified:** 2

## Accomplishments
- SM-2 Calculate function implemented as pure Go function with no DB dependencies
- All 13 tests pass: 10 table-driven cases + EF floor invariant + interval progression test
- Ease factor formula correctly implemented with floor enforced via math.Max
- Interval schedule verified: 1 day (first) -> 6 days (second) -> round(prev*EF) (subsequent)

## Task Commits

Each task was committed atomically:

1. **RED: Failing SM-2 tests** - `e612632` (test)
2. **GREEN: SM-2 implementation** - `459adc9` (feat)

_Note: TDD plan - RED commit before implementation, GREEN commit after all tests pass_

## Files Created/Modified
- `backend/internal/srs/sm2.go` - SM2Input/SM2Output types and Calculate function
- `backend/internal/srs/sm2_test.go` - Table-driven tests with edge cases and invariant checks

## Decisions Made
- Used `math.Round` for interval calculation to match SM-2 spec (round half-up)
- Used `math.Max(newEF, 1.3)` for floor enforcement - simple and unambiguous
- EF formula applied on both success and failure (only floor differs)
- Pure function with `now time.Time` parameter enables deterministic testing

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- SM-2 algorithm ready to be wired to the database via review_schedule table
- Next plans can call `srs.Calculate()` with state loaded from DB and save results back
- All edge cases verified; implementation is production-ready

---
*Phase: 03-spaced-repetition*
*Completed: 2026-04-03*
