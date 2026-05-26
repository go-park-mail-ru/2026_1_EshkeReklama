BEGIN;

DROP TRIGGER IF EXISTS set_updated_at_partner_block_daily_earning ON eshkere.partner_block_daily_earning;
DROP TRIGGER IF EXISTS set_updated_at_ad_campaign_daily_spend ON eshkere.ad_campaign_daily_spend;

DROP TABLE IF EXISTS eshkere.partner_block_daily_earning;
DROP TABLE IF EXISTS eshkere.ad_campaign_daily_spend;

ALTER TABLE eshkere.partner_block
    DROP CONSTRAINT IF EXISTS check_partner_block_revenue_share_bps,
    DROP COLUMN IF EXISTS revenue_share_bps;

ALTER TABLE eshkere.partner
    DROP CONSTRAINT IF EXISTS check_partner_balance_positive,
    DROP COLUMN IF EXISTS balance;

ALTER TABLE eshkere.ad_campaign
    DROP CONSTRAINT IF EXISTS check_ad_campaign_cpm_price_positive,
    DROP COLUMN IF EXISTS cpm_price;

COMMIT;
