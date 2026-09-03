-- За сколько до занятия его нужно материализовать в trainings (см.
-- RecurringService.generateForRule). DEFAULT — 28 дней в наносекундах,
-- совпадает со старым фиксированным горизонтом, который это поле заменяет.
ALTER TABLE recurring_trainings ADD COLUMN lead_time_ns BIGINT NOT NULL DEFAULT 2419200000000000;
