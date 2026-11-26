DROP TRIGGER default_section ON users;

CREATE TABLE legacy_default_section_members (
    user_id integer PRIMARY KEY REFERENCES users ON DELETE CASCADE
);

INSERT INTO legacy_default_section_members (user_id)
SELECT m.user_id
FROM section_members m
JOIN users u ON u.id = m.user_id
WHERE m.section_id = 1 AND NOT m.is_admin AND NOT u.is_admin AND NOT u.is_approved
  AND NOT EXISTS (
      SELECT 1
      FROM bookings b
      JOIN trainings t ON t.id = b.training_id
      WHERE b.user_id = m.user_id AND t.section_id = m.section_id
        AND b.status = 'active' AND t.date > now()
  );

DELETE FROM section_members m
USING legacy_default_section_members legacy
WHERE m.section_id = 1 AND m.user_id = legacy.user_id;

ALTER TABLE users ADD COLUMN profile_locked boolean NOT NULL DEFAULT false;

UPDATE users u
SET profile_locked = true
WHERE u.is_approved OR EXISTS (SELECT 1 FROM section_members m WHERE m.user_id = u.id);

CREATE FUNCTION lock_section_member_profile() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    UPDATE users SET profile_locked = true WHERE id = NEW.user_id AND NOT profile_locked;
    RETURN NEW;
END;
$$;

CREATE TRIGGER section_member_profile AFTER INSERT ON section_members
FOR EACH ROW EXECUTE FUNCTION lock_section_member_profile();

CREATE TABLE section_invites (
    id serial PRIMARY KEY,
    section_id integer NOT NULL REFERENCES sections ON DELETE CASCADE,
    token_hash text NOT NULL UNIQUE,
    created_by integer REFERENCES users ON DELETE SET NULL,
    expires_at timestamptz NOT NULL,
    max_uses integer NOT NULL CHECK (max_uses > 0),
    uses integer NOT NULL DEFAULT 0 CHECK (uses BETWEEN 0 AND max_uses),
    revoked_at timestamptz
);

CREATE INDEX section_invites_section_id_idx ON section_invites (section_id, id DESC);

CREATE OR REPLACE FUNCTION notify_training_waitlist(target integer) RETURNS void LANGUAGE sql AS $$
    INSERT INTO outbox(topic, payload)
    SELECT 'OnTrainingPlaceAvailable', jsonb_build_object(
        'TrainingId', t.id, 'Description', t.description, 'Date', t.date,
        'ChatIds', jsonb_agg(DISTINCT i.external_id::bigint)
    )
    FROM trainings t
    JOIN training_waitlist w ON w.training_id = t.id
    JOIN users u ON u.id = w.user_id AND u.is_registered
    LEFT JOIN section_members m ON m.user_id = w.user_id AND m.section_id = t.section_id
    JOIN user_identities i ON i.user_id = w.user_id AND i.provider = 'telegram'
    WHERE t.id = target AND t.is_active AND t.date > now()
      AND (u.is_admin OR m.user_id IS NOT NULL)
      AND (SELECT count(*) FROM bookings b WHERE b.training_id = t.id AND b.status = 'active') < t.max_quantity
    GROUP BY t.id;
$$;
