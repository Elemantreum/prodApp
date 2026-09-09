CREATE TABLE trainings (
    id SERIAL PRIMARY KEY,
    description TEXT NOT NULL,
    date TIMESTAMPTZ NOT NULL,
    max_quantity INTEGER NOT NULL CHECK (max_quantity > 0),
    duration_ns BIGINT NOT NULL CHECK (duration_ns > 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_trainings_date ON trainings(date);
CREATE INDEX idx_trainings_is_active ON trainings(is_active);
