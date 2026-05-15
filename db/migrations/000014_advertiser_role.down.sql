BEGIN;

ALTER TABLE eshkere.advertiser
    DROP COLUMN IF EXISTS role;

DROP TYPE IF EXISTS eshkere.advertiser_role;

COMMIT;
