-- name: ListExercisesByTopicID :many
SELECT id, topic_id, title, description, difficulty_order FROM exercises WHERE topic_id = $1 ORDER BY difficulty_order;

-- name: GetExerciseByID :one
SELECT id, topic_id, title, description, difficulty_order, system_prompt FROM exercises WHERE id = $1;
