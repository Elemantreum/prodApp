CREATE TABLE training_waitlist (
    training_id INTEGER NOT NULL REFERENCES trainings(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (training_id, user_id)
);

CREATE INDEX training_waitlist_by_user ON training_waitlist(user_id);

CREATE FUNCTION notify_training_waitlist(target INTEGER) RETURNS void LANGUAGE sql AS $$
    INSERT INTO outbox(topic, payload)
    SELECT 'OnTrainingPlaceAvailable', jsonb_build_object(
        'TrainingId', t.id, 'Description', t.description, 'Date', t.date,
        'ChatIds', jsonb_agg(DISTINCT i.external_id::bigint)
    )
    FROM trainings t
    JOIN training_waitlist w ON w.training_id = t.id
    JOIN users u ON u.id = w.user_id AND u.is_approved
    JOIN user_identities i ON i.user_id = w.user_id AND i.provider = 'telegram'
    WHERE t.id = target AND t.is_active AND t.date > now()
      AND (SELECT count(*) FROM bookings b WHERE b.training_id = t.id AND b.status = 'active') < t.max_quantity
    GROUP BY t.id;
$$;

CREATE FUNCTION lock_booking_training() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    PERFORM id FROM trainings WHERE id = OLD.training_id FOR UPDATE;
    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER lock_waitlist_on_booking_change BEFORE UPDATE OR DELETE ON bookings
FOR EACH ROW EXECUTE FUNCTION lock_booking_training();

CREATE FUNCTION update_booking_waitlist() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    IF TG_OP <> 'DELETE' AND NEW.status = 'active' THEN
        DELETE FROM training_waitlist WHERE training_id = NEW.training_id AND user_id = NEW.user_id;
    END IF;
    IF TG_OP <> 'INSERT' AND OLD.status = 'active' THEN
        IF TG_OP = 'DELETE' OR NEW.status <> 'active' OR NEW.training_id <> OLD.training_id THEN
            PERFORM notify_training_waitlist(OLD.training_id);
        END IF;
    END IF;
    RETURN NULL;
END;
$$;

CREATE TRIGGER booking_waitlist_changed AFTER INSERT OR UPDATE OR DELETE ON bookings
FOR EACH ROW EXECUTE FUNCTION update_booking_waitlist();

CREATE FUNCTION training_waitlist_capacity_changed() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    PERFORM notify_training_waitlist(NEW.id);
    RETURN NULL;
END;
$$;

CREATE TRIGGER training_capacity_changed AFTER UPDATE ON trainings
FOR EACH ROW WHEN (NEW.max_quantity > OLD.max_quantity OR (NEW.is_active AND NOT OLD.is_active))
EXECUTE FUNCTION training_waitlist_capacity_changed();
