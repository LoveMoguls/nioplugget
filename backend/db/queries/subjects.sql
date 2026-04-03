-- name: ListSubjects :many
SELECT id, name, slug, display_order FROM subjects ORDER BY display_order;

-- name: GetSubjectBySlug :one
SELECT id, name, slug, display_order FROM subjects WHERE slug = $1;
