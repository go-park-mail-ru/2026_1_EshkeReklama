ALTER TABLE eshkere.advertiser_notification_settings
    DROP COLUMN IF EXISTS telegram_chat_id,
    DROP COLUMN IF EXISTS telegram_enabled;
