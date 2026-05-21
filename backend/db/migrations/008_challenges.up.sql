-- backend/db/migrations/008_challenges.up.sql

CREATE TABLE challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID NOT NULL REFERENCES parents(id) ON DELETE CASCADE,
    created_by_role TEXT NOT NULL CHECK (created_by_role IN ('parent', 'child')),
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    cover_emoji TEXT NOT NULL DEFAULT '📚',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE challenge_exercises (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    challenge_id UUID NOT NULL REFERENCES challenges(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    system_prompt TEXT NOT NULL,
    display_order INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE sessions
    ALTER COLUMN exercise_id DROP NOT NULL,
    ADD COLUMN challenge_exercise_id UUID REFERENCES challenge_exercises(id),
    ADD CONSTRAINT sessions_one_exercise_type CHECK (
        (exercise_id IS NOT NULL) != (challenge_exercise_id IS NOT NULL)
    );

CREATE INDEX idx_challenges_parent_id ON challenges(parent_id);
CREATE INDEX idx_challenge_exercises_challenge_id ON challenge_exercises(challenge_id);
