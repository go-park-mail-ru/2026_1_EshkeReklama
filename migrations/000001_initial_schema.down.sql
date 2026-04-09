BEGIN;

-- Удаление триггеров
DROP TRIGGER IF EXISTS set_updated_at_advertiser ON eshkere.advertiser;
DROP TRIGGER IF EXISTS set_updated_at_ad_campaign ON eshkere.ad_campaign;
DROP TRIGGER IF EXISTS set_updated_at_ad_group ON eshkere.ad_group;
DROP TRIGGER IF EXISTS set_updated_at_ad ON eshkere.ad;
DROP TRIGGER IF EXISTS set_updated_at_partner ON eshkere.partner;
DROP TRIGGER IF EXISTS set_updated_at_partner_site ON eshkere.partner_site;

-- Удаление функций
DROP FUNCTION IF EXISTS update_updated_at_column;

-- Удаление таблиц
DROP TABLE IF EXISTS eshkere.ad_action;
DROP TABLE IF EXISTS eshkere.partner_site;
DROP TABLE IF EXISTS eshkere.partner;
DROP TABLE IF EXISTS eshkere.ad;
DROP TABLE IF EXISTS eshkere.ad_group;
DROP TABLE IF EXISTS eshkere.ad_campaign;
DROP TABLE IF EXISTS eshkere.advertiser;
DROP TABLE IF EXISTS eshkere.region;
DROP TABLE IF EXISTS eshkere.topic;

-- Удаление типов ENUM
DROP TYPE IF EXISTS eshkere.action_type;
DROP TYPE IF EXISTS eshkere.gender_type;
DROP TYPE IF EXISTS eshkere.status_type;

COMMIT;