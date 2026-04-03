-- name: UpsertReviewSchedule :one
INSERT INTO review_schedule (student_id, exercise_id, ease_factor, interval_days, repetition_count, next_review, last_reviewed_at)
VALUES ($1, $2, $3, $4, $5, $6, NOW())
ON CONFLICT (student_id, exercise_id)
DO UPDATE SET
    ease_factor = $3,
    interval_days = $4,
    repetition_count = $5,
    next_review = $6,
    last_reviewed_at = NOW()
RETURNING *;

-- name: ListDueReviews :many
SELECT
    rs.id,
    rs.exercise_id,
    rs.next_review,
    rs.interval_days,
    rs.ease_factor,
    e.title AS exercise_title,
    t.name AS topic_name,
    t.slug AS topic_slug,
    s.name AS subject_name,
    s.slug AS subject_slug
FROM review_schedule rs
JOIN exercises e ON e.id = rs.exercise_id
JOIN topics t ON t.id = e.topic_id
JOIN subjects s ON s.id = t.subject_id
WHERE rs.student_id = $1
  AND rs.next_review <= NOW()
ORDER BY rs.next_review ASC;

-- name: GetReviewSchedule :one
SELECT * FROM review_schedule WHERE student_id = $1 AND exercise_id = $2;
