# Phase 3: Spaced Repetition - Research

**Researched:** 2026-04-03
**Status:** Complete

## SM-2 Algorithm

### Core Algorithm
The SM-2 (SuperMemo 2) algorithm calculates review intervals using:

1. **Ease Factor (EF)**: Starts at 2.5, adjusted after each review:
   ```
   EF' = EF + (0.1 - (5 - score) * (0.08 + (5 - score) * 0.02))
   ```
   Floor: 1.3 (per SRS-04, prevents "ease hell")

2. **Interval calculation**:
   - After 1st review: 1 day
   - After 2nd review: 6 days
   - After nth review: `previous_interval * EF`

3. **Score mapping** (1-5):
   - Score < 3: Reset interval to 1 (item not learned, restart)
   - Score >= 3: Progress to next interval

4. **Next review date**: `ended_at + interval days`

### Overdue Items
When a student reviews an overdue item:
- Use the actual elapsed time, not the scheduled interval
- If they score well on an overdue item, the long gap doesn't penalize them
- Standard SM-2 handles this naturally — the interval grows from current date regardless of how overdue

## Database Design

### New Table: `review_schedule`
```sql
CREATE TABLE review_schedule (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    exercise_id UUID NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    ease_factor REAL NOT NULL DEFAULT 2.5,
    interval_days INT NOT NULL DEFAULT 0,
    repetition_count INT NOT NULL DEFAULT 0,
    next_review TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_reviewed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(student_id, exercise_id)
);

CREATE INDEX idx_review_schedule_due ON review_schedule(student_id, next_review);
```

Key decisions:
- **UNIQUE(student_id, exercise_id)**: One schedule per student-exercise pair
- **Composite index on (student_id, next_review)**: Efficient "due now" queries
- **ease_factor defaults to 2.5**: SM-2 standard starting value
- **interval_days = 0**: First review happens immediately (first real interval calculated on first score)

### Integration with Existing Schema
- `sessions.score` already exists (1-5, written by Phase 2 scoring)
- `sessions.exercise_id` links to the exercise being reviewed
- After a session ends with a score, update/upsert `review_schedule` for that student+exercise

## Backend Architecture

### New Package: `backend/internal/srs/`
- `sm2.go` — Pure SM-2 algorithm (no DB dependencies, easy to test)
- `handler.go` — HTTP handler for GET /api/reviews/due
- `store.go` — ChatStore-like interface for review_schedule queries
- `hook.go` — Function called from EndSession to update review schedule

### SM-2 Pure Function
```go
type SM2Input struct {
    Score           int
    EaseFactor      float64
    IntervalDays    int
    RepetitionCount int
}

type SM2Output struct {
    EaseFactor      float64
    IntervalDays    int
    RepetitionCount int
    NextReview      time.Time
}

func Calculate(input SM2Input, now time.Time) SM2Output
```

### Integration Point: EndSession Hook
When `ChatHandler.EndSession` completes with a score, it should call the SRS upsert:
- If no `review_schedule` row exists for student+exercise: INSERT with calculated values
- If row exists: UPDATE with new SM-2 output
- Use PostgreSQL `ON CONFLICT ... DO UPDATE` for atomic upsert

### New API Endpoint
```
GET /api/reviews/due
```
- Auth: Child JWT required
- Returns exercises where `next_review <= NOW()` for the authenticated student
- Joins with exercises, topics, subjects to return display data
- Sorted by most overdue first (next_review ASC)
- Response includes topic name, subject name, days overdue

## Frontend Design

### Review Cards on /study Page
The `/study` page currently shows subject selection cards. Add a "Dags att repetera" section above the subjects:

```
[Review cards section - only visible when items are due]
  "Dags att repetera" heading
  Horizontal scroll or grid of review cards
  Each card: ämne + ämnesområde + "X dagar sedan" + "Repetera" button

[Existing subject selection grid below]
```

### Card Content
- Subject name (e.g., "Biologi")
- Topic name (e.g., "Ekologi")  
- Exercise title
- Time since due: "X dagar sedan" (calculated from next_review)
- "Repetera" button links to starting a new session on that exercise

### Empty State
When no items are due: Don't show the section at all (not an empty state message in the review area — the subject cards are already there as the primary action)

### UX Principles (from CONTEXT.md)
- Gentle nudge, not pressure
- No streak counters or shame mechanics
- "du har inte övat på detta pa ett tag" tone

## Existing Code to Modify

### Files Modified
1. **New migration**: `backend/db/migrations/004_review_schedule.up.sql` + `.down.sql`
2. **New sqlc queries**: `backend/db/queries/review_schedule.sql`
3. **Regenerate sqlc**: Updates `models.go` and generates `review_schedule.sql.go`
4. **New package**: `backend/internal/srs/` (sm2.go, handler.go, store.go, hook.go)
5. **Modify**: `backend/internal/chat/handler.go` — EndSession calls SRS upsert
6. **Modify**: `backend/internal/chat/store.go` — Add SRS store dependency or pass through
7. **Modify**: `backend/cmd/server/main.go` — Register new route, wire SRS handler
8. **Modify**: `frontend/src/lib/api.ts` — Add reviews API
9. **Modify**: `frontend/src/routes/study/+page.svelte` — Add review cards section

### Established Patterns to Follow
- sqlc for all DB queries (no raw SQL in handlers)
- Interface-based store pattern (ChatStore, ContentStore, etc.)
- Chi router with middleware groups
- `auth.ChildOnly` middleware for student-only routes
- shadcn-svelte Card component for UI cards
- `apiFetch` wrapper in api.ts

## Risk Assessment

### Low Risk
- SM-2 algorithm is well-documented and straightforward
- Database schema addition is non-breaking
- Frontend addition is additive (new section on existing page)

### Medium Risk
- EndSession hook integration — must not break existing session-end flow if SRS upsert fails (use fire-and-forget or log-and-continue)
- Query performance on review_schedule with many students — mitigated by composite index

### Mitigations
- SM-2 as pure function with unit tests — easy to verify correctness
- SRS upsert failure should not fail the EndSession response (log error, continue)
- EF floor of 1.3 enforced in algorithm, not just DB constraint

---

## RESEARCH COMPLETE
