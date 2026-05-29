-- Удаление данных нагрузочного теста (title LIKE 'LOADTEST_%').
-- psql ... -f perf_test/cleanup.sql

BEGIN;

DELETE FROM eshkere.ad_action
WHERE ad_id IN (SELECT id FROM eshkere.ad WHERE title LIKE 'LOADTEST_%');

DELETE FROM eshkere.ad WHERE title LIKE 'LOADTEST_%';

DELETE FROM eshkere.ad_group
WHERE name LIKE 'LOADTEST_%'
  AND NOT EXISTS (SELECT 1 FROM eshkere.ad a WHERE a.ad_group_id = ad_group.id);

DELETE FROM eshkere.ad_campaign
WHERE name LIKE 'LOADTEST_%'
  AND NOT EXISTS (SELECT 1 FROM eshkere.ad_group g WHERE g.ad_campaign_id = ad_campaign.id);

COMMIT;
