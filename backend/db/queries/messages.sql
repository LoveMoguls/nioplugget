-- name: CreateMessage :one
INSERT INTO messages (session_id, role, content) VALUES ($1, $2, $3) RETURNING id, session_id, role, content, created_at;

-- name: ListMessagesBySessionID :many
SELECT id, session_id, role, content, created_at FROM messages WHERE session_id = $1 ORDER BY created_at ASC;

-- name: CountMessagesBySessionID :one
SELECT COUNT(*) FROM messages WHERE session_id = $1;
