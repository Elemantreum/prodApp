CREATE TABLE section_venues (
    section_id integer NOT NULL REFERENCES sections ON DELETE CASCADE,
    name text NOT NULL CHECK (name <> '' AND btrim(name) = name)
);

CREATE UNIQUE INDEX section_venues_name ON section_venues (section_id, lower(name));

INSERT INTO section_venues (section_id, name)
SELECT section_id, btrim(venue)
FROM trainings
WHERE btrim(venue) <> ''
GROUP BY section_id, btrim(venue)
ORDER BY section_id, btrim(venue)
ON CONFLICT DO NOTHING;

ALTER TABLE recurring_trainings ADD COLUMN venue text NOT NULL DEFAULT '';
