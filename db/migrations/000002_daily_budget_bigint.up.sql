BEGIN;

ALTER TABLE eshkere.ad_campaign
    ALTER COLUMN daily_budget TYPE BIGINT USING ROUND(daily_budget * 100)::BIGINT,
    ALTER COLUMN daily_budget SET DEFAULT 0;

COMMIT;
