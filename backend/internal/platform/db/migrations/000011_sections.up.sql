CREATE TABLE sections (
    id serial PRIMARY KEY,
    name text NOT NULL CHECK (length(trim(name)) BETWEEN 1 AND 120),
    logo_url text NOT NULL DEFAULT ''
);
INSERT INTO sections(name) VALUES ('Продвинутый уровень');
CREATE TABLE section_members (
    section_id integer NOT NULL REFERENCES sections ON DELETE CASCADE,
    user_id integer NOT NULL REFERENCES users ON DELETE CASCADE,
    is_admin boolean NOT NULL DEFAULT false,
    PRIMARY KEY(section_id, user_id)
);
INSERT INTO section_members SELECT 1, id, is_admin FROM users;
ALTER TABLE trainings ADD COLUMN section_id integer NOT NULL DEFAULT 1 REFERENCES sections;
ALTER TABLE recurring_trainings ADD COLUMN section_id integer NOT NULL DEFAULT 1 REFERENCES sections;
CREATE INDEX trainings_section_date ON trainings(section_id,date);
UPDATE users SET is_admin = (id = 1) WHERE is_admin IS DISTINCT FROM (id = 1);
CREATE FUNCTION join_original_section() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    INSERT INTO section_members(section_id,user_id)
    SELECT id, NEW.id FROM sections WHERE id = 1;
    RETURN NEW;
END;
$$;
CREATE TRIGGER default_section AFTER INSERT ON users FOR EACH ROW EXECUTE FUNCTION join_original_section();
ALTER TABLE trainings ADD COLUMN venue text NOT NULL DEFAULT '';
