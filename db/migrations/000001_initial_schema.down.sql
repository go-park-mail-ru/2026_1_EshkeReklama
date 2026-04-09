BEGIN;

-- Удаление триггеров
DROP TRIGGER IF EXISTS set_updated_at_advertiser ON eshekere.advertiser;
DROP TRIGGER IF EXISTS set_updated_at_ad_campaign ON eshekere.ad_campaign;
DROP TRIGGER IF EXISTS set_updated_at_ad_group ON eshekere.ad_group;
DROP TRIGGER IF EXISTS set_updated_at_ad ON eshekere.ad;
DROP TRIGGER IF EXISTS set_updated_at_partner ON eshekere.partner;
DROP TRIGGER IF EXISTS set_updated_at_partner_site ON eshekere.partner_site;

-- Удаление функций
DROP FUNCTION IF EXISTS update_updated_at_column;

-- Удаление таблиц
DROP TABLE IF EXISTS eshekere.ad_action;
DROP TABLE IF EXISTS eshekere.partner_site;
DROP TABLE IF EXISTS eshekere.partner;
DROP TABLE IF EXISTS eshekere.ad;
DROP TABLE IF EXISTS eshekere.ad_group;
DROP TABLE IF EXISTS eshekere.ad_campaign;
DROP TABLE IF EXISTS eshekere.advertiser;
DROP TABLE IF EXISTS eshekere.region;
DROP TABLE IF EXISTS eshekere.topic;

-- Удаление типов ENUM
DROP TYPE IF EXISTS eshekere.action_type;
DROP TYPE IF EXISTS eshekere.gender_type;
DROP TYPE IF EXISTS eshekere.status_type;

COMMIT;