-- backend/db/queries/challenges.sql

-- name: CreateChallenge :one
INSERT INTO challenges (parent_id, created_by_role, title, description, cover_emoji)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, parent_id, created_by_role, title, description, cover_emoji, created_at;

-- name: CreateChallengeExercise :one
INSERT INTO challenge_exercises (challenge_id, title, description, system_prompt, display_order)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, challenge_id, title, description, system_prompt, display_order, created_at;

-- name: ListChallengesByParentID :many
SELECT id, parent_id, created_by_role, title, description, cover_emoji, created_at
FROM challenges
WHERE parent_id = $1
ORDER BY created_at DESC;

-- name: GetChallengeByID :one
SELECT id, parent_id, created_by_role, title, description, cover_emoji, created_at
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
SELECT
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
ORDER BY ce.display_order, s.score DESC NULLS LAST;

-- name: DeleteChallenge :exec
DELETE FROM challenges WHERE id = $1 AND parent_id = $2;
