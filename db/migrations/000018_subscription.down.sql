BEGIN;

ALTER TABLE eshkere.advertiser
    DROP COLUMN IF EXISTS tariff_expires_at;

ALTER TYPE eshkere.tariff_type RENAME VALUE 'basic' TO 'noob';

COMMIT;
