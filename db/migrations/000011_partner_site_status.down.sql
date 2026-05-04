BEGIN;

ALTER TABLE eshkere.partner_site
    DROP CONSTRAINT IF EXISTS partner_site_status_check;

ALTER TABLE eshkere.partner_site
    DROP COLUMN IF EXISTS status;

COMMIT;
