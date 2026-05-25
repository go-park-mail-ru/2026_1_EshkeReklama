DROP TABLE IF EXISTS eshkere.advertiser_autopay_settings;
DROP TABLE IF EXISTS eshkere.advertiser_notification_settings;
DROP TABLE IF EXISTS eshkere.payment_transaction;

ALTER TABLE eshkere.advertiser
    DROP COLUMN IF EXISTS saved_payment_method_title,
    DROP COLUMN IF EXISTS saved_payment_method_id;
