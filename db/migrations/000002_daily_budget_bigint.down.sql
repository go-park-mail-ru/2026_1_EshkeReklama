BEGIN;

ALTER TABLE eshkere.ad_campaign
    ALTER COLUMN daily_budget TYPE DECIMAL(15, 2) USING (daily_budget::DECIMAL / 100),
    ALTER COLUMN daily_budget SET DEFAULT 0;

COMMIT;
