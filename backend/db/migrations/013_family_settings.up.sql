CREATE TABLE family_settings (
    id SMALLINT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    code_hash TEXT NOT NULL,
    device_epoch INT NOT NULL DEFAULT 1,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
