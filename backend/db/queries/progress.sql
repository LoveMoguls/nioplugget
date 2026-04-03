-- name: GetStudentProgressBySubject :many
SELECT s.id, s.name, s.slug,
  COUNT(DISTINCT se.id)::int AS total_sessions,
  COALESCE(ROUND(AVG(se.score)::numeric, 1), 0)::float8 AS avg_score,
  COUNT(DISTINCT se.exercise_id)::int AS unique_exercises
FROM subjects s
LEFT JOIN topics t ON t.subject_id = s.id
LEFT JOIN exercises e ON e.topic_id = t.id
LEFT JOIN sessions se ON se.exercise_id = e.id AND se.student_id = $1 AND se.ended_at IS NOT NULL AND se.score IS NOT NULL
GROUP BY s.id, s.name, s.slug, s.display_order
ORDER BY s.display_order;

-- name: GetStudentProgressByTopic :many
SELECT t.id, t.name, t.slug,
  COUNT(DISTINCT se.id)::int AS total_sessions,
  COALESCE(ROUND(AVG(se.score)::numeric, 1), 0)::float8 AS avg_score,
  COUNT(DISTINCT se.exercise_id)::int AS unique_exercises
FROM topics t
LEFT JOIN exercises e ON e.topic_id = t.id
LEFT JOIN sessions se ON se.exercise_id = e.id AND se.student_id = $1 AND se.ended_at IS NOT NULL AND se.score IS NOT NULL
WHERE t.subject_id = $2
GROUP BY t.id, t.name, t.slug, t.display_order
ORDER BY t.display_order;

-- name: ListCompletedSessions :many
SELECT se.id, se.score, se.started_at, se.ended_at,
  e.title AS exercise_title,
  t.name AS topic_name,
  s.name AS subject_name
FROM sessions se
JOIN exercises e ON e.id = se.exercise_id
JOIN topics t ON t.id = e.topic_id
JOIN subjects s ON s.id = t.subject_id
WHERE se.student_id = $1 AND se.ended_at IS NOT NULL AND se.score IS NOT NULL
ORDER BY se.ended_at DESC;
