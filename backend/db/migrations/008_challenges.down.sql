-- backend/db/migrations/008_challenges.down.sql

ALTER TABLE sessions
    DROP CONSTRAINT sessions_one_exercise_type,
    DROP COLUMN challenge_exercise_id,
    ALTER COLUMN exercise_id SET NOT NULL;

DROP TABLE challenge_exercises;
DROP TABLE challenges;
