ALTER TABLE challenges ADD COLUMN created_by_student_id UUID REFERENCES students(id) ON DELETE SET NULL;
