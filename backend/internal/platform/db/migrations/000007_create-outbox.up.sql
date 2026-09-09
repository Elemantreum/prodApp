-- Transactional Outbox: сервисы пишут сюда вместо прямой публикации в
-- Kafka, отдельный релей (см. main.go) вычитывает неопубликованное и шлёт
-- в Kafka. topic = имя топика/события (совпадает с именами структур событий
-- и топиков в tg-notifier), payload — уже сериализованное в JSON тело события.
--
-- ВАЖНО: сейчас запись сюда НЕ в одной транзакции с записью в доменную
-- таблицу (training.AddTraining/UpdateTraining, user.EnsureProfile просто
-- делают отдельный INSERT следом) — это осознанно оставлено как TODO,
-- добавляется позже отдельно.
CREATE TABLE outbox (
    id BIGSERIAL PRIMARY KEY,
    topic VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at TIMESTAMPTZ
);

-- Частичный индекс — держит быстрым "найди неопубликованное" даже когда
-- опубликованные строки годами копятся в таблице.
CREATE INDEX idx_outbox_unpublished ON outbox (created_at) WHERE published_at IS NULL;
