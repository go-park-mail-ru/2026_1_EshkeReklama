BEGIN;

ALTER TABLE eshkere.partner_block
    DROP CONSTRAINT IF EXISTS partner_block_status_check;

UPDATE eshkere.partner_block
SET status = CASE
    WHEN status = 'active' THEN 'active'
    ELSE 'draft'
END;

ALTER TABLE eshkere.partner_block
    ALTER COLUMN status SET DEFAULT 'draft';

COMMIT;
