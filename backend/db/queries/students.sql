-- name: CreateStudent :one
INSERT INTO students (parent_id, name, invite_code, invite_expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetStudentsByParentID :many
SELECT * FROM students
WHERE parent_id = $1
ORDER BY created_at ASC;

-- name: ActivateStudent :one
UPDATE students
SET pin_hash = $2,
    activated_at = NOW(),
    invite_code = NULL,
    invite_expires_at = NULL
WHERE invite_code = $1
  AND activated_at IS NULL
  AND invite_expires_at > NOW()
RETURNING *;

-- name: GetStudentByNameAndParent :one
SELECT * FROM students
WHERE name = $1
  AND parent_id = $2
LIMIT 1;

-- name: GetStudentByID :one
SELECT * FROM students
WHERE id = $1
LIMIT 1;

-- name: ListStudentNamesByParentID :many
SELECT id, name FROM students
WHERE parent_id = $1
ORDER BY name ASC;
