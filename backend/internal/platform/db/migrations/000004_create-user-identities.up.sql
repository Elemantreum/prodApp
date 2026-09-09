CREATE TABLE user_identities (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    provider VARCHAR(20) NOT NULL,
    external_id VARCHAR(255) NOT NULL,
    -- Telegram-username (без @), нужен чтобы строить t.me/<username> ссылку в
    -- списках "кто записан"/посещаемость. Для VK не нужен — там ссылка
    -- vk.com/id<external_id> строится и без него, поле просто остаётся пустым.
    username VARCHAR(255) NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE(provider, external_id)
);

CREATE INDEX idx_user_identities_user_id ON user_identities(user_id);
