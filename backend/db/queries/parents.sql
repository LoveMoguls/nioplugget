-- name: CreateParent :one
INSERT INTO parents (email, password_hash, gdpr_consent_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetParentByEmail :one
SELECT * FROM parents
WHERE email = $1
LIMIT 1;

-- name: GetParentByID :one
SELECT * FROM parents
WHERE id = $1
LIMIT 1;
