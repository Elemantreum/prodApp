CREATE TABLE bookings (
                          id SERIAL PRIMARY KEY,
                          user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                          training_id INTEGER NOT NULL REFERENCES trainings(id) ON DELETE CASCADE,
                          status VARCHAR(20) NOT NULL DEFAULT 'active'
                              CHECK (status IN ('active', 'cancelled', 'completed', 'not_completed')),
                          booking_date TIMESTAMPTZ NOT NULL DEFAULT now(),
                          unbooking_date TIMESTAMPTZ
);

CREATE INDEX idx_bookings_user_id ON bookings(user_id);
CREATE INDEX idx_bookings_training_id ON bookings(training_id);
CREATE INDEX idx_bookings_status ON bookings(status);

CREATE UNIQUE INDEX idx_bookings_unique_active
    ON bookings(user_id, training_id)
    WHERE status = 'active';