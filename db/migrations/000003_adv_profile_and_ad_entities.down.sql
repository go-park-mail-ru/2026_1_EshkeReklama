BEGIN;

ALTER TABLE eshkere.ad
    DROM COLUMN IF EXISTS total_spent;

ALTER TABLE eshkere.ad_group
    DROP COLUMN IF EXISTS total_spent;

ALTER TABLE eshkere.ad_campaign
    DROP COLUMN IF EXISTS total_spent,
    DROP COLUMN IF EXISTS main_action;

ALTER TABLE eshkere.advertiser
    DROP COLUMN IF NOT EXISTS tariff,
    DROP COLUMN IF NOT EXISTS city,
    DROP COLUMN IF NOT EXISTS company,
    DROP COLUMN IF NOT EXISTS surname;

DROP TYPE IF EXISTS eshekere.tariff_type;

COMMIT;
