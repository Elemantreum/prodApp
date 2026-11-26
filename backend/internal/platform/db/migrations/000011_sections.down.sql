ALTER TABLE trainings DROP COLUMN venue;
DROP TRIGGER default_section ON users;
DROP FUNCTION join_original_section();
UPDATE users u SET is_admin = m.is_admin FROM section_members m WHERE m.user_id = u.id AND m.section_id = 1;
ALTER TABLE recurring_trainings DROP COLUMN section_id;
ALTER TABLE trainings DROP COLUMN section_id;
DROP TABLE section_members;
DROP TABLE sections;
