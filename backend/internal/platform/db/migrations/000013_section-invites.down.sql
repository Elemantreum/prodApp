CREATE OR REPLACE FUNCTION notify_training_waitlist(target INTEGER) RETURNS void LANGUAGE sql AS $$
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

DROP TABLE section_invites;
DROP TRIGGER section_member_profile ON section_members;
DROP FUNCTION lock_section_member_profile();
ALTER TABLE users DROP COLUMN profile_locked;

INSERT INTO section_members (section_id, user_id)
SELECT 1, user_id FROM legacy_default_section_members
ON CONFLICT DO NOTHING;

DROP TABLE legacy_default_section_members;

CREATE TRIGGER default_section AFTER INSERT ON users
FOR EACH ROW EXECUTE FUNCTION join_original_section();
