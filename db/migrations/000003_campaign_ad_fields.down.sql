BEGIN;

ALTER TABLE eshkere.ad_campaign
    DROP COLUMN IF EXISTS title,
    DROP COLUMN IF EXISTS short_desc,
    DROP COLUMN IF EXISTS image_url,
    DROP COLUMN IF EXISTS target_url;

COMMIT;
