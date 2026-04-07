-- Rollback: remove Geografi and Historia subjects and all associated topics and exercises
-- ON DELETE CASCADE in schema handles topics and exercises automatically
DELETE FROM subjects WHERE slug IN ('geografi', 'historia');
