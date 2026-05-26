BEGIN;

ALTER TABLE eshkere.advertiser
    ADD COLUMN IF NOT EXISTS email TEXT,
    ADD COLUMN IF NOT EXISTS phone_number TEXT;

UPDATE eshkere.advertiser
SET
    email = COALESCE(email, CONCAT('rollback-', id, '@local.invalid')),
    phone_number = COALESCE(phone_number, LPAD(id::text, 10, '0'));

ALTER TABLE eshkere.advertiser
    ALTER COLUMN email SET NOT NULL,
    ALTER COLUMN phone_number SET NOT NULL;

ALTER TABLE eshkere.advertiser
    ADD CONSTRAINT advertiser_email_key UNIQUE (email),
    ADD CONSTRAINT advertiser_phone_number_key UNIQUE (phone_number),
    ADD CONSTRAINT check_advertiser_phone_number_length CHECK (LENGTH(phone_number) = 10);

COMMIT;
