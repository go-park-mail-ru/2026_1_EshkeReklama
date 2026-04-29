BEGIN;

ALTER TABLE eshkere.advertiser
    ADD COLUMN IF NOT EXISTS password_hash TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS password_salt TEXT NOT NULL DEFAULT '';

ALTER TABLE eshkere.advertiser
    ALTER COLUMN password_hash DROP DEFAULT,
    ALTER COLUMN password_salt DROP DEFAULT;

COMMIT;
