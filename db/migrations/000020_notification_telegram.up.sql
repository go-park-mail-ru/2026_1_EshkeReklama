ALTER TABLE eshkere.advertiser_notification_settings
    ADD COLUMN IF NOT EXISTS telegram_enabled BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS telegram_chat_id TEXT;
