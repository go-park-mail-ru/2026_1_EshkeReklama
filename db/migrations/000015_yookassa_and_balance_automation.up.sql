ALTER TABLE eshkere.advertiser
    ADD COLUMN IF NOT EXISTS saved_payment_method_id TEXT,
    ADD COLUMN IF NOT EXISTS saved_payment_method_title TEXT;

CREATE TABLE IF NOT EXISTS eshkere.payment_transaction (
    id            TEXT PRIMARY KEY,
    advertiser_id INT NOT NULL REFERENCES eshkere.advertiser(id),
    amount        BIGINT NOT NULL,
    status        TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS eshkere.advertiser_notification_settings (
    advertiser_id      INT PRIMARY KEY REFERENCES eshkere.advertiser(id),
    email_enabled      BOOLEAN NOT NULL DEFAULT false,
    warning_threshold  BIGINT NOT NULL DEFAULT 500,
    critical_threshold BIGINT NOT NULL DEFAULT 100,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS eshkere.advertiser_autopay_settings (
    advertiser_id    INT PRIMARY KEY REFERENCES eshkere.advertiser(id),
    enabled          BOOLEAN NOT NULL DEFAULT false,
    threshold_amount BIGINT NOT NULL DEFAULT 5000,
    top_up_amount    BIGINT NOT NULL DEFAULT 30000,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
