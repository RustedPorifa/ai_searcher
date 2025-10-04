DROP TABLE IF EXISTS users CASCADE;

CREATE TABLE IF NOT EXISTS users (
    user_id BIGINT PRIMARY KEY,
    role VARCHAR(50) DEFAULT 'member',
    requests_remaining INTEGER DEFAULT 0,
    referred_by UUID UNIQUE DEFAULT NULL,
    is_refferal_rewarded BOOLEAN DEFAULT false
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_referred_by ON users(referred_by);
