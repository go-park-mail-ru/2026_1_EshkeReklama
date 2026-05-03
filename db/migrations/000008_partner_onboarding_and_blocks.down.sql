BEGIN;

DROP TRIGGER IF EXISTS set_updated_at_partner_block_geo_rule ON eshkere.partner_block_geo_rule;
DROP TRIGGER IF EXISTS set_updated_at_partner_block ON eshkere.partner_block;

DROP TABLE IF EXISTS eshkere.partner_block_geo_rule;
DROP TABLE IF EXISTS eshkere.partner_block;

ALTER TABLE eshkere.partner_site
    DROP COLUMN IF EXISTS site_name;

ALTER TABLE eshkere.partner_site
    RENAME COLUMN domain TO url;

ALTER TABLE eshkere.partner_site
    ADD COLUMN IF NOT EXISTS topic_id INT DEFAULT 1 REFERENCES eshkere.topic(id) ON DELETE SET DEFAULT,
    ADD COLUMN IF NOT EXISTS region_id INT DEFAULT 1 REFERENCES eshkere.region(id) ON DELETE SET DEFAULT,
    ADD COLUMN IF NOT EXISTS age_from INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS age_to INT NOT NULL DEFAULT 100,
    ADD COLUMN IF NOT EXISTS gender eshkere.gender_type NOT NULL DEFAULT 'any';

ALTER TABLE eshkere.partner
    DROP COLUMN IF EXISTS last_name,
    DROP COLUMN IF EXISTS first_name,
    DROP COLUMN IF EXISTS middle_name,
    DROP COLUMN IF EXISTS birth_date,
    DROP COLUMN IF EXISTS country_code,
    DROP COLUMN IF EXISTS registration_region_code,
    DROP COLUMN IF EXISTS cooperation_form,
    DROP COLUMN IF EXISTS payout_currency;

ALTER TABLE eshkere.partner
    ADD COLUMN IF NOT EXISTS name TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS email TEXT UNIQUE,
    ADD COLUMN IF NOT EXISTS phone_number TEXT UNIQUE,
    ADD COLUMN IF NOT EXISTS password_hash TEXT,
    ADD COLUMN IF NOT EXISTS password_salt TEXT,
    ADD COLUMN IF NOT EXISTS balance BIGINT NOT NULL DEFAULT 0;

COMMIT;
