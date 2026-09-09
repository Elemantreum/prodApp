-- Правило регулярной тренировки: одна "якорная" дата/время первого занятия,
-- дальше повторяется каждые 7 дней от неё, пока is_active. Никакого RRULE —
-- просто фиксированный недельный шаг от anchor_date, без пересчёта дня
-- недели/таймзоны на сервере (см. RecurringService.GenerateOccurrences).
CREATE TABLE recurring_trainings (
    id SERIAL PRIMARY KEY,
    description VARCHAR(255) NOT NULL,
    anchor_date TIMESTAMPTZ NOT NULL,
    lead_time_ns BIGINT NOT NULL DEFAULT 2419200000000000,
    max_quantity INT NOT NULL,
    duration_ns BIGINT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Сгенерированные занятия — обычные строки в trainings, просто помеченные
-- ссылкой на правило. ON DELETE SET NULL: если правило когда-нибудь удалят
-- (сейчас это не предусмотрено, только деактивация), уже созданные
-- тренировки не должны исчезать вместе с ним.
ALTER TABLE trainings ADD COLUMN recurring_training_id INT REFERENCES recurring_trainings(id) ON DELETE SET NULL;

-- Идемпотентность генерации: повторный проход по одному и тому же правилу
-- на ту же дату не создаст дубликат (INSERT ... ON CONFLICT DO NOTHING).
CREATE UNIQUE INDEX idx_trainings_recurring_date ON trainings(recurring_training_id, date);
