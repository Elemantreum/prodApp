DROP INDEX IF EXISTS idx_trainings_recurring_date;
ALTER TABLE trainings DROP COLUMN recurring_training_id;
DROP TABLE recurring_trainings;
