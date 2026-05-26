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
