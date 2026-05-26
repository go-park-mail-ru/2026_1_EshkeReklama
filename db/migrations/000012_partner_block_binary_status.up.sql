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
