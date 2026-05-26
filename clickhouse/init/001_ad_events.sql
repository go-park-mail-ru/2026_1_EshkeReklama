CREATE DATABASE IF NOT EXISTS eshkere;

CREATE TABLE IF NOT EXISTS eshkere.ad_events
(
    event_id UUID,
    event_type LowCardinality(String),
    request_id String,
    occurred_at DateTime64(3, 'UTC'),
    event_date Date MATERIALIZED toDate(occurred_at),

    visitor_id String,
    advertiser_id UInt64,
    campaign_id UInt64,
    ad_group_id UInt64,
    ad_id UInt64,
    partner_block_id UInt64,
    partner_site_id UInt64,
    topic_id UInt64,

    price Int64,
    partner_reward Int64,
    platform_revenue Int64
)
ENGINE = MergeTree
PARTITION BY toYYYYMM(event_date)
ORDER BY (event_date, advertiser_id, campaign_id, ad_id, event_type, occurred_at);
