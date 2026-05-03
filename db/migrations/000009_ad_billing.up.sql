BEGIN;

ALTER TABLE eshkere.ad_campaign
    ADD COLUMN IF NOT EXISTS cpm_price BIGINT NOT NULL DEFAULT 0;

ALTER TABLE eshkere.partner
    ADD COLUMN IF NOT EXISTS balance BIGINT NOT NULL DEFAULT 0;

ALTER TABLE eshkere.partner_block
    ADD COLUMN IF NOT EXISTS revenue_share_bps INT NOT NULL DEFAULT 7000;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'check_ad_campaign_cpm_price_positive'
    ) THEN
        ALTER TABLE eshkere.ad_campaign
            ADD CONSTRAINT check_ad_campaign_cpm_price_positive CHECK (cpm_price >= 0);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'check_partner_balance_positive'
    ) THEN
        ALTER TABLE eshkere.partner
            ADD CONSTRAINT check_partner_balance_positive CHECK (balance >= 0);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'check_partner_block_revenue_share_bps'
    ) THEN
        ALTER TABLE eshkere.partner_block
            ADD CONSTRAINT check_partner_block_revenue_share_bps CHECK (revenue_share_bps BETWEEN 0 AND 10000);
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS eshkere.ad_campaign_daily_spend (
    campaign_id       INT NOT NULL REFERENCES eshkere.ad_campaign(id) ON DELETE CASCADE,
    spend_date        DATE NOT NULL,
    advertiser_spend  BIGINT NOT NULL DEFAULT 0,
    partner_reward    BIGINT NOT NULL DEFAULT 0,
    platform_revenue  BIGINT NOT NULL DEFAULT 0,
    impressions       BIGINT NOT NULL DEFAULT 0,
    created_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at        TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (campaign_id, spend_date),
    CHECK (advertiser_spend >= 0),
    CHECK (partner_reward >= 0),
    CHECK (platform_revenue >= 0),
    CHECK (impressions >= 0)
);

CREATE TABLE IF NOT EXISTS eshkere.partner_block_daily_earning (
    partner_block_id  INT NOT NULL REFERENCES eshkere.partner_block(id) ON DELETE CASCADE,
    earning_date      DATE NOT NULL,
    reward            BIGINT NOT NULL DEFAULT 0,
    impressions       BIGINT NOT NULL DEFAULT 0,
    is_settled        BOOLEAN NOT NULL DEFAULT FALSE,
    settled_at        TIMESTAMP WITH TIME ZONE,
    created_at        TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at        TIMESTAMP WITH TIME ZONE,
    PRIMARY KEY (partner_block_id, earning_date),
    CHECK (reward >= 0),
    CHECK (impressions >= 0)
);

CREATE TRIGGER set_updated_at_ad_campaign_daily_spend
    BEFORE UPDATE
    ON eshkere.ad_campaign_daily_spend
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER set_updated_at_partner_block_daily_earning
    BEFORE UPDATE
    ON eshkere.partner_block_daily_earning
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

COMMIT;
