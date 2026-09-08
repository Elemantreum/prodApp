    -- Telegram-username (без @), нужен чтобы строить t.me/<username> ссылку в
    -- списках "кто записан"/посещаемость. Для VK не нужен — там ссылка
    -- vk.com/id<external_id> строится и без него, поле просто остаётся пустым.
    ALTER TABLE user_identities ADD COLUMN username VARCHAR(255) NOT NULL DEFAULT '';
