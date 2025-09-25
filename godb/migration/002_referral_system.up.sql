-- 002_referral_system.up.sql
CREATE TABLE referral_links (
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    referral_uuid UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    is_active BOOLEAN DEFAULT TRUE,
);

CREATE TABLE referral_relations (
    id SERIAL PRIMARY KEY,
    referrer_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    referee_id INTEGER REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    referral_uuid UUID NOT NULL,
    bonus_credited BOOLEAN DEFAULT FALSE,
);

-- Индексы для скорости
CREATE INDEX idx_referral_links_uuid ON referral_links(referral_uuid);
CREATE INDEX idx_referral_links_user ON referral_links(user_id);
CREATE INDEX idx_referral_relations_referrer ON referral_relations(referrer_id);
CREATE INDEX idx_referral_relations_referee ON referral_relations(referee_id);
