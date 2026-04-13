BEGIN;

DROP TRIGGER IF EXISTS set_updated_at_ad_feed_link ON eshkere.ad_feed_link;
DROP TABLE IF EXISTS eshkere.ad_feed_link;

ALTER TABLE eshkere.advertiser
    DROP COLUMN IF EXISTS avatar_url;

COMMIT;
