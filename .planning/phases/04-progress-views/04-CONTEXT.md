# Phase 4: Progress Views - Context

**Gathered:** 2026-04-03
**Status:** Ready for planning

<domain>
## Phase Boundary

Students and parents can view progress data. Read-only layer over accumulated session data. Covers PROG-01..03.

</domain>

<decisions>
## Implementation Decisions

### Student progress view
- Per-subject cards (Bio/Samhälle/Matte) showing: antal genomförda pass, snittbetyg, svagheter
- Strengths/weaknesses shown via color-coded topic areas: green (4-5), yellow (3), red (1-2)
- Simple bar charts showing score per topic area

### Parent overview
- Detailed view: summary stats (antal pass senaste veckan, snittbetyg per ämne) + full list of completed sessions with date and score
- No chat history visible to parents — only scores and topics
- Accessible from parent dashboard (/dashboard/child/:id)

### Data visualization
- Simple bar charts — score per topic area per subject
- Color-coded (green/yellow/red) topic indicators
- No complex graphs or trend lines in v1

### Claude's Discretion
- Chart library choice (lightweight, SSR-compatible)
- Exact card layout and spacing
- How to handle topics with no sessions yet (grey/dimmed?)

</decisions>

<code_context>
## Existing Code Insights

### Reusable Assets
- `backend/internal/database/queries/` — sessions, review_schedule queries
- `frontend/src/lib/components/` — shadcn-svelte Card, etc.
- `frontend/src/routes/(parent)/dashboard/` — parent dashboard already exists

### Established Patterns
- Chi handlers returning JSON aggregates
- shadcn-svelte components
- API client in frontend/src/lib/api.ts

### Integration Points
- New API: GET /api/progress/:studentId — aggregated progress data
- New API: GET /api/progress/:studentId/sessions — session history list
- Frontend: /progress route (student), /dashboard/child/:id enhanced (parent)

</code_context>

<specifics>
## Specific Ideas

- Progress should feel informative, not judgmental — no rankings or comparisons
- Color coding is gentle (muted green/yellow/red, not traffic-light bright)

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 04-progress-views*
*Context gathered: 2026-04-03*
