BEGIN;

CREATE TYPE eshkere.advertiser_role AS ENUM ('user', 'admin');

ALTER TABLE eshkere.advertiser
    ADD COLUMN role eshkere.advertiser_role NOT NULL DEFAULT 'user';

COMMIT;
