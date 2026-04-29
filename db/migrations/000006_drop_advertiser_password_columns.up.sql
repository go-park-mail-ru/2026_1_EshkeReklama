BEGIN;

ALTER TABLE eshkere.advertiser
    DROP COLUMN IF EXISTS password_hash,
    DROP COLUMN IF EXISTS password_salt;

COMMIT;
