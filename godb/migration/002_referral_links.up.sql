DROP TABLE IF EXISTS referral_links CASCADE;

CREATE TABLE IF NOT EXISTS referral_links (
    user_id BIGINT PRIMARY KEY,
    referral_uuid UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    usage_count INTEGER DEFAULT 0
);

-- Индексы
CREATE INDEX IF NOT EXISTS idx_referral_links_uuid ON referral_links(referral_uuid);
CREATE INDEX IF NOT EXISTS idx_referral_links_user_id ON referral_links(user_id);
