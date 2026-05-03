BEGIN;

ALTER TABLE eshkere.partner_site
    DROP CONSTRAINT IF EXISTS partner_site_partner_id_fkey;

ALTER TABLE eshkere.partner_site
    ADD CONSTRAINT partner_site_partner_id_fkey
        FOREIGN KEY (partner_id) REFERENCES eshkere.advertiser(id) ON DELETE CASCADE;

COMMIT;
