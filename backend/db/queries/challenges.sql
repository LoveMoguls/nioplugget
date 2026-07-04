-- backend/db/queries/challenges.sql

-- name: CreateChallenge :one
INSERT INTO challenges (parent_id, created_by_role, title, description, cover_emoji, created_by_student_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, parent_id, created_by_role, title, description, cover_emoji, created_at, published, created_by_student_id;

-- name: CreateChallengeExercise :one
INSERT INTO challenge_exercises (challenge_id, title, description, system_prompt, display_order)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, challenge_id, title, description, system_prompt, display_order, created_at;

-- name: ListChallengesByParentID :many
SELECT c.id, c.parent_id, c.created_by_role, c.title, c.description, c.cover_emoji, c.created_at, c.published,
    COALESCE(st.name, '')::text AS creator_name
FROM challenges c
LEFT JOIN students st ON st.id = c.created_by_student_id
WHERE c.parent_id = $1
ORDER BY c.created_at DESC;

-- name: ListPublishedChallengesByParentID :many
SELECT c.id, c.parent_id, c.created_by_role, c.title, c.description, c.cover_emoji, c.created_at, c.published,
    COALESCE(st.name, '')::text AS creator_name
FROM challenges c
LEFT JOIN students st ON st.id = c.created_by_student_id
WHERE c.parent_id = $1 AND c.published = true
ORDER BY c.created_at DESC;

-- name: PublishChallenge :one
UPDATE challenges
SET title = $1, published = true
WHERE id = $2 AND parent_id = $3
RETURNING id, parent_id, created_by_role, title, description, cover_emoji, created_at, published, created_by_student_id;

-- name: GetChallengeByID :one
SELECT id, parent_id, created_by_role, title, description, cover_emoji, created_at, published, created_by_student_id
FROM challenges
WHERE id = $1;

-- name: GetChallengeExerciseByID :one
SELECT id, challenge_id, title, description, system_prompt, display_order, created_at
FROM challenge_exercises
WHERE id = $1;

-- name: ListChallengeExercisesByChallengeID :many
SELECT id, challenge_id, title, description, system_prompt, display_order, created_at
FROM challenge_exercises
WHERE challenge_id = $1
ORDER BY display_order;

-- name: ListChallengeExercisesWithProgress :many
SELECT DISTINCT ON (ce.display_order, ce.id)
    ce.id,
    ce.title,
    ce.description,
    ce.display_order,
    s.id AS session_id,
    s.score AS score
FROM challenge_exercises ce
LEFT JOIN sessions s ON s.challenge_exercise_id = ce.id
    AND s.student_id = $2
    AND s.ended_at IS NOT NULL
WHERE ce.challenge_id = $1
ORDER BY ce.display_order, ce.id, s.score DESC NULLS LAST;

-- name: DeleteChallenge :exec
DELETE FROM challenges WHERE id = $1 AND parent_id = $2;
