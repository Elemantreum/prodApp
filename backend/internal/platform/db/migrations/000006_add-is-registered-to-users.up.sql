-- Существующие аккаунты считаем уже "зарегистрированными" (DEFAULT TRUE),
-- чтобы не выкидывать текущих тестеров на обязательный экран регистрации
-- задним числом. Новые аккаунты создаются с is_registered = FALSE явно —
-- см. UserService.EnsureProfile.
ALTER TABLE users ADD COLUMN is_registered BOOLEAN NOT NULL DEFAULT TRUE;
