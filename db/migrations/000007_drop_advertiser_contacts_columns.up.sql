BEGIN;

ALTER TABLE eshkere.advertiser
    DROP COLUMN IF EXISTS email,
    DROP COLUMN IF EXISTS phone_number;

COMMIT;
