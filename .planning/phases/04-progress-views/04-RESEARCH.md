# Phase 4: Progress Views — Research

**Researched:** 2026-04-03
**Status:** Complete

## Phase Goal

Students can see their own strengths and weaknesses per subject, and parents can see an overview of their child's activity — both derived purely from existing session data.

## Requirements

- **PROG-01**: Elev kan se en översikt av sin progress per ämne
- **PROG-02**: Elev kan se styrkor och svagheter baserat på sessionshistorik
- **PROG-03**: Förälder kan se sitt barns progress och ämnesöversikt

## Existing Data Model

### Available Tables
- `sessions` — student_id, exercise_id, score (1-5), summary, started_at, ended_at
- `exercises` — topic_id, title, description, difficulty_order
- `topics` — subject_id, name, slug
- `subjects` — name, slug, display_order
- `review_schedule` — student_id, exercise_id, ease_factor, interval_days, next_review
- `students` — id, parent_id, name

### Existing Queries
- `ListSessionsByStudentID` — returns all sessions for a student, ordered by started_at DESC
- `ListDueReviews` — joins exercises/topics/subjects for enriched due-item rows
- `ListSubjects` — all subjects
- `ListTopicsBySubjectID` — topics for a subject
- `GetReviewSchedule` — per student+exercise

### Key Insight: No New Tables Needed
All progress data derives from existing `sessions` + `exercises` + `topics` + `subjects` joins. We need new **queries** (aggregations), not new schema.

## Backend Design

### New SQL Queries Needed

1. **GetStudentProgressBySubject** — Aggregate sessions per subject:
   ```sql
   SELECT s.id, s.name, s.slug,
     COUNT(DISTINCT se.id) AS total_sessions,
     COALESCE(AVG(se.score)::numeric(3,1), 0) AS avg_score,
     COUNT(DISTINCT se.exercise_id) AS unique_exercises
   FROM subjects s
   LEFT JOIN topics t ON t.subject_id = s.id
   LEFT JOIN exercises e ON e.topic_id = t.id
   LEFT JOIN sessions se ON se.exercise_id = e.id AND se.student_id = $1 AND se.ended_at IS NOT NULL
   GROUP BY s.id, s.name, s.slug
   ORDER BY s.display_order
   ```

2. **GetStudentProgressByTopic** — Aggregate sessions per topic within a subject:
   ```sql
   SELECT t.id, t.name, t.slug,
     COUNT(DISTINCT se.id) AS total_sessions,
     COALESCE(AVG(se.score)::numeric(3,1), 0) AS avg_score,
     COUNT(DISTINCT se.exercise_id) AS unique_exercises
   FROM topics t
   LEFT JOIN exercises e ON e.topic_id = t.id
   LEFT JOIN sessions se ON se.exercise_id = e.id AND se.student_id = $1 AND se.ended_at IS NOT NULL
   WHERE t.subject_id = $2
   GROUP BY t.id, t.name, t.slug
   ORDER BY t.display_order
   ```

3. **ListCompletedSessionsForStudent** — For parent view, enriched session list:
   ```sql
   SELECT se.id, se.score, se.started_at, se.ended_at,
     e.title AS exercise_title,
     t.name AS topic_name,
     s.name AS subject_name
   FROM sessions se
   JOIN exercises e ON e.id = se.exercise_id
   JOIN topics t ON t.id = e.topic_id
   JOIN subjects s ON s.id = t.subject_id
   WHERE se.student_id = $1 AND se.ended_at IS NOT NULL
   ORDER BY se.ended_at DESC
   ```

### New API Endpoints

1. **GET /api/progress** (child auth) — Student's own progress overview
   - Returns: per-subject stats + per-topic breakdowns with color codes

2. **GET /api/progress/:studentId** (parent auth) — Parent views child's progress
   - Authorization: parent must own this student (students.parent_id = parent.id)
   - Returns: same structure as student view

3. **GET /api/progress/:studentId/sessions** (parent auth) — Session history list
   - Returns: paginated list of completed sessions with exercise/topic/subject info

### New Backend Package: `progress`
Following existing patterns (srs, content, chat):
- `progress/handler.go` — HTTP handlers
- `progress/store.go` — ProgressStore interface + QueriesStore adapter

## Frontend Design

### Student Progress View (`/progress`)
- Auth: child role required
- Layout: Per-subject cards (Bio/Samhälle/Matte)
- Each card shows: antal genomförda pass, snittbetyg, bar chart of topic scores
- Color coding: green (4-5), yellow (3), red (1-2), grey (no data)
- Linked from study page or nav

### Parent Child View (`/dashboard/child/[id]`)
- Auth: parent role required
- Layout: Summary stats at top + full session history below
- Summary: antal pass senaste veckan, snittbetyg per ämne
- Session list: date, exercise title, topic, subject, score
- No chat history visible (per CONTEXT decision)

### Chart Library Decision (Claude's Discretion)
**Choice: No external chart library. Use CSS/HTML bar charts.**
Rationale:
- Only simple horizontal bar charts needed (score per topic)
- CSS bars are SSR-compatible, zero JS bundle cost
- No dependency to maintain
- shadcn-svelte already provides the design tokens (colors, spacing)
- Pattern: `<div style="width: {score/5 * 100}%">` with Tailwind background colors

### Color Coding
Muted, non-judgmental palette per CONTEXT:
- Green: `bg-emerald-200` (score 4-5) — mastered
- Yellow: `bg-amber-200` (score 3) — developing
- Red: `bg-rose-200` (score 1-2) — needs work
- Grey: `bg-gray-100` (no sessions) — not started

## Architecture Decisions

1. **No new migrations** — All data exists, just new queries
2. **Separate package** — `progress` package follows srs/content/chat pattern
3. **CSS-only charts** — No chart library dependency
4. **Parent authorization** — Verify parent_id ownership in handler, not middleware (one-off check)
5. **No pagination v1** — Session list returns all; paginate in v2 if needed

## Risk Assessment

- **Low risk**: Pure read-only layer, no data mutations
- **Data availability**: Sessions/scores from Phase 3 may be sparse in dev — ensure views handle empty state gracefully
- **Performance**: LEFT JOINs with aggregation on small dataset is fine for v1

## RESEARCH COMPLETE
