-- 001_init_users.up.sql
CREATE TABLE users (
    telegram_id BIGINT UNIQUE NOT NULL,
);

CREATE TABLE user_balances (
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    requests_remaining INTEGER DEFAULT 5, -- 5 бесплатных запросов
);
