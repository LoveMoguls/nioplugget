# Project Retrospective

*A living document updated after each milestone. Lessons feed forward into future planning.*

## Milestone: v1.0 — MVP

**Shipped:** 2026-04-06
**Phases:** 6 | **Plans:** 20 | **Timeline:** 3 days (2026-04-03 → 2026-04-06)

### What Was Built
- Full parent/child auth system with invite links, PIN login, AES-256-GCM API key encryption
- Socratic AI dialogue engine with SSE streaming — AI never gives direct answers
- 37 NP-calibrated exercise prompts across Biologi, Samhällskunskap, Matematik
- SM-2 spaced repetition with session scoring and "due today" cards
- Student progress view (per-subject cards, CSS-only bar charts, strengths/weaknesses)
- Parent oversight view (child progress overview, session history)
- Mobile-first UI with calm blue/green palette, dark mode, hamburger menu
- Landing page with hero, BYOK explanation, and FAQ

### What Worked
- **GSD wave-based parallel execution** — Wave 2 plans (04-02 + 04-03) ran in parallel successfully, saving time on independent frontend work
- **Phase 6 NP upgrade** — Adding a dedicated phase for NP-calibrated prompts after Phase 5 was the right call; prompt quality improved significantly
- **CSS-only charts** — No chart library dependency kept the frontend lean; worked well for progress bars
- **Goroutine for SM-2 update** — Clean async pattern that never blocks the session-end response

### What Was Inefficient
- **REQUIREMENTS.md drift** — Phase 2 and 5 requirements weren't checked off as phases completed, requiring manual correction at milestone completion. Should update checkboxes immediately after each phase executes.
- **ROADMAP.md plan checkboxes** — Many plan-level checkboxes in ROADMAP were unchecked despite phases being complete (documentation gap from prior session work)
- **Phase 4 executed but undocumented** — Core implementation existed in `e472b71` (prior session) but phase had no SUMMARY.md. The re-execution pass mostly just documented existing work.

### Patterns Established
- **BYOK + parent-child model** works cleanly for this use case — carry forward to similar apps
- **NP verb hierarchy in prompts** (Beskriv=E, Förklara=C, Resonera=A) is a strong pattern for Swedish educational AI
- **Per-phase store interface** (e.g., `QueriesStore`, `ProgressStore`) prevents circular imports and enables unit testing — replicate in future Go phases
- **Circular import prevention via helper duplication** — `writeJSON`/`uuidToString` helpers duplicated per package; documented trade-off, consistent pattern

### Key Lessons
1. **Update requirement checkboxes after each phase** — don't let REQUIREMENTS.md drift; the gap compounds and creates noise at milestone completion
2. **Document phase work even when implementation precedes planning** — if code was written in a prior session, still run the execute-phase workflow to create SUMMARY.md and link the commits
3. **Phase 6 added organically worked well** — inserting a prompt-quality phase after Phase 5 (based on user feedback about NP calibration) shows the roadmap should stay flexible through v1

### Cost Observations
- Model mix: sonnet throughout (balanced profile)
- Sessions: 3-4 sessions over 3 days
- Notable: Most implementation work happened in prior sessions; GSD execution passes were fast (documentation + verification)

---

## Cross-Milestone Trends

| Milestone | Phases | Plans | Days | LOC |
|-----------|--------|-------|------|-----|
| v1.0 MVP | 6 | 20 | 3 | ~10,400 |
