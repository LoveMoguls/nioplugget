-- name: CreateSession :one
INSERT INTO sessions (student_id, exercise_id) VALUES ($1, $2) RETURNING id, student_id, exercise_id, score, summary, started_at, ended_at;

-- name: GetSessionByID :one
SELECT id, student_id, exercise_id, score, summary, started_at, ended_at FROM sessions WHERE id = $1;

-- name: EndSession :one
UPDATE sessions SET ended_at = NOW(), score = $2, summary = $3 WHERE id = $1 AND ended_at IS NULL RETURNING id, student_id, exercise_id, score, summary, started_at, ended_at;

-- name: ListSessionsByStudentID :many
SELECT id, student_id, exercise_id, score, summary, started_at, ended_at FROM sessions WHERE student_id = $1 ORDER BY started_at DESC;
