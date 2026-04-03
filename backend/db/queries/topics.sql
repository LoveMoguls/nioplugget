-- name: ListTopicsBySubjectID :many
SELECT id, subject_id, name, slug, display_order FROM topics WHERE subject_id = $1 ORDER BY display_order;

-- name: GetTopicBySlug :one
SELECT t.id, t.subject_id, t.name, t.slug, t.display_order FROM topics t JOIN subjects s ON t.subject_id = s.id WHERE s.slug = $1 AND t.slug = $2;
