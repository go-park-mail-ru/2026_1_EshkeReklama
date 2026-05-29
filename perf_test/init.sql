-- perf_test/init.sql
-- Снимок DDL до оптимизаций под нагрузочное тестирование.
-- Сгенерировано: 2026-05-29T17:48:05Z (UTC)
-- Git commit: 2d5fc01
-- Источник: db/migrations/*.up.sql (по порядку имени файла)
--
-- Пересборка: ./perf_test/generate_init_sql.sh

-- >>> 000001_initial_schema.up.sql
BEGIN;

CREATE SCHEMA IF NOT EXISTS eshkere;

CREATE TYPE eshkere.action_type AS ENUM (
    'look', 'click'
    );

CREATE TYPE eshkere.gender_type AS ENUM (
    'man', 'woman', 'any'
    );

CREATE TYPE eshkere.status_type AS ENUM (
    'turned_off', 'moderation', 'working', 'rejected', 'not_enough_money'
    );

CREATE TABLE eshkere.topic (
    id      INT     PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name    TEXT    NOT NULL UNIQUE
);

CREATE TABLE eshkere.region (
    id      INT     PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name    TEXT    NOT NULL UNIQUE
);

INSERT INTO eshkere.topic (name) VALUES ('Любой');
INSERT INTO eshkere.region (name) VALUES ('Любой');

CREATE TABLE eshkere.advertiser (
    id              INT                         PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name            TEXT                        NOT NULL,
    balance         BIGINT                      NOT NULL DEFAULT 0,
    created_at      TIMESTAMP WITH TIME ZONE    DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE,
    CONSTRAINT check_advertiser_name_length CHECK (LENGTH(name) <= 255),
    CONSTRAINT check_advertiser_balance_positive CHECK (balance >= 0)
);

CREATE TABLE eshkere.ad_campaign (
    id              INT                         PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    advertiser_id   INT                         REFERENCES eshkere.advertiser(id) ON DELETE CASCADE,
    status          eshkere.status_type         NOT NULL DEFAULT 'turned_off',
    name            TEXT                        NOT NULL,
    daily_budget    BIGINT                      NOT NULL DEFAULT 0,
    created_at      TIMESTAMP WITH TIME ZONE    DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE,
    CONSTRAINT check_ad_campaign_name_length CHECK (LENGTH(name) <= 255),
    CONSTRAINT check_ad_campaign_daily_budget_positive CHECK (daily_budget >= 0)
);

CREATE TABLE eshkere.ad_group (
    id              INT                         PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    ad_campaign_id  INT                         REFERENCES eshkere.ad_campaign(id) ON DELETE CASCADE,
    topic_id        INT                         DEFAULT 1 REFERENCES eshkere.topic(id) ON DELETE SET DEFAULT,
    region_id       INT                         DEFAULT 1 REFERENCES eshkere.region(id) ON DELETE SET DEFAULT,
    name            TEXT                        NOT NULL,
    age_from        INT                         NOT NULL,
    age_to          INT                         NOT NULL,
    gender          eshkere.gender_type         NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE    DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE,
    CONSTRAINT check_ad_group_name_length CHECK (LENGTH(name) <= 255),
    CONSTRAINT check_ad_group_age CHECK (age_from <= age_to AND age_from BETWEEN 0 AND 150 AND age_to BETWEEN 0 AND 150)
);

CREATE TABLE eshkere.ad (
    id          INT                         PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    ad_group_id INT                         REFERENCES eshkere.ad_group(id) ON DELETE CASCADE,
    status      eshkere.status_type         NOT NULL DEFAULT 'turned_off',
    title       TEXT                        NOT NULL,
    short_desc  TEXT                        NOT NULL,
    image_url   TEXT                        ,
    target_url  TEXT                        NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE    DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE,
    CONSTRAINT check_ad_title_length CHECK (LENGTH(title) <= 60),
    CONSTRAINT check_ad_short_desc_length CHECK (LENGTH(short_desc) <= 255)
);

CREATE TABLE eshkere.partner (
    id              INT                         PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name            TEXT                        NOT NULL,
    email           TEXT                        NOT NULL UNIQUE,
    phone_number    TEXT                        NOT NULL UNIQUE,
    password_hash   TEXT                        NOT NULL,
    password_salt   TEXT                        NOT NULL,
    balance         BIGINT                      NOT NULL DEFAULT 0,
    created_at      TIMESTAMP WITH TIME ZONE    DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE,
    CONSTRAINT check_partner_name_length CHECK (LENGTH(name) <= 255),
    CONSTRAINT check_partner_phone_number_length CHECK (LENGTH(phone_number) = 10),
    CONSTRAINT check_partner_balance_positive CHECK (balance >= 0)
);

CREATE TABLE eshkere.partner_site (
    id          INT                         PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    partner_id  INT                         REFERENCES eshkere.partner(id) ON DELETE CASCADE,
    topic_id    INT                         DEFAULT 1 REFERENCES eshkere.topic(id) ON DELETE SET DEFAULT,
    region_id   INT                         DEFAULT 1 REFERENCES eshkere.region(id) ON DELETE SET DEFAULT,
    age_from    INT                         NOT NULL,
    age_to      INT                         NOT NULL,
    gender      eshkere.gender_type         NOT NULL DEFAULT 'any',
    url         TEXT                        NOT NULL UNIQUE,
    created_at  TIMESTAMP WITH TIME ZONE    DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE,
    CONSTRAINT check_partner_site_age CHECK (age_from <= age_to AND age_from BETWEEN 0 AND 150 AND age_to BETWEEN 0 AND 150)
);

CREATE TABLE eshkere.ad_action (
    id              INT                         PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    ad_id           INT                         REFERENCES eshkere.ad(id) ON DELETE CASCADE,
    partner_site_id INT                         REFERENCES eshkere.partner_site(id) ON DELETE CASCADE,
    region_id       INT                         REFERENCES eshkere.region(id) ON DELETE CASCADE,
    action          eshkere.action_type         NOT NULL DEFAULT 'look',
    age             INT                         NOT NULL DEFAULT 25,
    gender          eshkere.gender_type         NOT NULL DEFAULT 'any',
    created_at      TIMESTAMP WITH TIME ZONE    DEFAULT NOW(),
    CONSTRAINT check_ad_action_age CHECK (age BETWEEN 0 AND 150)
);

-- Создание функции обновления атрибута обновления
CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

-- Настройка триггеров обновления атрибута обновления
CREATE TRIGGER set_updated_at_advertiser
    BEFORE UPDATE
    ON eshkere.advertiser
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_updated_at_ad_campaign
    BEFORE UPDATE
    ON eshkere.ad_campaign
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_updated_at_ad_group
    BEFORE UPDATE
    ON eshkere.ad_group
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_updated_at_ad
    BEFORE UPDATE
    ON eshkere.ad
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_updated_at_partner
    BEFORE UPDATE
    ON eshkere.partner
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_updated_at_partner_site
    BEFORE UPDATE
    ON eshkere.partner_site
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();


COMMIT;

-- >>> 000002_profile_and_feed.up.sql
BEGIN;

ALTER TABLE eshkere.advertiser
    ADD COLUMN IF NOT EXISTS avatar_url TEXT;

CREATE TABLE IF NOT EXISTS eshkere.ad_feed_link (
    ad_campaign_id INT PRIMARY KEY REFERENCES eshkere.ad_campaign(id) ON DELETE CASCADE,
    token         TEXT NOT NULL UNIQUE,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at    TIMESTAMP WITH TIME ZONE
);

CREATE TRIGGER set_updated_at_ad_feed_link
    BEFORE UPDATE
    ON eshkere.ad_feed_link
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

COMMIT;

-- >>> 000003_adv_profile_and_ad_entities.up.sql
BEGIN;

CREATE TYPE eshkere.tariff_type AS ENUM (
    'noob', 'pro', 'cheater'
);

ALTER TABLE eshkere.advertiser
    ADD COLUMN IF NOT EXISTS surname TEXT,
    ADD COLUMN IF NOT EXISTS company TEXT,
    ADD COLUMN IF NOT EXISTS city    TEXT,
    ADD COLUMN IF NOT EXISTS tariff  eshkere.tariff_type NOT NULL DEFAULT 'noob';

ALTER TABLE eshkere.ad_campaign
    ADD COLUMN IF NOT EXISTS main_action eshkere.action_type NOT NULL DEFAULT 'look',
    ADD COLUMN IF NOT EXISTS total_spent BIGINT              NOT NULL DEFAULT 0;

ALTER TABLE eshkere.ad_group
    ADD COLUMN IF NOT EXISTS total_spent BIGINT NOT NULL DEFAULT 0;

ALTER TABLE eshkere.ad
    ADD COLUMN IF NOT EXISTS total_spent BIGINT NOT NULL DEFAULT 0;

COMMIT;
-- >>> 000004_appeal.up.sql
BEGIN;

CREATE TYPE eshkere.appeal_status AS ENUM (
    'open', 'in_progress', 'closed'
);

CREATE TYPE eshkere.appeal_category AS ENUM (
    'bug', 'suggestion', 'complaint', 'question' -- баг, фича, продуктовая жалоба, вопрос по пользованию
);

CREATE TABLE eshkere.appeal (
    id              INT                         PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    advertiser_id   INT                         REFERENCES eshkere.advertiser(id) ON DELETE SET NULL,
    status          eshkere.appeal_status       NOT NULL DEFAULT 'open',
    category        eshkere.appeal_category     NOT NULL,
    title           TEXT                        NOT NULL,
    description     TEXT                        NOT NULL,
    image_url       TEXT,
    name            TEXT                        NOT NULL,
    email           TEXT                        NOT NULL,
    created_at      TIMESTAMP WITH TIME ZONE    DEFAULT NOW(),
    updated_at      TIMESTAMP WITH TIME ZONE,
    CONSTRAINT check_appeal_title_length CHECK (LENGTH(title) <= 100)
);


CREATE TRIGGER set_updated_at_appeal
    BEFORE UPDATE
    ON eshkere.appeal
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

COMMIT;
-- >>> 000005_auth_credentials.up.sql
CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE IF NOT EXISTS auth.credentials (
    id            BIGSERIAL PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    phone         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP
);

-- >>> 000006_drop_advertiser_password_columns.up.sql
BEGIN;

ALTER TABLE eshkere.advertiser
    DROP COLUMN IF EXISTS password_hash,
    DROP COLUMN IF EXISTS password_salt;

COMMIT;

-- >>> 000007_drop_advertiser_contacts_columns.up.sql
BEGIN;

ALTER TABLE eshkere.advertiser
    DROP COLUMN IF EXISTS email,
    DROP COLUMN IF EXISTS phone_number;

COMMIT;

-- >>> 000008_partner_onboarding_and_blocks.up.sql
BEGIN;

ALTER TABLE eshkere.partner
    DROP COLUMN IF EXISTS name,
    DROP COLUMN IF EXISTS balance,
    DROP COLUMN IF EXISTS email,
    DROP COLUMN IF EXISTS phone_number,
    DROP COLUMN IF EXISTS password_hash,
    DROP COLUMN IF EXISTS password_salt;

ALTER TABLE eshkere.partner
    ADD COLUMN IF NOT EXISTS last_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS first_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS middle_name TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS birth_date DATE NOT NULL DEFAULT DATE '1970-01-01',
    ADD COLUMN IF NOT EXISTS country_code TEXT NOT NULL DEFAULT 'RU',
    ADD COLUMN IF NOT EXISTS registration_region_code TEXT NOT NULL DEFAULT 'RU-MOW',
    ADD COLUMN IF NOT EXISTS cooperation_form TEXT NOT NULL DEFAULT 'self_employed',
    ADD COLUMN IF NOT EXISTS payout_currency TEXT NOT NULL DEFAULT 'RUB';

ALTER TABLE eshkere.partner_site
    DROP COLUMN IF EXISTS topic_id,
    DROP COLUMN IF EXISTS region_id,
    DROP COLUMN IF EXISTS age_from,
    DROP COLUMN IF EXISTS age_to,
    DROP COLUMN IF EXISTS gender;

ALTER TABLE eshkere.partner_site
    RENAME COLUMN url TO domain;

ALTER TABLE eshkere.partner_site
    ADD COLUMN IF NOT EXISTS site_name TEXT NOT NULL DEFAULT '';

ALTER TABLE eshkere.partner_site
    ALTER COLUMN domain TYPE TEXT;

ALTER TABLE eshkere.partner_site
    RENAME CONSTRAINT partner_site_url_key TO partner_site_domain_key;

CREATE TABLE IF NOT EXISTS eshkere.partner_block (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    partner_site_id INT NOT NULL REFERENCES eshkere.partner_site(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    block_type TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft',
    embed_token TEXT NOT NULL UNIQUE,
    cpm_strategy TEXT NOT NULL DEFAULT 'max_income',
    amp_mode TEXT NOT NULL DEFAULT 'disabled',
    size_mode TEXT NOT NULL DEFAULT 'adaptive',
    border_mode TEXT NOT NULL DEFAULT 'auto',
    corner_mode TEXT NOT NULL DEFAULT 'auto',
    theme TEXT NOT NULL DEFAULT 'light',
    interscroller_mode TEXT NOT NULL DEFAULT 'auto',
    interscroller_background_color TEXT,
    self_ad_settings JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS eshkere.partner_block_geo_rule (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    partner_block_id INT NOT NULL REFERENCES eshkere.partner_block(id) ON DELETE CASCADE,
    geo_code TEXT NOT NULL,
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    cpmv BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT partner_block_geo_rule_unique UNIQUE (partner_block_id, geo_code)
);

CREATE TRIGGER set_updated_at_partner_block
    BEFORE UPDATE
    ON eshkere.partner_block
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_updated_at_partner_block_geo_rule
    BEFORE UPDATE
    ON eshkere.partner_block_geo_rule
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

COMMIT;

-- >>> 000009_ad_billing.up.sql
BEGIN;

ALTER TABLE eshkere.ad_campaign
    ADD COLUMN IF NOT EXISTS cpm_price BIGINT NOT NULL DEFAULT 0;

ALTER TABLE eshkere.partner
    ADD COLUMN IF NOT EXISTS balance BIGINT NOT NULL DEFAULT 0;

ALTER TABLE eshkere.partner_block
    ADD COLUMN IF NOT EXISTS revenue_share_bps INT NOT NULL DEFAULT 7000;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'check_ad_campaign_cpm_price_positive'
    ) THEN
        ALTER TABLE eshkere.ad_campaign
            ADD CONSTRAINT check_ad_campaign_cpm_price_positive CHECK (cpm_price >= 0);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'check_partner_balance_positive'
    ) THEN
        ALTER TABLE eshkere.partner
            ADD CONSTRAINT check_partner_balance_positive CHECK (balance >= 0);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'check_partner_block_revenue_share_bps'
    ) THEN
        ALTER TABLE eshkere.partner_block
            ADD CONSTRAINT check_partner_block_revenue_share_bps CHECK (revenue_share_bps BETWEEN 0 AND 10000);
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS eshkere.ad_campaign_daily_spend (
    campaign_id       INT NOT NULL REFERENCES eshkere.ad_campaign(id) ON DELETE CASCADE,
    spend_date        DATE NOT NULL,
    advertiser_spend  BIGINT NOT NULL DEFAULT 0,
    partner_reward    BIGINT NOT NULL DEFAULT 0,
    platform_revenue  BIGINT NOT NULL DEFAULT 0,
    impressions       BIGINT NOT NULL DEFAULT 0,
    created_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at        TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (campaign_id, spend_date),
    CHECK (advertiser_spend >= 0),
    CHECK (partner_reward >= 0),
    CHECK (platform_revenue >= 0),
    CHECK (impressions >= 0)
);

CREATE TABLE IF NOT EXISTS eshkere.partner_block_daily_earning (
    partner_block_id  INT NOT NULL REFERENCES eshkere.partner_block(id) ON DELETE CASCADE,
    earning_date      DATE NOT NULL,
    reward            BIGINT NOT NULL DEFAULT 0,
    impressions       BIGINT NOT NULL DEFAULT 0,
    is_settled        BOOLEAN NOT NULL DEFAULT FALSE,
    settled_at        TIMESTAMP WITH TIME ZONE,
    created_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at        TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (partner_block_id, earning_date),
    CHECK (reward >= 0),
    CHECK (impressions >= 0)
);

CREATE TRIGGER set_updated_at_ad_campaign_daily_spend
    BEFORE UPDATE
    ON eshkere.ad_campaign_daily_spend
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_updated_at_partner_block_daily_earning
    BEFORE UPDATE
    ON eshkere.partner_block_daily_earning
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

COMMIT;

-- >>> 000010_partner_site_owner_to_advertiser.up.sql
BEGIN;

ALTER TABLE eshkere.partner_site
    DROP CONSTRAINT IF EXISTS partner_site_partner_id_fkey;

ALTER TABLE eshkere.partner_site
    ADD CONSTRAINT partner_site_partner_id_fkey
        FOREIGN KEY (partner_id) REFERENCES eshkere.advertiser(id) ON DELETE CASCADE;

COMMIT;

-- >>> 000011_partner_site_status.up.sql
BEGIN;

ALTER TABLE eshkere.partner_site
    ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'draft';

ALTER TABLE eshkere.partner_site
    DROP CONSTRAINT IF EXISTS partner_site_status_check;

ALTER TABLE eshkere.partner_site
    ADD CONSTRAINT partner_site_status_check
        CHECK (status IN (
            'draft',
            'pending_review',
            'active',
            'rejected',
            'blocked',
            'archived'
        ));

COMMIT;

-- >>> 000012_partner_block_binary_status.up.sql
BEGIN;

ALTER TABLE eshkere.partner_block
    ALTER COLUMN status SET DEFAULT 'inactive';

UPDATE eshkere.partner_block
SET status = CASE
    WHEN status = 'active' THEN 'active'
    ELSE 'inactive'
END;

ALTER TABLE eshkere.partner_block
    DROP CONSTRAINT IF EXISTS partner_block_status_check;

ALTER TABLE eshkere.partner_block
    ADD CONSTRAINT partner_block_status_check
        CHECK (status IN ('active', 'inactive'));

COMMIT;

-- >>> 000013_vkid_auth.up.sql
ALTER TABLE auth.credentials
    ALTER COLUMN email DROP NOT NULL,
    ALTER COLUMN phone DROP NOT NULL,
    ALTER COLUMN password_hash DROP NOT NULL;

ALTER TABLE auth.credentials
    ADD COLUMN IF NOT EXISTS vk_user_id BIGINT UNIQUE;

-- >>> 000014_advertiser_role.up.sql
BEGIN;

CREATE TYPE eshkere.advertiser_role AS ENUM ('user', 'admin');

ALTER TABLE eshkere.advertiser
    ADD COLUMN role eshkere.advertiser_role NOT NULL DEFAULT 'user';

COMMIT;

-- >>> 000015_yookassa_and_balance_automation.up.sql
ALTER TABLE eshkere.advertiser
    ADD COLUMN IF NOT EXISTS saved_payment_method_id TEXT,
    ADD COLUMN IF NOT EXISTS saved_payment_method_title TEXT;

CREATE TABLE IF NOT EXISTS eshkere.payment_transaction (
    id            TEXT PRIMARY KEY,
    advertiser_id INT NOT NULL REFERENCES eshkere.advertiser(id),
    amount        BIGINT NOT NULL,
    status        TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS eshkere.advertiser_notification_settings (
    advertiser_id      INT PRIMARY KEY REFERENCES eshkere.advertiser(id),
    email_enabled      BOOLEAN NOT NULL DEFAULT false,
    warning_threshold  BIGINT NOT NULL DEFAULT 500,
    critical_threshold BIGINT NOT NULL DEFAULT 100,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS eshkere.advertiser_autopay_settings (
    advertiser_id    INT PRIMARY KEY REFERENCES eshkere.advertiser(id),
    enabled          BOOLEAN NOT NULL DEFAULT false,
    threshold_amount BIGINT NOT NULL DEFAULT 5000,
    top_up_amount    BIGINT NOT NULL DEFAULT 30000,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- >>> 000016_topics.up.sql
BEGIN;

INSERT INTO eshkere.topic (name) VALUES
    ('Технологии'),
    ('Бизнес'),
    ('Красота и здоровье'),
    ('Авто'),
    ('Недвижимость'),
    ('Еда и рестораны'),
    ('Путешествия'),
    ('Спорт'),
    ('Мода'),
    ('Образование')
    ON CONFLICT (name) DO NOTHING;

COMMIT;
-- >>> 000017_regions.up.sql
BEGIN;

INSERT INTO eshkere.region (name) VALUES
    ('Москва'),
    ('Санкт-Петербург'),
    ('Казань'),
    ('Екатеринбург'),
    ('Новосибирск'),
    ('Краснодар'),
    ('Нижний Новгород'),
    ('Самара'),
    ('Ростов-на-Дону')
    ON CONFLICT (name) DO NOTHING;

COMMIT;
-- >>> 000018_subscription.up.sql
BEGIN;

-- Переименовываем значение enum 'noob' -> 'basic'
ALTER TYPE eshkere.tariff_type RENAME VALUE 'noob' TO 'basic';

-- Добавляем поле для хранения даты окончания подписки
-- NULL = бессрочно (basic и cheater), для pro = дата окончания
ALTER TABLE eshkere.advertiser
    ADD COLUMN IF NOT EXISTS tariff_expires_at TIMESTAMPTZ;

COMMIT;

-- >>> 000019_payment_type.up.sql
ALTER TABLE eshkere.payment_transaction
    ADD COLUMN IF NOT EXISTS payment_type TEXT NOT NULL DEFAULT 'balance';

-- >>> 000020_notification_telegram.up.sql
ALTER TABLE eshkere.advertiser_notification_settings
    ADD COLUMN IF NOT EXISTS telegram_enabled BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS telegram_chat_id TEXT;

