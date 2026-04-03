-- name: UpsertAPIKey :one
INSERT INTO api_keys (parent_id, encrypted_key)
VALUES ($1, $2)
ON CONFLICT (parent_id)
DO UPDATE SET encrypted_key = EXCLUDED.encrypted_key,
              updated_at = NOW()
RETURNING *;

-- name: GetAPIKeyByParentID :one
SELECT * FROM api_keys
WHERE parent_id = $1
LIMIT 1;

-- name: DeleteAPIKeyByParentID :exec
DELETE FROM api_keys
WHERE parent_id = $1;
