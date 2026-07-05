-- name: GetFamilySettings :one
SELECT * FROM family_settings WHERE id = 1;

-- name: UpsertFamilyCode :one
INSERT INTO family_settings (id, code_hash)
VALUES (1, $1)
ON CONFLICT (id) DO UPDATE SET
    code_hash = $1,
    device_epoch = family_settings.device_epoch + 1,
    updated_at = NOW()
RETURNING *;

-- name: ListParents :many
SELECT id, email FROM parents ORDER BY created_at;

-- name: ListAllStudents :many
SELECT id, name FROM students ORDER BY created_at;
