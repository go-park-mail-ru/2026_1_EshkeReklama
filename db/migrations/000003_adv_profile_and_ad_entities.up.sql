BEGIN;

CREATE TYPE eshkere.tariff_type AS ENUM (
    'noob', 'pro', 'cheater'
);

ALTER TABLE eshkere.advertiser
    ADD COLUMN IF NOT EXISTS surname TEXT,
    ADD COLUMN IF NOT EXISTS company TEXT,
    ADD COLUMN IF NOT EXISTS city    TEXT,
    ADD COLUMN IF NOT EXISTS tariff  eshkere.tariff_type NOT NULL DEFAULT 'noob';

ALTER TABLE eshkere.ad_campaign
    ADD COLUMN IF NOT EXISTS main_action eshkere.action_type NOT NULL DEFAULT 'look',
    ADD COLUMN IF NOT EXISTS total_spent BIGINT              NOT NULL DEFAULT 0;

ALTER TABLE eshkere.ad_group
    ADD COLUMN IF NOT EXISTS total_spent BIGINT NOT NULL DEFAULT 0;

ALTER TABLE eshkere.ad
    ADD COLUMN IF NOT EXISTS total_spent BIGINT NOT NULL DEFAULT 0;

COMMIT;