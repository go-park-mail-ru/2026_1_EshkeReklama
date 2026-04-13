BEGIN;

ALTER TABLE eshkere.advertiser
    ADD COLUMN IF NOT EXISTS avatar_url TEXT;

CREATE TABLE IF NOT EXISTS eshkere.ad_feed_link (
    advertiser_id INT PRIMARY KEY REFERENCES eshkere.advertiser(id) ON DELETE CASCADE,
    token         TEXT NOT NULL UNIQUE,
    created_at    TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at    TIMESTAMP WITH TIME ZONE
);

CREATE TRIGGER set_updated_at_ad_feed_link
    BEFORE UPDATE
    ON eshkere.ad_feed_link
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

COMMIT;
