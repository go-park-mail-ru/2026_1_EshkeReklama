BEGIN;

DROP TRIGGER IF EXISTS set_updated_at_appeal ON eshkere.appeal;

DROP TABLE IF EXISTS eshkere.appeal;

DROP TYPE IF EXISTS eshkere.appeal_category;
DROP TYPE IF EXISTS eshkere.appeal_status;

COMMIT;