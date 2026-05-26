BEGIN;

ALTER TABLE eshkere.partner_site
    ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'draft';

ALTER TABLE eshkere.partner_site
    DROP CONSTRAINT IF EXISTS partner_site_status_check;

ALTER TABLE eshkere.partner_site
    ADD CONSTRAINT partner_site_status_check
        CHECK (status IN (
            'draft',
            'pending_review',
            'active',
            'rejected',
            'blocked',
            'archived'
        ));

COMMIT;
