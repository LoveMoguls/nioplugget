---
phase: 02-content-ai-chat
plan: 01
status: complete
started: 2026-04-03
completed: 2026-04-03
duration: ~10min
---

# Plan 02-01 Summary: DB Schema + Exercise Seed Data

## What Was Built
Database schema for content and sessions (5 new tables), seed data for 36 exercises across 3 subjects and 12 topics, and sqlc-generated Go queries.

## Tasks Completed

| # | Task | Status | Files |
|---|------|--------|-------|
| 1 | Schema migration (002) | Done | backend/db/migrations/002_content_and_sessions.{up,down}.sql |
| 2 | Seed migration (003) | Done | backend/db/migrations/003_seed_exercises.{up,down}.sql |
| 3 | sqlc queries + generation | Done | backend/db/queries/{subjects,topics,exercises,sessions,messages}.sql + generated .go files |

## Key Files

### Created
- `backend/db/migrations/002_content_and_sessions.up.sql` - 5 tables: subjects, topics, exercises, sessions, messages
- `backend/db/migrations/003_seed_exercises.up.sql` - 36 exercises with detailed Socratic system prompts in Swedish
- `backend/db/queries/subjects.sql` - ListSubjects, GetSubjectBySlug
- `backend/db/queries/topics.sql` - ListTopicsBySubjectID, GetTopicBySlug (named params)
- `backend/db/queries/exercises.sql` - ListExercisesByTopicID, GetExerciseByID
- `backend/db/queries/sessions.sql` - CreateSession, GetSessionByID, EndSession, ListSessionsByStudentID
- `backend/db/queries/messages.sql` - CreateMessage, ListMessagesBySessionID, CountMessagesBySessionID

## Decisions
- Used `sqlc.arg()` named parameters for GetTopicBySlug to avoid `Slug`/`Slug_2` field names
- 36 exercises total (3 per topic area, 4 topics per subject, 3 subjects)
- Each system prompt includes: rules, misconceptions, example questions, opening greeting

## Self-Check: PASSED
